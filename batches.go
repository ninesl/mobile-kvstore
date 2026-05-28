package kvstore

import (
	"fmt"

	"github.com/ninesl/kvstore/sqlstore"
)

type sqliteEntryPair struct {
	ID      *int64
	MetaTag *string
}

func (identity Identity) toSqliteEntry() sqliteEntryPair {
	id := int64(identity.ID)

	return sqliteEntryPair{
		ID:      &id,
		MetaTag: &identity.MetaTag,
	}
}

func appendDistinctBlobIDs(blobIDs []int, seen map[int64]struct{}, rows []int64) []int {
	for _, blobID := range rows {
		// Dedupe across chunks so fixed-size prepared queries can replace one dynamic query.
		if _, ok := seen[blobID]; ok {
			continue
		}
		seen[blobID] = struct{}{}
		blobIDs = append(blobIDs, int(blobID))
	}
	return blobIDs
}

func entriesTo10Params(scope Scope, entries []sqliteEntryPair) (sqlstore.GetBlobEntriesByScopeAnd10EntriesParams, error) {
	if len(entries) == 0 {
		return sqlstore.GetBlobEntriesByScopeAnd10EntriesParams{}, fmt.Errorf("entry batch 10 requires at least one entry")
	}
	if len(entries) > 10 {
		return sqlstore.GetBlobEntriesByScopeAnd10EntriesParams{}, fmt.Errorf("entry batch 10 got %d entries", len(entries))
	}

	params := sqlstore.GetBlobEntriesByScopeAnd10EntriesParams{
		Namespace: scope.Namespace,
		Subject:   int64(scope.Subject),
	}
	for i, entry := range entries {
		if err := setEntryBatch10Slot(&params, i, entry); err != nil {
			return sqlstore.GetBlobEntriesByScopeAnd10EntriesParams{}, err
		}
	}
	return params, nil
}

func entriesToInsert10Params(scope Scope, entries []sqliteEntryPair, blobID int) (sqlstore.InsertBlobEntriesByScopeAnd10EntriesParams, error) {
	if len(entries) == 0 {
		return sqlstore.InsertBlobEntriesByScopeAnd10EntriesParams{}, fmt.Errorf("entry insert batch 10 requires at least one entry")
	}
	if len(entries) > 10 {
		return sqlstore.InsertBlobEntriesByScopeAnd10EntriesParams{}, fmt.Errorf("entry insert batch 10 got %d entries", len(entries))
	}

	params := sqlstore.InsertBlobEntriesByScopeAnd10EntriesParams{
		Namespace: scope.Namespace,
		Subject:   int64(scope.Subject),
		BlobID:    int64(blobID),
	}
	for i, entry := range entries {
		if err := setInsertEntryBatch10Slot(&params, i, entry); err != nil {
			return sqlstore.InsertBlobEntriesByScopeAnd10EntriesParams{}, err
		}
	}
	return params, nil
}

func entriesTo100Params(scope Scope, entries []sqliteEntryPair) (sqlstore.GetBlobEntriesByScopeAnd100EntriesParams, error) {
	if len(entries) == 0 {
		return sqlstore.GetBlobEntriesByScopeAnd100EntriesParams{}, fmt.Errorf("entry batch 100 requires at least one entry")
	}
	if len(entries) > 100 {
		return sqlstore.GetBlobEntriesByScopeAnd100EntriesParams{}, fmt.Errorf("entry batch 100 got %d entries", len(entries))
	}

	params := sqlstore.GetBlobEntriesByScopeAnd100EntriesParams{
		Namespace: scope.Namespace,
		Subject:   int64(scope.Subject),
	}
	for i, entry := range entries {
		if err := setEntryBatch100Slot(&params, i, entry); err != nil {
			return sqlstore.GetBlobEntriesByScopeAnd100EntriesParams{}, err
		}
	}
	return params, nil
}

