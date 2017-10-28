package datastore

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

var ErrNoSuchEntity = datastore.ErrNoSuchEntity

func Put(ctx context.Context, key *Key, src interface{}) (*Key, error) {
	if mock, ok := isMock(ctx); ok {
		return mock.put(ctx, key, src)
	}

	dsKey := ConvertKeyToDsKey(ctx, key)
	k, err := datastore.Put(ctx, dsKey, src)
	return ConvertDsKeyToKey(k), err
}

func Get(ctx context.Context, key *Key, dst interface{}) error {
	if mock, ok := isMock(ctx); ok {
		return mock.get(ctx, key, dst)
	}

	dsKey := ConvertKeyToDsKey(ctx, key)
	return datastore.Get(ctx, dsKey, dst)
}

func PutMulti(ctx context.Context, keys []*Key, src interface{}) ([]*Key, error) {
	dsKeys := make([]*datastore.Key, len(keys))
	for i := range dsKeys {
		dsKeys[i] = ConvertKeyToDsKey(ctx, keys[i])
	}

	dsKeys, err := datastore.PutMulti(ctx, dsKeys, src)
	if err != nil {
		return nil, err
	}

	for i := range keys {
		keys[i] = ConvertDsKeyToKey(dsKeys[i])
	}
	return keys, nil
}

func AllocateIDs(ctx context.Context, kind string, parent *Key, n int) (low, high int64, err error) {
	var parentds *datastore.Key
	if parent != nil {
		parentds = ConvertKeyToDsKey(ctx, parent)
	}
	return datastore.AllocateIDs(ctx, kind, parentds, n)
}
