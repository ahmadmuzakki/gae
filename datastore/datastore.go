package datastore

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func Put(ctx context.Context, key *Key, src interface{}) (*Key, error) {
	if mock, ok := isMock(ctx); ok {
		return mock.put(ctx, key, src)
	}

	dsKey := convertKeyToDsKey(ctx, key)
	k, err := datastore.Put(ctx, dsKey, src)
	return convertDsKeyToKey(ctx, k), err
}

func Get(ctx context.Context, key *Key, dst interface{}) error {
	if mock, ok := isMock(ctx); ok {
		return mock.get(ctx, key, dst)
	}

	dsKey := convertKeyToDsKey(ctx, key)
	return datastore.Get(ctx, dsKey, dst)
}
