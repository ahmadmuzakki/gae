package nds

import (
	"github.com/ahmadmuzakki/gae/datastore"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
)

func Delete(ctx context.Context, key *datastore.Key) error {
	dskey := datastore.ConvertKeyToDsKey(ctx, key)
	return nds.Delete(ctx, dskey)
}
