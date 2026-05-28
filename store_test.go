package kvstore_test

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"reflect"
	"sort"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ninesl/kvstore"
)

// mock scope and identity data inline
// create a sqlite inmemory db.conn
// this is used for New(conn) to then insert the items
//
// unit test happy path of each interface method of Store
//
// we want roughly 2000 items, 500 of them are duplicates and will
// not be reinserted bc of the blobkey serialization
//
// we need a test data type that uses a test-level encode() decode() implementation for it as well
//
// can validate decode output via custom comparison of each embedded field's value
//
// each type should be something like

// use these
//
//	for id := range MAX_TEST_IDS {
//
//	}
const MAX_TEST_IDS = 10

const TESTTYPE_NAMESPACE = "testType"

const TESTTYPEMANAGER_NAMESPACE = "testTypeManager"

func newTestStore(t *testing.T) kvstore.Store {
	t.Helper()

	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open test sqlite db: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("close test sqlite db: %v", err)
		}
	})

	if err := kvstore.InitDB(conn); err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	return kvstore.New(conn)
}

func newClosedTestStore(t *testing.T) kvstore.Store {
	t.Helper()

	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open closed test sqlite db: %v", err)
	}
	if err := kvstore.InitDB(conn); err != nil {
		t.Fatalf("create closed test schema: %v", err)
	}
	store := kvstore.New(conn)
	if err := conn.Close(); err != nil {
		t.Fatalf("close test sqlite db: %v", err)
	}

	return store
}

func mockTestInnerTypes() []testInnerType {
	innerTypes := make([]testInnerType, MAX_TEST_IDS)
	for id := range MAX_TEST_IDS {
		innerTypes[id] = testInnerType{
			ID:    id + 1,
			Title: fmt.Sprintf("test inner type %02d", id+1),
		}
	}

	return innerTypes
}

func mockTestTypes() []*testType {
	innerTypes := mockTestInnerTypes()

	return []*testType{
		{
			id:           1,
			shared_id:    100,
			field_value:  innerTypes[0],
			field_ptr:    &innerTypes[1],
			fields_value: []testInnerType{innerTypes[2], innerTypes[3], innerTypes[4]},
			fields_ptr:   []*testInnerType{&innerTypes[5], &innerTypes[6]},
		},
		{
			id:           2,
			shared_id:    100,
			field_value:  innerTypes[3],
			field_ptr:    &innerTypes[4],
			fields_value: []testInnerType{innerTypes[5], innerTypes[6]},
			fields_ptr:   []*testInnerType{&innerTypes[7], &innerTypes[8], &innerTypes[9]},
		},
		{
			id:           3,
			shared_id:    200,
			field_value:  innerTypes[6],
			field_ptr:    &innerTypes[7],
			fields_value: []testInnerType{innerTypes[8], innerTypes[9], innerTypes[0]},
			fields_ptr:   []*testInnerType{&innerTypes[1], &innerTypes[2], &innerTypes[3]},
		},
	}
}

func mockTestTypeWithIdentities(identityCount int) *testType {
	testType := mockTestTypes()[0]
	testType.fields_value = make([]testInnerType, identityCount)
	for id := range identityCount {
		testType.fields_value[id] = testInnerType{
			ID:    id + 1_000,
			Title: fmt.Sprintf("batch identity %03d", id+1),
		}
	}
	testType.fields_ptr = nil

	return testType
}

func storeMockTestTypes(t *testing.T, store kvstore.Store) []*testType {
	t.Helper()

	testTypes := mockTestTypes()
	for _, testType := range testTypes {
		if err := store.Store(testType); err != nil {
			t.Fatalf("store testType %d: %v", testType.id, err)
		}
	}

	return testTypes
}

func requireCount(t *testing.T, store kvstore.Store, filter kvstore.EntryFilter, want int) {
	t.Helper()

	got, err := store.Count(filter)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}

func requireDecodedTestTypes(t *testing.T, got []any, want ...*testType) {
	t.Helper()

	gotByID := make(map[int]*testType, len(got))
	for _, item := range got {
		decoded, ok := item.(*testType)
		if !ok {
			t.Fatalf("decoded item type = %T, want *testType", item)
		}
		gotByID[decoded.id] = decoded
	}

	if len(gotByID) != len(want) {
		t.Fatalf("decoded len = %d, want %d", len(gotByID), len(want))
	}
	for _, expected := range want {
		actual, ok := gotByID[expected.id]
		if !ok {
			t.Fatalf("missing decoded testType %d", expected.id)
		}
		if !testTypesEqual(actual, expected) {
			t.Fatalf("decoded testType %d = %+v, want %+v", expected.id, actual, expected)
		}
	}
}

