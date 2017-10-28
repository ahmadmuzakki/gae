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

	k.dsKey = ConvertKeyToDsKey(ctx, k)
	return k
}

func NewIncompleteKey(ctx context.Context, kind string, parent *Key) *Key {
	return NewKey(ctx, kind, "", 0, parent)
}

func ConvertKeyToDsKey(ctx context.Context, key *Key) *datastore.Key {
	if key == nil {
		return nil
	}

	parent := ConvertKeyToDsKey(ctx, key.parent)
	return datastore.NewKey(ctx, key.kind, key.stringID, key.intID, parent)
}

func ConvertDsKeyToKey(key *datastore.Key) *Key {
	if key == nil {
		return nil
	}

	parent := ConvertDsKeyToKey(key.Parent())

	k := &Key{
		kind:      key.Kind(),
		parent:    parent,
		intID:     key.IntID(),
		stringID:  key.StringID(),
		namespace: key.Namespace(),
	}
	return k
}

func convertDsKeysToKeys(ctx context.Context, dsKeys []*datastore.Key) []*Key {
	newKeys := make([]*Key, 0, len(dsKeys))
	for _, dsKey := range dsKeys {
		key := ConvertDsKeyToKey(dsKey)
		newKeys = append(newKeys, key)
	}

	return newKeys
}

func (k *Key) String() string {
	if k.dsKey == nil {
		mock := MockKey(*k)
		return (&mock).String()
	}
	return k.dsKey.String()
}

func (k *Key) Encode() string {
	if k.dsKey == nil {
		mock := MockKey(*k)
		return (&mock).Encode()
	}
	return k.dsKey.Encode()
}

func (k *Key) Parent() *Key {
	return k.parent
}

func (k *Key) IntID() int64 {
	return k.dsKey.IntID()
}

func DecodeKey(encoded string) (*Key, error) {
	dskey, err := datastore.DecodeKey(encoded)
	if err != nil {
		return nil, err
	}

	key := ConvertDsKeyToKey(dskey)
	return key, nil
}
