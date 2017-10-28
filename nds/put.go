package nds

import (
	"github.com/ahmadmuzakki/gae/datastore"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	ds "google.golang.org/appengine/datastore"
)

func Put(ctx context.Context, key *datastore.Key, val interface{}) (*datastore.Key, error) {
	dskey := datastore.ConvertKeyToDsKey(ctx, key)
	dskey, err := nds.Put(ctx, dskey, val)
	key = datastore.ConvertDsKeyToKey(dskey)
	return key, err
}

func PutMulti(ctx context.Context, keys []*datastore.Key, vals interface{}) ([]*datastore.Key, error) {
	dsKeys := make([]*ds.Key, len(keys))
	for i := range dsKeys {
		dsKeys[i] = datastore.ConvertKeyToDsKey(ctx, keys[i])
	}

	dsKeys, err := nds.PutMulti(ctx, dsKeys, vals)
	if err != nil {
		return nil, err
	}

	for i := range keys {
		keys[i] = datastore.ConvertDsKeyToKey(dsKeys[i])
	}
	return keys, nil
}
