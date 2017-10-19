package datastore

import (
	"errors"
	"fmt"
	"github.com/ahmadmuzakki/gae/internal"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"reflect"
)

type DatastoreMock struct {
	mocks []*Mock
	keys  []*Key
}

type Mock struct {
	namespace string
	action    string
	key       *Key
	param     interface{}
	expect    expectation
}

type MockKey Key

func (k *MockKey) String() string {
	id := k.stringID
	if id == "" {
		id = fmt.Sprint(k.intID)
	}
	return fmt.Sprintf("/%s,%s", k.kind, id)
}

func (dm *DatastoreMock) MockIncompleteKey(ctx context.Context, kind string, parent *Key) *Key {
	return dm.MockKey(ctx, kind, "", 0, parent)
}

func (dm *DatastoreMock) MockKey(ctx context.Context, kind string, stringID string, intID int64, parent *Key) *Key {
	k := &Key{
		kind:      kind,
		parent:    parent,
		intID:     intID,
		stringID:  stringID,
		namespace: internal.GetNamespace(ctx),
	}

	if dm.keys == nil {
		dm.keys = make([]*Key, 0)
	}
	dm.keys = append(dm.keys, k)

	return k
}

func (dm *DatastoreMock) newKey(ctx context.Context, kind string, stringID string, intID int64, parent *Key) *Key {
	if len(dm.keys) == 0 {
		return nil
	}

	k := dm.keys[0]
	if len(dm.keys) > 1 {
		dm.keys = dm.keys[1:]
	} else {
		dm.keys = nil
	}
	return k
}

type expectation struct {
	key   *Key
	err   error
	value interface{}
}

const (
	PutAction = "Put"
)

func (dm *DatastoreMock) put(ctx context.Context, key *Key, src interface{}) (*Key, error) {
	if len(dm.mocks) == 0 {
		return nil, errors.New("No more expectation")
	}

	mock := dm.mocks[0]
	if mock.action != PutAction {
		return nil, fmt.Errorf("Action %s is not expected", mock.action)
	}

	if !reflect.DeepEqual(mock.key, key) {
		return nil, errors.New("Key not equal")
	}

	if ns := internal.GetNamespace(ctx); mock.namespace != ns {
		return nil, fmt.Errorf("Expected to called with namespace %s but current namespace is %s", mock.namespace, ns)
	}

	if len(dm.mocks) > 1 {
		dm.mocks = dm.mocks[1:]
	} else {
		dm.mocks = nil
	}

	return mock.key, nil
}

func (dm *DatastoreMock) Get(ctx context.Context, key *Key, dst interface{}) (*datastore.Key, error) {
	return nil, nil
}

func (dm *DatastoreMock) MockPut(key *Key, src interface{}) *Mock {
	m := &Mock{
		action: PutAction,
		param:  src,
		key:    key,
	}
	dm.appendMock(m)
	return m
}

func (dm *DatastoreMock) appendMock(m *Mock) {
	if dm.mocks == nil {
		dm.mocks = make([]*Mock, 0)
	}

	dm.mocks = append(dm.mocks, m)
}

func (m *Mock) WithNameSpace(ns string) *Mock {
	m.namespace = ns
	return m
}

func (m *Mock) WillReturnKeyErr(key *Key, err error) *Mock {
	m.expect.key = key
	m.expect.err = err
	return m
}

func (m *Mock) ExpectValue(val interface{}) *Mock {
	m.expect.value = val
	return m
}

func isMock(ctx context.Context) (*DatastoreMock, bool) {
	m, ok := ctx.Value("datastore_mock").(*DatastoreMock)
	return m, ok
}
