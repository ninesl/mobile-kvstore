package kvstore

import (
	"context"
	"database/sql"
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
// returns existing blob IDs for refs in this scope.
func (items ItemArgs) hasBlob(s storer) ([]int, error) {
	var errs []error
	if items.Scope.Namespace == "" {
		errs = append(errs, fmt.Errorf("namespace is required"))
	}
	if len(items.Refs) == 0 {
		errs = append(errs, fmt.Errorf("no item refs to check for blob"))
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	refs := make([]sqliteRefPair, 0, len(items.Refs))
	for i, ref := range items.Refs {
		if ref.MetaTag == "" {
			errs = append(errs, fmt.Errorf("refs[%d]: meta_tag is required", i))
			continue
		}
		refs = append(refs, ref.toSqlitePair())
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	blobIDs := make([]int, 0)
	seen := make(map[int64]struct{})
	// Process 100-ref chunks first, then the modulo/% remainder through the 10-ref query.
	// Fixed chunk sizes let sqlc keep prepared queries while nil pair slots handle short batches.
	for len(refs) >= 100 {
		params, err := refsTo100Params(items.Scope, refs[:100])
		if err != nil {
			return nil, err
		}
		rows, err := s.queries.GetBlobRefsByScopeAnd100Refs(context.Background(), params)
		if err != nil {
			return nil, err
		}
		blobIDs = appendDistinctBlobIDs(blobIDs, seen, rows)
		refs = refs[100:]
	}

	for len(refs) > 0 {
		n := min(10, len(refs))
		params, err := refsTo10Params(items.Scope, refs[:n])
		if err != nil {
			return nil, err
		}
		rows, err := s.queries.GetBlobRefsByScopeAnd10Refs(context.Background(), params)
		if err != nil {
			return nil, err
		}
		blobIDs = appendDistinctBlobIDs(blobIDs, seen, rows)
		refs = refs[n:]
	}

	return blobIDs, nil
}

func (s storer) Get(namespace string, subject int, decode DecoderFunc) ([]any, error) {
	rows, err := s.queries.GetBlobValuesByScope(context.Background(),
		sqlstore.GetBlobValuesByScopeParams{
			Namespace: namespace,
			Subject:   int64(subject),
		})
	if err != nil {
		return nil, err
	}

	return decodeResults(blobRowsFromScope(rows), decode)
}

func (s storer) GetByRef(f Filters, decode DecoderFunc) ([]any, error) {
	rows, err := s.queries.GetBlobValuesByRef(context.Background(),
		sqlstore.GetBlobValuesByRefParams{
			Namespace: f.Scope.Namespace,
			Subject:   int64(f.Scope.Subject),
			ID:        int64(f.Refs.ID),
			MetaTag:   f.Refs.MetaTag,
		})
	if err != nil {
		return nil, err
	}

	return decodeResults(blobRowsFromRef(rows), decode)
}

type blobRow struct {
	BlobID int64
	Data   []byte
}

func blobRowsFromScope(rows []sqlstore.GetBlobValuesByScopeRow) []blobRow {
	results := make([]blobRow, 0, len(rows))
	for _, row := range rows {
		results = append(results, blobRow{BlobID: row.BlobID, Data: row.BlobValue})
	}
	return results
}

func blobRowsFromRef(rows []sqlstore.GetBlobValuesByRefRow) []blobRow {
	results := make([]blobRow, 0, len(rows))
	for _, row := range rows {
		results = append(results, blobRow{BlobID: row.BlobID, Data: row.BlobValue})
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
func (s storer) storeBlob(data any, encode EncoderFunc) (int, error) {
	blob, err := encode(data)
	if err != nil {
		return -1, err
	}
	blobID, err := s.queries.UpsertBlob(context.Background(),
		sqlstore.UpsertBlobParams{
			BlobKey:   blob.SerializedKey,
			BlobValue: blob.Data,
			UpdatedAt: time.Now().UnixNano(),
		})
	return int(blobID), err
}

func (s storer) storeRef(scope BatchScope, ref ItemRef, blobID int) error {
	return s.queries.InsertBlobRef(context.Background(),
		sqlstore.InsertBlobRefParams{
			Namespace: scope.Namespace,
			Subject:   int64(scope.Subject),
			ID:        int64(ref.ID),
			MetaTag:   ref.MetaTag,
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

	// assign/find blob_id
	// TODO: TTL logic
	if blobIDs, err = items.hasBlob(s); err != nil {
		return err
	} else if len(blobIDs) == 0 {
		var blobID int
		blobID, err = s.storeBlob(items.Data, items.Encode)
		if err != nil {
			return err
		}
		blobIDs = append(blobIDs, blobID)
	}

	var errs []error
	for _, blobID := range blobIDs {
		for _, ref := range items.Refs {
			if err := s.storeRef(items.Scope, ref, blobID); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// Expects a sqlite file, can be a blank .db or one that was created by this package
func New(conn *sql.DB) Store {
	return storer{
		queries: sqlstore.New(conn),
	}
}
