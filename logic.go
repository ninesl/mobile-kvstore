package kvstore

import (
	"context"
	"errors"
	"time"

	"github.com/ninesl/kvstore/sqlstore"
)

type storer struct {
	queries *sqlstore.Queries
}

func (s storer) GetByFilter(filter EntryFilter, decode DecoderFunc) ([]any, error) {
	rows, err := s.queries.GetBlobValuesByFilter(context.Background(), getBlobValuesParams(filter))
	if err != nil {
		return nil, err
	}
	return decodeResults(blobRowsFromFilter(rows), decode)
}

// returns (-1, err) if error while getting count
func (s storer) Count(filter EntryFilter) (int, error) {
	count, err := s.queries.CountBlobEntriesByFilter(context.Background(), countBlobEntriesParams(filter))
	if err != nil {
		return -1, err
	}
	return int(count), nil
}

func (s storer) ClearByScope(scope Scope) error {
	return s.deleteBlobsByFilter(EntryFilter{
		Namespace: &scope.Namespace,
		Subject:   &scope.Subject,
	})
}

func (s storer) ClearByIdentity(identity Identity) error {
	return s.deleteBlobsByFilter(EntryFilter{
		ID:      &identity.ID,
		MetaTag: &identity.MetaTag,
	})
}

func (s storer) ClearByFilter(filter EntryFilter) error {
	if countFilterEmpty(filter) {
		return errors.New("ClearByFilter requires at least one filter field")
	}
	return s.deleteBlobsByFilter(filter)
}

func (s storer) ClearEverything() error {
	return s.deleteBlobsByFilter(EntryFilter{})
}

func (s storer) ClearByBlobKey(key string) error {
	return s.queries.DeleteBlobByBlobKey(context.Background(), key)
}

// if filter is empty, deletes all blobs
func (s storer) deleteBlobsByFilter(filter EntryFilter) error {
	return s.queries.DeleteBlobsByFilter(context.Background(), deleteBlobsParams(filter))
}

func countBlobEntriesParams(filter EntryFilter) sqlstore.CountBlobEntriesByFilterParams {
	params := sqlstore.CountBlobEntriesByFilterParams{}
	if filter.Namespace != nil {
		params.Namespace = *filter.Namespace
	}
	if filter.Subject != nil {
		params.Subject = int64(*filter.Subject)
	}
	if filter.ID != nil {
		params.ID = int64(*filter.ID)
	}
	if filter.MetaTag != nil {
		params.MetaTag = *filter.MetaTag
	}
	return params
}

func getBlobValuesParams(filter EntryFilter) sqlstore.GetBlobValuesByFilterParams {
	params := sqlstore.GetBlobValuesByFilterParams{}
	if filter.Namespace != nil {
		params.Namespace = *filter.Namespace
	}
	if filter.Subject != nil {
		params.Subject = int64(*filter.Subject)
	}
	if filter.ID != nil {
		params.ID = int64(*filter.ID)
	}
	if filter.MetaTag != nil {
		params.MetaTag = *filter.MetaTag
	}
	return params
}

func deleteBlobsParams(filter EntryFilter) sqlstore.DeleteBlobsByFilterParams {
	params := sqlstore.DeleteBlobsByFilterParams{}
	if filter.Namespace != nil {
		params.Namespace = *filter.Namespace
	}
	if filter.Subject != nil {
		params.Subject = int64(*filter.Subject)
	}
	if filter.ID != nil {
		params.ID = int64(*filter.ID)
	}
	if filter.MetaTag != nil {
		params.MetaTag = *filter.MetaTag
	}
	return params
}

func countFilterEmpty(filter EntryFilter) bool {
	return filter.Namespace == nil &&
		filter.Subject == nil &&
		filter.ID == nil &&
		filter.MetaTag == nil
}

type blobRow struct {
	BlobID  int64
	BlobKey string
	Data    []byte
}

func blobRowsFromFilter(rows []sqlstore.GetBlobValuesByFilterRow) []blobRow {
	results := make([]blobRow, 0, len(rows))
	for _, row := range rows {
		results = append(results, blobRow{
			BlobID:  row.BlobID,
			BlobKey: row.BlobKey,
			Data:    row.BlobValue,
		})
	}
	return results
}