func entriesToInsert100Params(scope Scope, entries []sqliteEntryPair, blobID int) (sqlstore.InsertBlobEntriesByScopeAnd100EntriesParams, error) {
	if len(entries) == 0 {
		return sqlstore.InsertBlobEntriesByScopeAnd100EntriesParams{}, fmt.Errorf("entry insert batch 100 requires at least one entry")
	}
	if len(entries) > 100 {
		return sqlstore.InsertBlobEntriesByScopeAnd100EntriesParams{}, fmt.Errorf("entry insert batch 100 got %d entries", len(entries))
	}

	params := sqlstore.InsertBlobEntriesByScopeAnd100EntriesParams{
		Namespace: scope.Namespace,
		Subject:   int64(scope.Subject),
		BlobID:    int64(blobID),
	}
	for i, entry := range entries {
		if err := setInsertEntryBatch100Slot(&params, i, entry); err != nil {
			return sqlstore.InsertBlobEntriesByScopeAnd100EntriesParams{}, err
		}
	}
	return params, nil
}

func setEntryBatch10Slot(params *sqlstore.GetBlobEntriesByScopeAnd10EntriesParams, index int, entry sqliteEntryPair) error {
	id := entry.ID
	metaTag := entry.MetaTag
	switch index {
	case 0:
		params.ID00 = id
		params.MetaTag00 = metaTag
	case 1:
		params.ID01 = id
		params.MetaTag01 = metaTag
	case 2:
		params.ID02 = id
		params.MetaTag02 = metaTag
	case 3:
		params.ID03 = id
		params.MetaTag03 = metaTag
	case 4:
		params.ID04 = id
		params.MetaTag04 = metaTag
	case 5:
		params.ID05 = id
		params.MetaTag05 = metaTag
	case 6:
		params.ID06 = id
		params.MetaTag06 = metaTag
	case 7:
		params.ID07 = id
		params.MetaTag07 = metaTag
	case 8:
		params.ID08 = id
		params.MetaTag08 = metaTag
	case 9:
		params.ID09 = id
		params.MetaTag09 = metaTag
	default:
		return fmt.Errorf("entry batch 10 slot %d exceeds capacity", index)
	}
	return nil
}

func setInsertEntryBatch10Slot(params *sqlstore.InsertBlobEntriesByScopeAnd10EntriesParams, index int, entry sqliteEntryPair) error {
	id := entry.ID
	metaTag := entry.MetaTag
	switch index {
	case 0:
		params.ID00 = id
		params.MetaTag00 = metaTag
	case 1:
		params.ID01 = id
		params.MetaTag01 = metaTag
	case 2:
		params.ID02 = id
		params.MetaTag02 = metaTag
	case 3:
		params.ID03 = id
		params.MetaTag03 = metaTag
	case 4:
		params.ID04 = id
		params.MetaTag04 = metaTag
	case 5:
		params.ID05 = id
		params.MetaTag05 = metaTag
	case 6:
		params.ID06 = id
		params.MetaTag06 = metaTag
	case 7:
		params.ID07 = id
		params.MetaTag07 = metaTag
	case 8:
		params.ID08 = id
		params.MetaTag08 = metaTag
	case 9:
		params.ID09 = id
		params.MetaTag09 = metaTag
	default:
		return fmt.Errorf("entry insert batch 10 slot %d exceeds capacity", index)
	}
	return nil
}