func testTypesEqual(a, b *testType) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.id != b.id || a.shared_id != b.shared_id || a.field_value != b.field_value {
		return false
	}
	if (a.field_ptr == nil) != (b.field_ptr == nil) {
		return false
	}
	if a.field_ptr != nil && *a.field_ptr != *b.field_ptr {
		return false
	}
	if !reflect.DeepEqual(a.fields_value, b.fields_value) {
		return false
	}
	if len(a.fields_ptr) != len(b.fields_ptr) {
		return false
	}
	for i := range a.fields_ptr {
		if (a.fields_ptr[i] == nil) != (b.fields_ptr[i] == nil) {
			return false
		}
		if a.fields_ptr[i] != nil && *a.fields_ptr[i] != *b.fields_ptr[i] {
			return false
		}
	}
	return true
}

func testTypeNamespaceFilter() kvstore.EntryFilter {
	return kvstore.EntryFilter{Namespace: ptr(TESTTYPE_NAMESPACE)}
}

func testTypeScopeFilter(subject int) kvstore.EntryFilter {
	return kvstore.EntryFilter{Namespace: ptr(TESTTYPE_NAMESPACE), Subject: &subject}
}

func ptr[T any](v T) *T {
	return &v
}

func sortedIDs(testTypes []*testType) []int {
	ids := make([]int, len(testTypes))
	for i, testType := range testTypes {
		ids[i] = testType.id
	}
	sort.Ints(ids)
	return ids
}

func TestStoreStoreAndGetByFilter(t *testing.T) {
	store := newTestStore(t)
	testTypes := storeMockTestTypes(t, store)

	got, err := store.GetByFilter(testTypeNamespaceFilter(), decodeTestType)
	if err != nil {
		t.Fatalf("get by filter: %v", err)
	}

	requireDecodedTestTypes(t, got, testTypes...)
}

func TestStoreCount(t *testing.T) {
	store := newTestStore(t)
	storeMockTestTypes(t, store)

	requireCount(t, store, testTypeNamespaceFilter(), 3)
	requireCount(t, store, testTypeScopeFilter(100), 2)
	requireCount(t, store, kvstore.EntryFilter{ID: ptr(1), MetaTag: ptr("testType")}, 1)
}