func decodeResults(results []blobRow, decode DecoderFunc) ([]any, error) {
	outputs := make([]any, 0, len(results))
	decoded := make(map[int64]any, len(results))
	errs := make([]error, 0)
	for _, row := range results {
		if _, ok := decoded[row.BlobID]; ok {
			continue
		}
		o, err := decode(row.Data)
		if err != nil {
			errs = append(errs, err)
		} else {
			decoded[row.BlobID] = o
			outputs = append(outputs, o)
		}
	}
	if len(errs) > 0 {
		return outputs, errors.Join(errs...)
	}
	return outputs, nil
}

// returns the autoincremented blob_id
func (s storer) storeBlob(data any, encode EncoderFunc, blobKey string) (int, error) {
	blob, err := encode(data)
	if err != nil {
		return -1, err
	}
	blobID, err := s.queries.UpsertBlob(context.Background(),
		sqlstore.UpsertBlobParams{
			BlobKey:   blobKey,
			BlobValue: blob,
			UpdatedAt: time.Now().UnixNano(),
		})
	return int(blobID), err
}

// data MUST be compat with encoderFunc.
// EncoderFunc should user-defined error
// if the data is wrong somehow
func (s storer) overwriteBlob(blobID int, data any, encode EncoderFunc, blobKey string) error {
	blob, err := encode(data)
	if err != nil {
		return err
	}
	return s.queries.UpdateBlobValue(context.Background(),
		sqlstore.UpdateBlobValueParams{
			BlobKey:   blobKey,
			BlobValue: blob,
			UpdatedAt: time.Now().UnixNano(),
			BlobID:    int64(blobID),
		})
}

func (s storer) storeIdentities(scope Scope, identities []Identity, blobID int) error {
	entries := make([]sqliteEntryPair, 0, len(identities))
	for _, identity := range identities {
		entries = append(entries, identity.toSqliteEntry())
	}

	for len(entries) >= 100 {
		params, err := entriesToInsert100Params(scope, entries[:100], blobID)
		if err != nil {
			return err
		}
		if err := s.queries.InsertBlobEntriesByScopeAnd100Entries(context.Background(), params); err != nil {
			return err
		}
		entries = entries[100:]
	}

	for len(entries) > 0 {
		n := min(10, len(entries))
		params, err := entriesToInsert10Params(scope, entries[:n], blobID)
		if err != nil {
			return err
		}
		if err := s.queries.InsertBlobEntriesByScopeAnd10Entries(context.Background(), params); err != nil {
			return err
		}
		entries = entries[n:]
	}
	return nil
}

// Store is smart because of checking if the item's
// shared blob exists or not and will not waste time Encode() ing
// FIXME: currently overwrites the blob if it exists, no TTL logic
//
// your underlying type you are storing must implement StoreItem
// each storeItem has ItemArgs that tell the Storer how the blob fits in the .db file
func (s storer) Store(storeItem StoreItem) error {
	var (
		items  = storeItem.ItemArgs()
		blobID int
		err    error
	)

	// check if storer has the serialized key/blob key
	if items.BlobKey == "" {
		return errors.New("StoreItem must have a blob key")
	}
	if len(items.Identities) == 0 {
		return errors.New("StoreItem must have identities")
	}

	blobID64, err := s.queries.GetBlobIDByBlobKey(context.Background(), items.BlobKey)
	if err != nil {
		return err
	}
	if blobID64 != nil {
		blobID = int(*blobID64)
		// TODO: TTL logic to verify blob is new?
		// should be based on scope and identity?
		if err := s.overwriteBlob(
			blobID,
			items.Data,
			items.Encode,
			items.BlobKey,
		); err != nil {
			return err
		}
	} else {
		blobID, err = s.storeBlob(
			items.Data, items.Encode, items.BlobKey)
		if err != nil {
			return err
		}
	}

	return s.storeIdentities(items.Scope, items.Identities, blobID)
}
