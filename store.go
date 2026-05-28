package kvstore

import (
	"database/sql"
	"embed"

	"github.com/ninesl/kvstore/sqlstore"
)

//go:embed sqlc/schema.sql
var schemaFS embed.FS

// InitDB initializes the kvstore schema on conn.
func InitDB(conn *sql.DB) error {
	schema, err := schemaFS.ReadFile("sqlc/schema.sql")
	if err != nil {
		return err
	}

	_, err = conn.Exec(string(schema))
	return err
}

// ItemArgs stores one blob payload and one or more Identities to that blob.
// Store only calls Encode when the Identities do not already point to one shared blob.
// FIXME: currently overwrites the blob if it exists, no TTL logic
type ItemArgs struct {

	// Scope defines the type/kind of data
	// namespace MUST be compatible with EncoderFunc/DecoderFunc
	Scope Scope

	// Identities are the other entries that should refernce this Data.
	// Useful for grouping, suggest len(Identities) == 1 for a single entry
	Identities []Identity

	// The implementation should generally reference the underlying
	// consumer type that is implementing StoreItem
	//  func (m *myType) ItemArgs() {
	// 	return ItemArgs{Data: m}
	//  }
	Data any

	// The EncoderFunc needs to be consistent across Namespaces
	// or at least the callee assumes the EncoderFunc will be compat
	// with the Data/
	Encode EncoderFunc

	// Serialized key, usually should be the identifier for your type
	// used for data de-duplication
	BlobKey string

	// filters assumes one blob_id unlike Get() which gets multiple blob_ids of the same type
	//	TTLRules Filters // FIXME: needs to be better defined/assigned/mode??? not sure on abrstraction yet
}

// Your data type needs to implement ItemArgs
type StoreItem interface {
	ItemArgs() ItemArgs
}

// ClearX will also delete associated data internally
type Store interface {
	Store(StoreItem) error

	//Get(scope Scope, decode DecoderFunc) ([]any, error)
	GetByFilter(EntryFilter, DecoderFunc) ([]any, error)

	// returns all unique data of filter
	// skips null fields for filter,
	//
	// ex: to count all instance of an ID in a namespace
	//  store.Count(EntryFilter{
	// 	Namespace: &someNamespace,
	// 	ID: &someID,
	//  })
	Count(filter EntryFilter) (int, error)

	// ALL CLEAR FUNCTIONS CASCADE AND DELETE ALL ENTRIES THAT IT FITS
	// a Clear() will find the blob_id, delete the blob_ids from the blob_entry table
	// the schema should handle cascade deleting if we just delete from blobs?

	// removes ALL data that have this scope
	ClearByScope(scope Scope) error

	// remove ALL data that share this Identity
	ClearByIdentity(identity Identity) error

	// remove ALL data matching a non-empty EntryFilter
	ClearByFilter(filter EntryFilter) error

	// remove ALL data
	ClearEverything() error

	// delete the single Data source from the blobkey/serialized key
	// removes all Scopes/Identities referencing this data
	ClearByBlobKey(key string) error
}

// Expects a sqlite connection, can be a blank .db or one that was created by this package
func New(conn *sql.DB) Store {
	return storer{
		queries: sqlstore.New(conn),
	}
}

// Scope identifies the type and scan scope for stored blobs.
// Namespace is the decode/type boundary, and Subject is the entity being scanned.
type Scope struct {
	// the TYPE, all namespaces should be decoded/encoded the same
	Namespace string
	// the subject be entered, for filtering purposes
	Subject int
}

// Identity indexes one blob by one primitive integer key.
// MetaTag gives ID meaning within a namespace/subject
type Identity struct {
	ID      int
	MetaTag string
}

// Assumes exact replica of types. No fields can be reordered, nothing can be renamed or adjusted AT ALL
// or will return nil, error
type DecoderFunc func([]byte) (any, error)

// EncoderFunc should be given the most current data
type EncoderFunc func(any) ([]byte, error)

// null fields are not considered for filtering
type EntryFilter struct {
	Namespace *string
	Subject   *int
	ID        *int
	MetaTag   *string
}