func setEntryBatch100Slot(params *sqlstore.GetBlobEntriesByScopeAnd100EntriesParams, index int, entry sqliteEntryPair) error {
	id := entry.ID
	metaTag := entry.MetaTag
	switch index {
	case 0:
		params.ID00 = id
		params.MetaTag00 = metaTag
	case 1:
		params.ID01 = id
		params.MetaTag01 = metaTag
	case 2:
		params.ID02 = id
		params.MetaTag02 = metaTag
	case 3:
		params.ID03 = id
		params.MetaTag03 = metaTag
	case 4:
		params.ID04 = id
		params.MetaTag04 = metaTag
	case 5:
		params.ID05 = id
		params.MetaTag05 = metaTag
	case 6:
		params.ID06 = id
		params.MetaTag06 = metaTag
	case 7:
		params.ID07 = id
		params.MetaTag07 = metaTag
	case 8:
		params.ID08 = id
		params.MetaTag08 = metaTag
	case 9:
		params.ID09 = id
		params.MetaTag09 = metaTag
	case 10:
		params.ID10 = id
		params.MetaTag10 = metaTag
	case 11:
		params.ID11 = id
		params.MetaTag11 = metaTag
	case 12:
		params.ID12 = id
		params.MetaTag12 = metaTag
	case 13:
		params.ID13 = id
		params.MetaTag13 = metaTag
	case 14:
		params.ID14 = id
		params.MetaTag14 = metaTag
	case 15:
		params.ID15 = id
		params.MetaTag15 = metaTag
	case 16:
		params.ID16 = id
		params.MetaTag16 = metaTag
	case 17:
		params.ID17 = id
		params.MetaTag17 = metaTag
	case 18:
		params.ID18 = id
		params.MetaTag18 = metaTag
	case 19:
		params.ID19 = id
		params.MetaTag19 = metaTag
	case 20:
		params.ID20 = id
		params.MetaTag20 = metaTag
	case 21:
		params.ID21 = id
		params.MetaTag21 = metaTag
	case 22:
		params.ID22 = id
		params.MetaTag22 = metaTag
	case 23:
		params.ID23 = id
		params.MetaTag23 = metaTag
	case 24:
		params.ID24 = id
		params.MetaTag24 = metaTag
	case 25:
		params.ID25 = id
		params.MetaTag25 = metaTag
	case 26:
		params.ID26 = id
		params.MetaTag26 = metaTag
	case 27:
		params.ID27 = id
		params.MetaTag27 = metaTag
	case 28:
		params.ID28 = id
		params.MetaTag28 = metaTag
	case 29:
		params.ID29 = id
		params.MetaTag29 = metaTag
	case 30:
		params.ID30 = id
		params.MetaTag30 = metaTag
	case 31:
		params.ID31 = id
		params.MetaTag31 = metaTag
	case 32:
		params.ID32 = id
		params.MetaTag32 = metaTag
	case 33:
		params.ID33 = id
		params.MetaTag33 = metaTag
	case 34:
		params.ID34 = id
		params.MetaTag34 = metaTag
	case 35:
		params.ID35 = id
		params.MetaTag35 = metaTag
	case 36:
		params.ID36 = id
		params.MetaTag36 = metaTag
	case 37:
		params.ID37 = id
		params.MetaTag37 = metaTag
	case 38:
		params.ID38 = id
		params.MetaTag38 = metaTag
	case 39:
		params.ID39 = id
		params.MetaTag39 = metaTag
	case 40:
		params.ID40 = id
		params.MetaTag40 = metaTag
	case 41:
		params.ID41 = id
		params.MetaTag41 = metaTag
	case 42:
		params.ID42 = id
		params.MetaTag42 = metaTag
	case 43:
		params.ID43 = id
		params.MetaTag43 = metaTag
	case 44:
		params.ID44 = id
		params.MetaTag44 = metaTag
	case 45:
		params.ID45 = id
		params.MetaTag45 = metaTag
	case 46:
		params.ID46 = id
		params.MetaTag46 = metaTag
	case 47:
		params.ID47 = id
		params.MetaTag47 = metaTag
	case 48:
		params.ID48 = id
		params.MetaTag48 = metaTag
	case 49:
		params.ID49 = id
		params.MetaTag49 = metaTag
	case 50:
		params.ID50 = id
		params.MetaTag50 = metaTag
	case 51:
		params.ID51 = id
		params.MetaTag51 = metaTag
	case 52:
		params.ID52 = id
		params.MetaTag52 = metaTag
	case 53:
		params.ID53 = id
		params.MetaTag53 = metaTag
	case 54:
		params.ID54 = id
		params.MetaTag54 = metaTag
	case 55:
		params.ID55 = id
		params.MetaTag55 = metaTag
	case 56:
		params.ID56 = id
		params.MetaTag56 = metaTag
	case 57:
		params.ID57 = id
		params.MetaTag57 = metaTag
	case 58:
		params.ID58 = id
		params.MetaTag58 = metaTag
	case 59:
		params.ID59 = id
		params.MetaTag59 = metaTag
	case 60:
		params.ID60 = id
		params.MetaTag60 = metaTag
	case 61:
		params.ID61 = id
		params.MetaTag61 = metaTag
	case 62:
		params.ID62 = id
		params.MetaTag62 = metaTag
	case 63:
		params.ID63 = id
		params.MetaTag63 = metaTag
	case 64:
		params.ID64 = id
		params.MetaTag64 = metaTag
	case 65:
		params.ID65 = id
		params.MetaTag65 = metaTag
	case 66:
		params.ID66 = id
		params.MetaTag66 = metaTag
	case 67:
		params.ID67 = id
		params.MetaTag67 = metaTag
	case 68:
		params.ID68 = id
		params.MetaTag68 = metaTag
	case 69:
		params.ID69 = id
		params.MetaTag69 = metaTag
	case 70:
		params.ID70 = id
		params.MetaTag70 = metaTag
	case 71:
		params.ID71 = id
		params.MetaTag71 = metaTag
	case 72:
		params.ID72 = id
		params.MetaTag72 = metaTag
	case 73:
		params.ID73 = id
		params.MetaTag73 = metaTag
	case 74:
		params.ID74 = id
		params.MetaTag74 = metaTag
	case 75:
		params.ID75 = id
		params.MetaTag75 = metaTag
	case 76:
		params.ID76 = id
		params.MetaTag76 = metaTag
	case 77:
		params.ID77 = id
		params.MetaTag77 = metaTag
	case 78:
		params.ID78 = id
		params.MetaTag78 = metaTag
	case 79:
		params.ID79 = id
		params.MetaTag79 = metaTag
	case 80:
		params.ID80 = id
		params.MetaTag80 = metaTag
	case 81:
		params.ID81 = id
		params.MetaTag81 = metaTag
	case 82:
		params.ID82 = id
		params.MetaTag82 = metaTag
	case 83:
		params.ID83 = id
		params.MetaTag83 = metaTag
	case 84:
		params.ID84 = id
		params.MetaTag84 = metaTag
	case 85:
		params.ID85 = id
		params.MetaTag85 = metaTag
	case 86:
		params.ID86 = id
		params.MetaTag86 = metaTag
	case 87:
		params.ID87 = id
		params.MetaTag87 = metaTag
	case 88:
		params.ID88 = id
		params.MetaTag88 = metaTag
	case 89:
		params.ID89 = id
		params.MetaTag89 = metaTag
	case 90:
		params.ID90 = id
		params.MetaTag90 = metaTag
	case 91:
		params.ID91 = id
		params.MetaTag91 = metaTag
	case 92:
		params.ID92 = id
		params.MetaTag92 = metaTag
	case 93:
		params.ID93 = id
		params.MetaTag93 = metaTag
	case 94:
		params.ID94 = id
		params.MetaTag94 = metaTag
	case 95:
		params.ID95 = id
		params.MetaTag95 = metaTag
	case 96:
		params.ID96 = id
		params.MetaTag96 = metaTag
	case 97:
		params.ID97 = id
		params.MetaTag97 = metaTag
	case 98:
		params.ID98 = id
		params.MetaTag98 = metaTag
	case 99:
		params.ID99 = id
		params.MetaTag99 = metaTag
	default:
		return fmt.Errorf("entry batch 100 slot %d exceeds capacity", index)
	}
	return nil
}

