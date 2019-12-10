package datastore

import (
	"cloud.google.com/go/datastore"
	"context"
	"os"
)

var (
	dataStoreClient *datastore.Client
	ErrNoSuchEntity = datastore.ErrNoSuchEntity
)

type Key datastore.Key
type PendingKey datastore.PendingKey
type Transaction struct {
	t *datastore.Transaction
}

func init() {
	var err error
	dataStoreClient, err = datastore.NewClient(context.Background(), os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		panic(err)
	}
}

func (k *Key) IntID() int64 {
	return k.ID
}

func (k *Key) StringID() string {
	return k.Name
}

func (t *Transaction) DeleteMulti(keys []*Key) (err error) {
	return t.t.DeleteMulti(getDataStoreKeys(keys))
}

func (t *Transaction) Delete(key *Key) error {
	return t.t.Delete((*datastore.Key)(key))
}

func (t *Transaction) PutMulti(keys []*Key, src interface{}) (ret []*PendingKey, err error) {
	dKeys, err := t.t.PutMulti(getDataStoreKeys(keys), src)
	return getOwnPendingKeys(dKeys), err
}

func (t *Transaction) Put(key *Key, src interface{}) (*PendingKey, error) {
	dKey, err := t.t.Put((*datastore.Key)(key), src)
	return (*PendingKey)(dKey), err
}

func (t *Transaction) GetMulti(keys []*Key, dst interface{}) (err error) {
	return t.t.GetMulti(getDataStoreKeys(keys), dst)
}

func (t *Transaction) Get(key *Key, dst interface{}) (err error) {
	return t.t.Get((*datastore.Key)(key), dst)
}

func (t *Transaction) Commit() (err error) {
	_, err = t.t.Commit()
	return err
}

// NewKey creates a new key.
// kind cannot be empty.
// Either one or both of stringID and intID must be zero. If both are zero,
// the key returned is incomplete.
// parent must either be a complete key or nil.
func NewKey(c context.Context, kind, stringID string, intID int64, parent *Key) *Key {
	parentKey := (*datastore.Key)(parent)
	theKey := &datastore.Key{
		Kind:   kind,
		Name:   stringID,
		ID:     intID,
		Parent: parentKey,
	}

	return (*Key)(theKey)
}

// NewIncompleteKey creates a new incomplete key.
// kind cannot be empty.
func NewIncompleteKey(ctx context.Context, kind string, parent *Key) *Key {
	return NewKey(ctx, kind, "", 0, parent)
}

func NewQuery(kind string) *datastore.Query {
	return datastore.NewQuery(kind)
}

func GetAll(ctx context.Context, q *datastore.Query, dst interface{}) (keys []*Key, err error) {
	dKeys, err := dataStoreClient.GetAll(ctx, q, dst)
	keys = getOwnKeys(dKeys)
	return keys, err
}

func Get(ctx context.Context, key *Key, dst interface{}) (err error) {
	return dataStoreClient.Get(ctx, (*datastore.Key)(key), dst)
}

func GetMulti(ctx context.Context, keys []*Key, dst interface{}) (err error) {
	dKeys := getDataStoreKeys(keys)
	return dataStoreClient.GetMulti(ctx, dKeys, dst)
}

func Put(ctx context.Context, key *Key, src interface{}) (*Key, error) {
	k, err := dataStoreClient.Put(ctx, (*datastore.Key)(key), src)
	return (*Key)(k), err
}

func PutMulti(ctx context.Context, keys []*Key, src interface{}) (ret []*Key, err error) {
	dKeys := getDataStoreKeys(keys)
	dKeys, err = dataStoreClient.PutMulti(ctx, dKeys, src)
	keys = getOwnKeys(dKeys)
	return keys, err
}

func Delete(ctx context.Context, key *Key) error {
	return dataStoreClient.Delete(ctx, (*datastore.Key)(key))
}

func DeleteMulti(ctx context.Context, keys []*Key) (err error) {
	return dataStoreClient.DeleteMulti(ctx, getDataStoreKeys(keys))
}

func Run(ctx context.Context, q *datastore.Query) *datastore.Iterator {
	return dataStoreClient.Run(ctx, q)
}

func Count(ctx context.Context, q *datastore.Query) (n int, err error) {

	return dataStoreClient.Count(ctx, q)
}

func DecodeCursor(s string) (datastore.Cursor, error) {

	return datastore.DecodeCursor(s)
}

// RunInTransaction runs f in a transaction. f is invoked with a Transaction
// that f should use for all the transaction's datastore operations.
//
// f must not call Commit or Rollback on the provided Transaction.
//
// If f returns nil, RunInTransaction commits the transaction,
// returning the Commit and a nil error if it succeeds. If the commit fails due
// to a conflicting transaction, RunInTransaction retries f with a new
// Transaction. It gives up and returns ErrConcurrentTransaction after three
// failed attempts (or as configured with MaxAttempts).
//
// If f returns non-nil, then the transaction will be rolled back and
// RunInTransaction will return the same error. The function f is not retried.
//
// Note that when f returns, the transaction is not committed. Calling code
// must not assume that any of f's changes have been committed until
// RunInTransaction returns nil.
//
// Since f may be called multiple times, f should usually be idempotent – that
// is, it should have the same result when called multiple times. Note that
// Transaction.Get will append when unmarshalling slice fields, so it is not
// necessarily idempotent.
func RunInTransaction(ctx context.Context, f func(tx *Transaction) error) (err error) {
	_, err = dataStoreClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		return f(&Transaction{tx})
	}, nil)
	return err
}

func getOwnPendingKeys(keys []*datastore.PendingKey) []*PendingKey {
	ownKeys := make([]*PendingKey, len(keys))
	for i, key := range keys {
		ownKeys[i] = (*PendingKey)(key)
	}

	return ownKeys
}

func getOwnKeys(keys []*datastore.Key) []*Key {
	ownKeys := make([]*Key, len(keys))
	for i, key := range keys {
		ownKeys[i] = (*Key)(key)
	}

	return ownKeys
}

func getDataStoreKeys(keys []*Key) []*datastore.Key {
	dKeys := make([]*datastore.Key, len(keys))
	for i, key := range keys {
		dKeys[i] = (*datastore.Key)(key)
	}

	return dKeys
}