package kvstore

// BatchScope identifies the type and scan scope for stored blobs.
// Namespace is the decode/type boundary, and Subject is the entity being scanned.
type BatchScope struct {
	Namespace string // the TYPE, all namespaces can be decoded/encoded the same
	Subject   int    // the subject be entered, for filtering purposes
}

// ItemRef indexes one blob by one primitive integer key.
// MetaTag gives ID meaning within a namespace/subject
type ItemRef struct {
	ID      int
	MetaTag string
}

type ItemBlob struct {
	// this key defines the data of the Blob. Usually fmt.Sprintf("%#v")
	// WARNING: you must keep the fields the same in the struct or the SerializedKey will be different
	SerializedKey string
	Data          []byte
	// the underlying data to be stored that is consumer defined
	// this value can be and SHOULD be shared between your namespaces/IDs
	// Expects a SINGLE undlering data type (embedded field, etc) not a slice types
}

// Assumes exact replica of types. No fields can be reordered, nothing can be renamed or adjusted AT ALL
// or will return nil, error
type DecoderFunc func([]byte) (any, error)

// EncoderFunc should be given the most current data
type EncoderFunc func(any) (ItemBlob, error)

// ItemArgs stores one blob payload and one or more refs to that blob.
// Store only calls Encode when the refs do not already point to one shared blob.
type ItemArgs struct {
	Scope  BatchScope
	Refs   []ItemRef
	Data   any
	Encode EncoderFunc
	// filters assumes one blob_id unlike Get() which gets multiple blob_ids of the same type
	//	TTLRules Filters // FIXME: needs to be better defined/assigned/mode??? not sure on abrstraction yet
}

// Your data type needs to implement ItemArgs
type StoreItem interface {
	ItemArgs() ItemArgs
}

type Filters struct {
	Scope BatchScope
	Refs  ItemRef
}

type Store interface {
	Store(StoreItem) error
	Get(namespace string, subject int, decode DecoderFunc) ([]any, error)
	GetByRef(Filters, DecoderFunc) ([]any, error)
}