func TestStoreClearByScope(t *testing.T) {
	store := newTestStore(t)
	storeMockTestTypes(t, store)

	if err := store.ClearByScope(kvstore.Scope{Namespace: TESTTYPE_NAMESPACE, Subject: 100}); err != nil {
		t.Fatalf("clear by scope: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 1)
	requireCount(t, store, testTypeScopeFilter(100), 0)
	requireCount(t, store, testTypeScopeFilter(200), 1)
}

func TestStoreClearByIdentity(t *testing.T) {
	store := newTestStore(t)
	storeMockTestTypes(t, store)

	if err := store.ClearByIdentity(kvstore.Identity{ID: 1, MetaTag: "testType"}); err != nil {
		t.Fatalf("clear by identity: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 2)
	requireCount(t, store, kvstore.EntryFilter{ID: ptr(1), MetaTag: ptr("testType")}, 0)
}

func TestStoreClearByFilter(t *testing.T) {
	store := newTestStore(t)
	storeMockTestTypes(t, store)

	if err := store.ClearByFilter(testTypeScopeFilter(200)); err != nil {
		t.Fatalf("clear by filter: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 2)
	requireCount(t, store, testTypeScopeFilter(200), 0)
}

func TestStoreClearEverything(t *testing.T) {
	store := newTestStore(t)
	storeMockTestTypes(t, store)

	if err := store.ClearEverything(); err != nil {
		t.Fatalf("clear everything: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 0)
}

func TestStoreClearByBlobKey(t *testing.T) {
	store := newTestStore(t)
	testTypes := storeMockTestTypes(t, store)

	if err := store.ClearByBlobKey(testTypes[0].ItemArgs().BlobKey); err != nil {
		t.Fatalf("clear by blob key: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 2)
	requireCount(t, store, kvstore.EntryFilter{ID: ptr(testTypes[0].id), MetaTag: ptr("testType")}, 0)
}

func TestStoreStoresHundredIdentityBatches(t *testing.T) {
	store := newTestStore(t)
	testType := mockTestTypeWithIdentities(105)

	if err := store.Store(testType); err != nil {
		t.Fatalf("store 100-plus identity testType: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 1)
	requireCount(t, store, kvstore.EntryFilter{ID: ptr(1_099), MetaTag: ptr("fields_value")}, 1)
	requireCount(t, store, kvstore.EntryFilter{ID: ptr(1_104), MetaTag: ptr("fields_value")}, 1)
}

func TestStoreOverwritesExistingBlob(t *testing.T) {
	store := newTestStore(t)
	testType := mockTestTypes()[0]
	if err := store.Store(testType); err != nil {
		t.Fatalf("store initial testType: %v", err)
	}

	testType.field_value = testInnerType{ID: 500, Title: "updated field value"}
	if err := store.Store(testType); err != nil {
		t.Fatalf("store updated testType: %v", err)
	}

	got, err := store.GetByFilter(kvstore.EntryFilter{ID: ptr(testType.id), MetaTag: ptr("testType")}, decodeTestType)
	if err != nil {
		t.Fatalf("get updated testType: %v", err)
	}
	requireDecodedTestTypes(t, got, testType)
}

func TestStoreErrorsOnMissingBlobKey(t *testing.T) {
	store := newTestStore(t)
	testType := mockTestTypes()[0]
	testType.id = 0

	item := testStoreItem{args: testType.ItemArgs()}
	item.args.BlobKey = ""

	if err := store.Store(item); err == nil {
		t.Fatal("store missing blob key error = nil")
	}
}

func TestStoreErrorsOnMissingIdentities(t *testing.T) {
	store := newTestStore(t)
	testType := mockTestTypes()[0]

	item := testStoreItem{args: testType.ItemArgs()}
	item.args.Identities = nil

	if err := store.Store(item); err == nil {
		t.Fatal("store missing identities error = nil")
	}
}

func TestStoreErrorsOnEncodeFailure(t *testing.T) {
	store := newTestStore(t)
	testType := mockTestTypes()[0]

	item := testStoreItem{args: testType.ItemArgs()}
	item.args.Encode = func(any) ([]byte, error) {
		return nil, fmt.Errorf("expected encode failure")
	}

	if err := store.Store(item); err == nil {
		t.Fatal("store encode failure error = nil")
	}
}

func TestStoreErrorsOnOverwriteEncodeFailure(t *testing.T) {
	store := newTestStore(t)
	testType := mockTestTypes()[0]
	if err := store.Store(testType); err != nil {
		t.Fatalf("store initial testType: %v", err)
	}

	item := testStoreItem{args: testType.ItemArgs()}
	item.args.Encode = func(any) ([]byte, error) {
		return nil, fmt.Errorf("expected overwrite encode failure")
	}

	if err := store.Store(item); err == nil {
		t.Fatal("store overwrite encode failure error = nil")
	}
}

func TestStoreGetByFilterReturnsDecodeError(t *testing.T) {
	store := newTestStore(t)
	storeMockTestTypes(t, store)

	got, err := store.GetByFilter(testTypeNamespaceFilter(), func([]byte) (any, error) {
		return nil, fmt.Errorf("expected decode failure")
	})
	if err == nil {
		t.Fatal("get by filter decode error = nil")
	}
	if len(got) != 0 {
		t.Fatalf("decoded len = %d, want 0", len(got))
	}
}

func TestStoreClearByFilterErrorsOnEmptyFilter(t *testing.T) {
	store := newTestStore(t)

	if err := store.ClearByFilter(kvstore.EntryFilter{}); err == nil {
		t.Fatal("clear by empty filter error = nil")
	}
}

func TestStoreMethodsReturnDatabaseErrors(t *testing.T) {
	store := newClosedTestStore(t)
	testType := mockTestTypes()[0]

	if err := store.Store(testType); err == nil {
		t.Fatal("store closed database error = nil")
	}
	if _, err := store.GetByFilter(testTypeNamespaceFilter(), decodeTestType); err == nil {
		t.Fatal("get by filter closed database error = nil")
	}
	if _, err := store.Count(testTypeNamespaceFilter()); err == nil {
		t.Fatal("count closed database error = nil")
	}
	if err := store.ClearByScope(kvstore.Scope{Namespace: TESTTYPE_NAMESPACE, Subject: 100}); err == nil {
		t.Fatal("clear by scope closed database error = nil")
	}
	if err := store.ClearByIdentity(kvstore.Identity{ID: 1, MetaTag: "testType"}); err == nil {
		t.Fatal("clear by identity closed database error = nil")
	}
	if err := store.ClearByFilter(testTypeNamespaceFilter()); err == nil {
		t.Fatal("clear by filter closed database error = nil")
	}
	if err := store.ClearEverything(); err == nil {
		t.Fatal("clear everything closed database error = nil")
	}
	if err := store.ClearByBlobKey(testType.ItemArgs().BlobKey); err == nil {
		t.Fatal("clear by blob key closed database error = nil")
	}
}

// shared_id of testType should be the id of the testTypeManager
// testTypeManager is a BLOB that is stored
// but we use it by recreating the testType
// that is stored for each one
type testTypeManager struct {
	id        int
	testTypes []*testType
}

type testStoreItem struct {
	args kvstore.ItemArgs
}

func (t testStoreItem) ItemArgs() kvstore.ItemArgs {
	return t.args
}

func (t *testTypeManager) ItemArgs() kvstore.ItemArgs {
	return kvstore.ItemArgs{
		Scope: kvstore.Scope{
			Namespace: TESTTYPEMANAGER_NAMESPACE,
			Subject:   t.id,
		},
		Identities: []kvstore.Identity{
			{
				ID:      t.id,
				MetaTag: "testTypeManager",
			},
		},
		Encode:  encodeTestTypeManager,
		Data:    t,
		BlobKey: fmt.Sprintf("%s:%d", TESTTYPEMANAGER_NAMESPACE, t.id),
	}
}

func encodeTestTypeManager(t any) ([]byte, error) {
	v, ok := t.(*testTypeManager)
	if !ok {
		return nil, fmt.Errorf("encode testTypeManager: expected *testTypeManager, got %T", t)
	}
	_ = v

	// need a way t oencode the testTypeManager
	// that will be able to rebuild the testType
	// but does not stored the testType
	// because we want dedeuplication of testType
	// our decode logic will need to be able to
	// decode the testTypeManager and then rebuild
	// the testType via the blobkey?
	// so we could store a slice of blobkeys
	// to be able to rebuild the testType
	// or we store the blob_ids of the testType
	// but that is very coupled to the implementation
	// but is that any better than storing the blobkey?

	return nil, nil
}

// testType is a BLOB that is stored
type testType struct {
	id           int
	shared_id    int
	field_value  testInnerType
	field_ptr    *testInnerType
	fields_value []testInnerType
	fields_ptr   []*testInnerType
}

func (t *testType) ItemArgs() kvstore.ItemArgs {
	return kvstore.ItemArgs{
		Scope: kvstore.Scope{
			Namespace: TESTTYPE_NAMESPACE,
			Subject:   t.shared_id,
		},
		Identities: determineTestTypeIdentities(t),
		Encode:     encodeTestType,
		Data:       t,
		BlobKey:    fmt.Sprintf("%s:%d:%d", TESTTYPE_NAMESPACE, t.id, t.shared_id),
	}
}

func determineTestTypeIdentities(t *testType) []kvstore.Identity {
	identities := []kvstore.Identity{
		{ID: t.id, MetaTag: "testType"},
		{ID: t.field_value.ID, MetaTag: "field_value"},
	}
	if t.field_ptr != nil {
		identities = append(identities, kvstore.Identity{ID: t.field_ptr.ID, MetaTag: "field_ptr"})
	}
	for _, field := range t.fields_value {
		identities = append(identities, kvstore.Identity{ID: field.ID, MetaTag: "fields_value"})
	}
	for _, field := range t.fields_ptr {
		if field != nil {
			identities = append(identities, kvstore.Identity{ID: field.ID, MetaTag: "fields_ptr"})
		}
	}

	return identities
}

func decodeTestType(data []byte) (any, error) {
	var wire testTypeWire
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&wire); err != nil {
		return nil, err
	}

	fieldsPtr := make([]*testInnerType, len(wire.FieldsPtr))
	for i := range wire.FieldsPtr {
		fieldsPtr[i] = &wire.FieldsPtr[i]
	}

	return &testType{
		id:           wire.ID,
		shared_id:    wire.SharedID,
		field_value:  wire.FieldValue,
		field_ptr:    &wire.FieldPtr,
		fields_value: wire.FieldsValue,
		fields_ptr:   fieldsPtr,
	}, nil
}

type testTypeWire struct {
	ID          int
	SharedID    int
	FieldValue  testInnerType
	FieldPtr    testInnerType
	FieldsValue []testInnerType
	FieldsPtr   []testInnerType
}

func encodeTestType(t any) ([]byte, error) {
	v, ok := t.(*testType)
	if !ok {
		return nil, fmt.Errorf("encode testType: expected *testType, got %T", t)
	}
	if v.field_ptr == nil {
		return nil, fmt.Errorf("encode testType: field_ptr is nil")
	}

	fieldsPtr := make([]testInnerType, len(v.fields_ptr))
	for i, innerType := range v.fields_ptr {
		if innerType == nil {
			return nil, fmt.Errorf("encode testType: fields_ptr[%d] is nil", i)
		}
		fieldsPtr[i] = *innerType
	}

	wire := testTypeWire{
		ID:          v.id,
		SharedID:    v.shared_id,
		FieldValue:  v.field_value,
		FieldPtr:    *v.field_ptr,
		FieldsValue: v.fields_value,
		FieldsPtr:   fieldsPtr,
	}

	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(wire); err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

type testInnerType struct {
	ID    int
	Title string
}

// Assumes exact replica of types. No fields can be reordered, nothing can be renamed or adjusted AT ALL
// or will return nil, error
// type DecoderFunc func([]byte) (any, error)

// EncoderFunc should be given the most current data
// type EncoderFunc func(any) (ItemBlob, error)

// we then need paths to get full code coverage on all errors.
// what's the idiomatic Go way to acomplish this
