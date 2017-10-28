package nds

import (
	"github.com/ahmadmuzakki/gae/datastore"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
)

func Get(ctx context.Context, key *datastore.Key, val interface{}) error {
	dskey := datastore.ConvertKeyToDsKey(ctx, key)
	return nds.Get(ctx, dskey, val)
}
