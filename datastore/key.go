package datastore

import (
	"context"
	"github.com/ahmadmuzakki/gae/internal"
	"google.golang.org/appengine/datastore"
)

type Key struct {
	kind      string
	stringID  string
	intID     int64
	parent    *Key
	appID     string
	namespace string
	dsKey     *datastore.Key
}

func NewKey(ctx context.Context, kind string, stringID string, intID int64, parent *Key) *Key {
	if m, ok := isMock(ctx); ok {
		return m.newKey(ctx, kind, stringID, intID, parent)
	}

	k := &Key{
		kind:      kind,
		parent:    parent,
		intID:     intID,
		stringID:  stringID,
		namespace: internal.GetNamespace(ctx),
	}

	k.dsKey = convertKeyToDsKey(ctx, k)
	return k
}

func NewIncompleteKey(ctx context.Context, kind string, parent *Key) *Key {
	return NewKey(ctx, kind, "", 0, parent)
}

func convertKeyToDsKey(ctx context.Context, key *Key) *datastore.Key {
	if key == nil {
		return nil
	}

	parent := convertKeyToDsKey(ctx, key.parent)
	return datastore.NewKey(ctx, key.kind, key.stringID, key.intID, parent)
}

func convertDsKeyToKey(ctx context.Context, key *datastore.Key) *Key {
	if key == nil {
		return nil
	}

	parent := convertDsKeyToKey(ctx, key.Parent())
	return NewKey(ctx, key.Kind(), key.StringID(), key.IntID(), parent)
}

func (k *Key) String() string {
	if k.dsKey == nil {
		mock := MockKey(*k)
		return (&mock).String()
	}
	return k.dsKey.String()
}
