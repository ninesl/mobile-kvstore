package kvstore_test

import (
	"bytes"
	"encoding/gob"
	"fmt"

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
const TESTINNERTYPE_NAMESPACE = "testInnerType"

// testType is a BLOB that is stored
type testType struct {
	id           int
	shared_id    int
	field_value  testInnerType
	field_ptr    *testInnerType
	fields_value []testInnerType
	fields_ptr   []*testInnerType
}

type testInnerType struct {
	ID    int
	Title string
}

func (t *testInnerType) ItemArgs() kvstore.ItemArgs {
	return kvstore.ItemArgs{}
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

func encodeTestInnerType(t any) ([]byte, error) {
	v, ok := t.(*testInnerType)
	if !ok {
		return nil, fmt.Errorf("encode testInnerType: expected *testInnerType, got %T", t)
	}

	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(*v); err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

// Assumes exact replica of types. No fields can be reordered, nothing can be renamed or adjusted AT ALL
// or will return nil, error
// type DecoderFunc func([]byte) (any, error)

// EncoderFunc should be given the most current data
// type EncoderFunc func(any) (ItemBlob, error)

func determineIdentities(t *testType) []kvstore.Identity {
	//for _, t := range t.fields_value
	return nil
}

func (t *testType) ItemArgs() kvstore.ItemArgs {
	return kvstore.ItemArgs{
		Scope: kvstore.Scope{
			Namespace: "testType",
			Subject:   t.id,
		},
	}
}

// we then need paths to get full code coverage on all errors.
// what's the idiomatic Go way to acomplish this
