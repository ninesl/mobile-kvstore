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

const nilTestInnerTypeID = -1

const (
	filterNamespace = 1 << iota
	filterSubject
	filterID
	filterMetaTag
)

var (
	testNamespaceValues = []string{TESTTYPE_NAMESPACE, "testTypeAlt"}
	testSubjectValues   = []int{100, 200, 300}
	testIdentityValues  = []int{1, 2, 3, 1000}
	testMetaTagValues   = []string{"testType", "field_value", "field_ptr", "fields_value", "fields_ptr"}
)

func newTestStore(t *testing.T) kvstore.Store {
	t.Helper()

	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open test sqlite db: %v", err)
	}
	conn.SetMaxOpenConns(1)
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
	conn.SetMaxOpenConns(1)
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

	if len(got) != len(want) {
		t.Fatalf("decoded len = %d, want %d", len(got), len(want))
	}

	gotByID := make(map[int]*testType, len(got))
	for _, item := range got {
		decoded, ok := item.(*testType)
		if !ok {
			t.Fatalf("decoded item type = %T, want *testType", item)
		}
		if _, exists := gotByID[decoded.id]; exists {
			t.Fatalf("duplicate decoded testType %d", decoded.id)
		}
		gotByID[decoded.id] = decoded
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
	namespace := TESTTYPE_NAMESPACE
	return kvstore.EntryFilter{Namespace: &namespace}
}

func testTypeScopeFilter(subject int) kvstore.EntryFilter {
	namespace := TESTTYPE_NAMESPACE
	return kvstore.EntryFilter{Namespace: &namespace, Subject: &subject}
}

func requireDecodedByFilter(t *testing.T, store kvstore.Store, filter kvstore.EntryFilter, want ...*testType) {
	t.Helper()

	got, err := store.GetByFilter(filter, decodeTestType)
	if err != nil {
		t.Fatalf("get by filter: %v", err)
	}
	requireDecodedTestTypes(t, got, want...)
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
	id := 1
	metaTag := "testType"
	requireCount(t, store, kvstore.EntryFilter{ID: &id, MetaTag: &metaTag}, 1)
}