func setInsertEntryBatch100Slot(params *sqlstore.InsertBlobEntriesByScopeAnd100EntriesParams, index int, entry sqliteEntryPair) error {
	id := entry.ID
	metaTag := entry.MetaTag
	switch index {
	case 0:
		params.ID00 = id
		params.MetaTag00 = metaTag
	case 1:
		params.ID01 = id
		params.MetaTag01 = metaTag
	case 2:
		params.ID02 = id
		params.MetaTag02 = metaTag
	case 3:
		params.ID03 = id
		params.MetaTag03 = metaTag
	case 4:
		params.ID04 = id
		params.MetaTag04 = metaTag
	case 5:
		params.ID05 = id
		params.MetaTag05 = metaTag
	case 6:
		params.ID06 = id
		params.MetaTag06 = metaTag
	case 7:
		params.ID07 = id
		params.MetaTag07 = metaTag
	case 8:
		params.ID08 = id
		params.MetaTag08 = metaTag
	case 9:
		params.ID09 = id
		params.MetaTag09 = metaTag
	case 10:
		params.ID10 = id
		params.MetaTag10 = metaTag
	case 11:
		params.ID11 = id
		params.MetaTag11 = metaTag
	case 12:
		params.ID12 = id
		params.MetaTag12 = metaTag
	case 13:
		params.ID13 = id
		params.MetaTag13 = metaTag
	case 14:
		params.ID14 = id
		params.MetaTag14 = metaTag
	case 15:
		params.ID15 = id
		params.MetaTag15 = metaTag
	case 16:
		params.ID16 = id
		params.MetaTag16 = metaTag
	case 17:
		params.ID17 = id
		params.MetaTag17 = metaTag
	case 18:
		params.ID18 = id
		params.MetaTag18 = metaTag
	case 19:
		params.ID19 = id
		params.MetaTag19 = metaTag
	case 20:
		params.ID20 = id
		params.MetaTag20 = metaTag
	case 21:
		params.ID21 = id
		params.MetaTag21 = metaTag
	case 22:
		params.ID22 = id
		params.MetaTag22 = metaTag
	case 23:
		params.ID23 = id
		params.MetaTag23 = metaTag
	case 24:
		params.ID24 = id
		params.MetaTag24 = metaTag
	case 25:
		params.ID25 = id
		params.MetaTag25 = metaTag
	case 26:
		params.ID26 = id
		params.MetaTag26 = metaTag
	case 27:
		params.ID27 = id
		params.MetaTag27 = metaTag
	case 28:
		params.ID28 = id
		params.MetaTag28 = metaTag
	case 29:
		params.ID29 = id
		params.MetaTag29 = metaTag
	case 30:
		params.ID30 = id
		params.MetaTag30 = metaTag
	case 31:
		params.ID31 = id
		params.MetaTag31 = metaTag
	case 32:
		params.ID32 = id
		params.MetaTag32 = metaTag
	case 33:
		params.ID33 = id
		params.MetaTag33 = metaTag
	case 34:
		params.ID34 = id
		params.MetaTag34 = metaTag
	case 35:
		params.ID35 = id
		params.MetaTag35 = metaTag
	case 36:
		params.ID36 = id
		params.MetaTag36 = metaTag
	case 37:
		params.ID37 = id
		params.MetaTag37 = metaTag
	case 38:
		params.ID38 = id
		params.MetaTag38 = metaTag
	case 39:
		params.ID39 = id
		params.MetaTag39 = metaTag
	case 40:
		params.ID40 = id
		params.MetaTag40 = metaTag
	case 41:
		params.ID41 = id
		params.MetaTag41 = metaTag
	case 42:
		params.ID42 = id
		params.MetaTag42 = metaTag
	case 43:
		params.ID43 = id
		params.MetaTag43 = metaTag
	case 44:
		params.ID44 = id
		params.MetaTag44 = metaTag
	case 45:
		params.ID45 = id
		params.MetaTag45 = metaTag
	case 46:
		params.ID46 = id
		params.MetaTag46 = metaTag
	case 47:
		params.ID47 = id
		params.MetaTag47 = metaTag
	case 48:
		params.ID48 = id
		params.MetaTag48 = metaTag
	case 49:
		params.ID49 = id
		params.MetaTag49 = metaTag
	case 50:
		params.ID50 = id
		params.MetaTag50 = metaTag
	case 51:
		params.ID51 = id
		params.MetaTag51 = metaTag
	case 52:
		params.ID52 = id
		params.MetaTag52 = metaTag
	case 53:
		params.ID53 = id
		params.MetaTag53 = metaTag
	case 54:
		params.ID54 = id
		params.MetaTag54 = metaTag
	case 55:
		params.ID55 = id
		params.MetaTag55 = metaTag
	case 56:
		params.ID56 = id
		params.MetaTag56 = metaTag
	case 57:
		params.ID57 = id
		params.MetaTag57 = metaTag
	case 58:
		params.ID58 = id
		params.MetaTag58 = metaTag
	case 59:
		params.ID59 = id
		params.MetaTag59 = metaTag
	case 60:
		params.ID60 = id
		params.MetaTag60 = metaTag
	case 61:
		params.ID61 = id
		params.MetaTag61 = metaTag
	case 62:
		params.ID62 = id
		params.MetaTag62 = metaTag
	case 63:
		params.ID63 = id
		params.MetaTag63 = metaTag
	case 64:
		params.ID64 = id
		params.MetaTag64 = metaTag
	case 65:
		params.ID65 = id
		params.MetaTag65 = metaTag
	case 66:
		params.ID66 = id
		params.MetaTag66 = metaTag
	case 67:
		params.ID67 = id
		params.MetaTag67 = metaTag
	case 68:
		params.ID68 = id
		params.MetaTag68 = metaTag
	case 69:
		params.ID69 = id
		params.MetaTag69 = metaTag
	case 70:
		params.ID70 = id
		params.MetaTag70 = metaTag
	case 71:
		params.ID71 = id
		params.MetaTag71 = metaTag
	case 72:
		params.ID72 = id
		params.MetaTag72 = metaTag
	case 73:
		params.ID73 = id
		params.MetaTag73 = metaTag
	case 74:
		params.ID74 = id
		params.MetaTag74 = metaTag
	case 75:
		params.ID75 = id
		params.MetaTag75 = metaTag
	case 76:
		params.ID76 = id
		params.MetaTag76 = metaTag
	case 77:
		params.ID77 = id
		params.MetaTag77 = metaTag
	case 78:
		params.ID78 = id
		params.MetaTag78 = metaTag
	case 79:
		params.ID79 = id
		params.MetaTag79 = metaTag
	case 80:
		params.ID80 = id
		params.MetaTag80 = metaTag
	case 81:
		params.ID81 = id
		params.MetaTag81 = metaTag
	case 82:
		params.ID82 = id
		params.MetaTag82 = metaTag
	case 83:
		params.ID83 = id
		params.MetaTag83 = metaTag
	case 84:
		params.ID84 = id
		params.MetaTag84 = metaTag
	case 85:
		params.ID85 = id
		params.MetaTag85 = metaTag
	case 86:
		params.ID86 = id
		params.MetaTag86 = metaTag
	case 87:
		params.ID87 = id
		params.MetaTag87 = metaTag
	case 88:
		params.ID88 = id
		params.MetaTag88 = metaTag
	case 89:
		params.ID89 = id
		params.MetaTag89 = metaTag
	case 90:
		params.ID90 = id
		params.MetaTag90 = metaTag
	case 91:
		params.ID91 = id
		params.MetaTag91 = metaTag
	case 92:
		params.ID92 = id
		params.MetaTag92 = metaTag
	case 93:
		params.ID93 = id
		params.MetaTag93 = metaTag
	case 94:
		params.ID94 = id
		params.MetaTag94 = metaTag
	case 95:
		params.ID95 = id
		params.MetaTag95 = metaTag
	case 96:
		params.ID96 = id
		params.MetaTag96 = metaTag
	case 97:
		params.ID97 = id
		params.MetaTag97 = metaTag
	case 98:
		params.ID98 = id
		params.MetaTag98 = metaTag
	case 99:
		params.ID99 = id
		params.MetaTag99 = metaTag
	default:
		return fmt.Errorf("entry insert batch 100 slot %d exceeds capacity", index)
	}
	return nil
}
