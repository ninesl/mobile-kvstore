package kvstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ninesl/kvstore/sqlstore"
)

type storer struct {
	queries *sqlstore.Queries
}

// TODO: this function will also have a routine or be in conjunction
// with validating TTL or whatever else it deemed needed
//
// returns existing blob IDs for identities in this scope.
func (items ItemArgs) hasBlob(s storer) ([]int, error) {
	var errs []error
	if items.Scope.Namespace == "" {
		errs = append(errs, fmt.Errorf("namespace is required"))
	}
	if len(items.Identities) == 0 {
		errs = append(errs, fmt.Errorf("no item identities to check for blob"))
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	entries := make([]sqliteEntryPair, 0, len(items.Identities))
	for i, identity := range items.Identities {
		if identity.MetaTag == "" {
			errs = append(errs, fmt.Errorf("identities[%d]: meta_tag is required", i))
			continue
		}
		entries = append(entries, identity.toSqliteEntry())
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	blobIDs := make([]int, 0)
	seen := make(map[int64]struct{})

	// Process 100-entry chunks first, then the modulo/% remainder through the 10-entry query.
	// Fixed chunk sizes let sqlc keep prepared queries while nil pair slots handle short batches.
	for len(entries) >= 100 {
		params, err := entriesTo100Params(items.Scope, entries[:100])
		if err != nil {
			return nil, err
		}
		rows, err := s.queries.GetBlobEntriesByScopeAnd100Entries(context.Background(), params)
		if err != nil {
			return nil, err
		}
		blobIDs = appendDistinctBlobIDs(blobIDs, seen, rows)
		entries = entries[100:]
	}

	for len(entries) > 0 {
		n := min(10, len(entries))
		params, err := entriesTo10Params(items.Scope, entries[:n])
		if err != nil {
			return nil, err
		}
		rows, err := s.queries.GetBlobEntriesByScopeAnd10Entries(context.Background(), params)
		if err != nil {
			return nil, err
		}
		blobIDs = appendDistinctBlobIDs(blobIDs, seen, rows)
		entries = entries[n:]
	}
	return blobIDs, nil
}

func (s storer) Get(scope Scope, decode DecoderFunc) ([]any, error) {
	rows, err := s.queries.GetBlobValuesByScope(
		context.Background(),
		sqlstore.GetBlobValuesByScopeParams{
			Namespace: scope.Namespace,
			Subject:   int64(scope.Subject),
		})
	if err != nil {
		return nil, err
	}
	return decodeResults(blobRowsFromScope(rows), decode)
}

func (s storer) GetByFilter(f Filter, decode DecoderFunc) ([]any, error) {
	rows, err := s.queries.GetBlobValuesByIdentity(
		context.Background(),
		sqlstore.GetBlobValuesByIdentityParams{
			Namespace: f.Scope.Namespace,
			Subject:   int64(f.Scope.Subject),
			ID:        int64(f.Identity.ID),
			MetaTag:   f.Identity.MetaTag,
		})
	if err != nil {
		return nil, err
	}
	return decodeResults(blobRowsFromIdentity(rows), decode)
}

func (s storer) Count(filter CountFilter) int { return 0 }

func (s storer) ClearByScope(scope Scope) error { return nil }

func (s storer) ClearBySubject(subject string) error { return nil }

func (s storer) ClearByIdentity(identity Identity) error { return nil }

func (s storer) ClearByScopeIdentity(scope Scope, identity Identity) error { return nil }

func (s storer) ClearByMetaTag(tag string) error { return nil }

func (s storer) ClearByBlobKey(key string) error { return nil }

type blobRow struct {
	BlobID  int64
	BlobKey string
	Data    []byte
}

func blobRowsFromScope(rows []sqlstore.GetBlobValuesByScopeRow) []blobRow {
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

func blobRowsFromIdentity(rows []sqlstore.GetBlobValuesByIdentityRow) []blobRow {
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

func (s storer) storeIdentity(scope Scope, identity Identity, blobID int) error {
	return s.queries.InsertBlobEntry(context.Background(),
		sqlstore.InsertBlobEntryParams{
			Namespace: scope.Namespace,
			Subject:   int64(scope.Subject),
			ID:        int64(identity.ID),
			MetaTag:   identity.MetaTag,
			BlobID:    int64(blobID),
		})
}

// Store is smart because of checking if the item's
// shared blob exists or not and will not waste time Encode() ing
// your underlying type you are storying must implement StoreItem
// each storeItem has ItemArgs that tell the Storer how to store and define how it fits in the .db file
func (s storer) Store(storeItem StoreItem) error {
	var (
		items   = storeItem.ItemArgs()
		blobIDs []int
		err     error
	)

	// check if storer has the serialized key

	// assign/find blob_id
	if blobIDs, err = items.hasBlob(s); err != nil {
		return err
	} else if len(blobIDs) == 0 {
		var blobID int
		blobID, err = s.storeBlob(items.Data, items.Encode, items.BlobKey)
		if err != nil {
			return err
		}
		blobIDs = append(blobIDs, blobID)
	} else {
		// TODO: TTL logic to verify blob is new?
		// should be based on scope and identity?
		for _, blobID := range blobIDs {
			if err := s.overwriteBlob(
				blobID,
				items.Data,
				items.Encode,
				items.BlobKey,
			); err != nil {
				return err
			}
		}
	}

	var errs []error
	for _, blobID := range blobIDs {
		for _, identity := range items.Identities {
			if err := s.storeIdentity(
				items.Scope,
				identity,
				blobID,
			); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}