func TestStoreClearByScope(t *testing.T) {
	store := newTestStore(t)
	testTypes := storeMockTestTypes(t, store)

	if err := store.ClearByScope(kvstore.Scope{Namespace: TESTTYPE_NAMESPACE, Subject: 100}); err != nil {
		t.Fatalf("clear by scope: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 1)
	requireCount(t, store, testTypeScopeFilter(100), 0)
	requireCount(t, store, testTypeScopeFilter(200), 1)
	requireDecodedByFilter(t, store, testTypeNamespaceFilter(), testTypes[2])
}

func TestStoreClearByIdentity(t *testing.T) {
	store := newTestStore(t)
	testTypes := storeMockTestTypes(t, store)

	if err := store.ClearByIdentity(kvstore.Identity{ID: 1, MetaTag: "testType"}); err != nil {
		t.Fatalf("clear by identity: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 2)
	id := 1
	metaTag := "testType"
	requireCount(t, store, kvstore.EntryFilter{ID: &id, MetaTag: &metaTag}, 0)
	requireDecodedByFilter(t, store, testTypeNamespaceFilter(), testTypes[1], testTypes[2])
}

func TestStoreClearByFilter(t *testing.T) {
	store := newTestStore(t)
	testTypes := storeMockTestTypes(t, store)

	if err := store.ClearByFilter(testTypeScopeFilter(200)); err != nil {
		t.Fatalf("clear by filter: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 2)
	requireCount(t, store, testTypeScopeFilter(200), 0)
	requireDecodedByFilter(t, store, testTypeNamespaceFilter(), testTypes[0], testTypes[1])
}

func TestStoreClearEverything(t *testing.T) {
	store := newTestStore(t)
	storeMockTestTypes(t, store)

	if err := store.ClearEverything(); err != nil {
		t.Fatalf("clear everything: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 0)
	requireDecodedByFilter(t, store, testTypeNamespaceFilter())
}

func TestStoreClearByBlobKey(t *testing.T) {
	store := newTestStore(t)
	testTypes := storeMockTestTypes(t, store)

	if err := store.ClearByBlobKey(testTypes[0].ItemArgs().BlobKey); err != nil {
		t.Fatalf("clear by blob key: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 2)
	id := testTypes[0].id
	metaTag := "testType"
	requireCount(t, store, kvstore.EntryFilter{ID: &id, MetaTag: &metaTag}, 0)
	requireDecodedByFilter(t, store, testTypeNamespaceFilter(), testTypes[1], testTypes[2])
}

func TestStoreStoresHundredIdentityBatches(t *testing.T) {
	store := newTestStore(t)
	testType := mockTestTypeWithIdentities(105)

	if err := store.Store(testType); err != nil {
		t.Fatalf("store 100-plus identity testType: %v", err)
	}

	requireCount(t, store, testTypeNamespaceFilter(), 1)
	id1099 := 1_099
	id1104 := 1_104
	metaTag := "fields_value"
	requireCount(t, store, kvstore.EntryFilter{ID: &id1099, MetaTag: &metaTag}, 1)
	requireCount(t, store, kvstore.EntryFilter{ID: &id1104, MetaTag: &metaTag}, 1)
	requireDecodedByFilter(t, store, testTypeNamespaceFilter(), testType)
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

	id := testType.id
	metaTag := "testType"
	got, err := store.GetByFilter(kvstore.EntryFilter{ID: &id, MetaTag: &metaTag}, decodeTestType)
	if err != nil {
		t.Fatalf("get updated testType: %v", err)
	}
	requireDecodedTestTypes(t, got, testType)
}

func TestStoreGetByFilterAllPermutations(t *testing.T) {
	store := newTestStore(t)
	fixture := mockFilterPermutationFixture(t, store)

	namespace := testNamespaceValues[0]
	subject := testSubjectValues[0]
	id := testIdentityValues[0]
	metaTag := testMetaTagValues[0]
	for _, tc := range filterPermutations(namespace, subject, id, metaTag, fixture.associations) {
		t.Run(tc.name, func(t *testing.T) {
			requireCount(t, store, tc.filter, len(tc.want))
			requireDecodedByFilter(t, store, tc.filter, tc.want...)
		})
	}
}

func TestEncodeDecodeTestTypeRoundTrip(t *testing.T) {
	innerTypes := mockTestInnerTypes()
	testTypes := append(mockTestTypes(), &testType{
		id:           4,
		shared_id:    300,
		field_value:  innerTypes[0],
		field_ptr:    nil,
		fields_value: nil,
		fields_ptr:   []*testInnerType{nil, &innerTypes[1]},
	})

	for _, testType := range testTypes {
		data, err := encodeTestType(testType)
		if err != nil {
			t.Fatalf("encode testType %d: %v", testType.id, err)
		}
		got, err := decodeTestType(data)
		if err != nil {
			t.Fatalf("decode testType %d: %v", testType.id, err)
		}
		requireDecodedTestTypes(t, []any{got}, testType)
	}
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

type expectedAssociation struct {
	namespace string
	subject   int
	id        int
	metaTag   string
	item      *testType
}

type filterPermutationFixture struct {
	items        []*testType
	associations []expectedAssociation
}

type filterPermutationCase struct {
	name   string
	filter kvstore.EntryFilter
	want   []*testType
}

func storeAssociation(t *testing.T, store kvstore.Store, item *testType, namespace string, subject int, id int, metaTag string) expectedAssociation {
	t.Helper()

	args := item.ItemArgs()
	args.Scope = kvstore.Scope{Namespace: namespace, Subject: subject}
	args.Identities = []kvstore.Identity{{ID: id, MetaTag: metaTag}}
	if err := store.Store(testStoreItem{args: args}); err != nil {
		t.Fatalf("store association %s/%d/%d/%s: %v", namespace, subject, id, metaTag, err)
	}

	return expectedAssociation{
		namespace: namespace,
		subject:   subject,
		id:        id,
		metaTag:   metaTag,
		item:      item,
	}
}

func mockFilterPermutationFixture(t *testing.T, store kvstore.Store) filterPermutationFixture {
	t.Helper()

	items := mockTestTypes()
	namespace := testNamespaceValues[0]
	altNamespace := testNamespaceValues[1]
	subject := testSubjectValues[0]
	altSubject := testSubjectValues[1]
	id := testIdentityValues[0]
	altID := testIdentityValues[1]
	metaTag := testMetaTagValues[0]
	altMetaTag := testMetaTagValues[1]

	associations := []expectedAssociation{
		storeAssociation(t, store, items[0], namespace, subject, id, metaTag),
		storeAssociation(t, store, items[1], namespace, subject, altID, metaTag),
		storeAssociation(t, store, items[2], namespace, altSubject, id, metaTag),
		storeAssociation(t, store, items[0], altNamespace, subject, id, metaTag),
		storeAssociation(t, store, items[1], namespace, subject, id, altMetaTag),
		storeAssociation(t, store, items[2], namespace, subject, id, metaTag),
	}

	return filterPermutationFixture{items: items, associations: associations}
}

func filterPermutations(namespace string, subject int, id int, metaTag string, associations []expectedAssociation) []filterPermutationCase {
	cases := make([]filterPermutationCase, 0, 16)
	for mask := 0; mask < 16; mask++ {
		var filter kvstore.EntryFilter
		name := "filter"
		if mask == 0 {
			name += "_empty"
		}
		if mask&filterNamespace != 0 {
			filter.Namespace = &namespace
			name += "_namespace"
		}
		if mask&filterSubject != 0 {
			filter.Subject = &subject
			name += "_subject"
		}
		if mask&filterID != 0 {
			filter.ID = &id
			name += "_id"
		}
		if mask&filterMetaTag != 0 {
			filter.MetaTag = &metaTag
			name += "_metaTag"
		}

		cases = append(cases, filterPermutationCase{
			name:   name,
			filter: filter,
			want:   expectedForFilter(filter, associations),
		})
	}

	return cases
}

func expectedForFilter(filter kvstore.EntryFilter, associations []expectedAssociation) []*testType {
	seen := make(map[string]*testType)
	for _, association := range associations {
		if filter.Namespace != nil && association.namespace != *filter.Namespace {
			continue
		}
		if filter.Subject != nil && association.subject != *filter.Subject {
			continue
		}
		if filter.ID != nil && association.id != *filter.ID {
			continue
		}
		if filter.MetaTag != nil && association.metaTag != *filter.MetaTag {
			continue
		}
		seen[association.item.ItemArgs().BlobKey] = association.item
	}

	out := make([]*testType, 0, len(seen))
	for _, item := range seen {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ItemArgs().BlobKey < out[j].ItemArgs().BlobKey
	})

	return out
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

	var fieldPtr *testInnerType
	if wire.FieldPtr.ID != nilTestInnerTypeID {
		fieldPtr = &testInnerType{ID: wire.FieldPtr.ID, Title: wire.FieldPtr.Title}
	}

	fieldsPtr := make([]*testInnerType, len(wire.FieldsPtr))
	for i, field := range wire.FieldsPtr {
		if field.ID == nilTestInnerTypeID {
			continue
		}
		fieldsPtr[i] = &testInnerType{ID: field.ID, Title: field.Title}
	}

	return &testType{
		id:           wire.ID,
		shared_id:    wire.SharedID,
		field_value:  wire.FieldValue,
		field_ptr:    fieldPtr,
		fields_value: wire.FieldsValue,
		fields_ptr:   fieldsPtr,
	}, nil
}

type testTypeWire struct {
	ID          int
	SharedID    int
	FieldValue  testInnerType
	FieldPtr    testInnerTypeWire
	FieldsValue []testInnerType
	FieldsPtr   []testInnerTypeWire
}

type testInnerTypeWire struct {
	ID    int
	Title string
}

func encodeTestType(t any) ([]byte, error) {
	v, ok := t.(*testType)
	if !ok {
		return nil, fmt.Errorf("encode testType: expected *testType, got %T", t)
	}
	fieldPtr := testInnerTypeWire{ID: nilTestInnerTypeID}
	if v.field_ptr != nil {
		fieldPtr = testInnerTypeWire{ID: v.field_ptr.ID, Title: v.field_ptr.Title}
	}

	fieldsPtr := make([]testInnerTypeWire, len(v.fields_ptr))
	for i, innerType := range v.fields_ptr {
		if innerType == nil {
			fieldsPtr[i] = testInnerTypeWire{ID: nilTestInnerTypeID}
			continue
		}
		fieldsPtr[i] = testInnerTypeWire{ID: innerType.ID, Title: innerType.Title}
	}

	wire := testTypeWire{
		ID:          v.id,
		SharedID:    v.shared_id,
		FieldValue:  v.field_value,
		FieldPtr:    fieldPtr,
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
