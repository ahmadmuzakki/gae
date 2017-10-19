package datastore

import (
	"errors"
	"fmt"
	"github.com/ahmadmuzakki/gae/internal"
	gaemock "github.com/ahmadmuzakki/gae/mock"
	"golang.org/x/net/context"
	"reflect"
)

type DatastoreMock struct {
	mocks []*MockAction
	keys  []*Key
}

type MockAction struct {
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

func Mock(ctx context.Context) (context.Context, *DatastoreMock) {
	gaemock.ValidateContext(ctx)
	mock := &DatastoreMock{}
	ctx = context.WithValue(ctx, "datastore_mock", mock)
	return ctx, mock
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
	ActionPut = "Put"
	ActionGet = "Get"
)

func (dm *DatastoreMock) put(ctx context.Context, key *Key, src interface{}) (*Key, error) {
	if len(dm.mocks) == 0 {
		return nil, errors.New("No more expectation")
	}

	mock := dm.mocks[0]
	if mock.action != ActionPut {
		return nil, fmt.Errorf("Action %s is not expected. Expected action is %s", mock.action, ActionPut)
	}

	if !reflect.DeepEqual(mock.key, key) {
		return nil, errors.New("Key not equal")
	}

	if ns := internal.GetNamespace(ctx); mock.namespace != ns {
		return nil, fmt.Errorf("Expected to called with namespace %s but current namespace is %s", mock.namespace, ns)
	}

	if !reflect.DeepEqual(mock.param, src) {
		return nil, fmt.Errorf("Source %+v doesn't match with %+v", src, mock.param)
	}

	dm.trimMock()

	return mock.key, nil
}

func (dm *DatastoreMock) get(ctx context.Context, key *Key, dst interface{}) error {
	if err := dm.checkExpectations(); err != nil {
		return err
	}

	mock := dm.mocks[0]

	if err := mock.checkAction(ActionGet); err != nil {
		return err
	}

	if err := mock.checkKey(key); err != nil {
		return err
	}

	if err := mock.checkNamespace(ctx); err != nil {
		return err
	}

	// only check the param since dst is empty struct
	typeDest := reflect.TypeOf(dst)
	typeParam := reflect.TypeOf(mock.param)
	if !reflect.DeepEqual(typeDest, typeParam) {
		return fmt.Errorf("Source %+v doesn't match with %+v", typeDest, typeParam)
	}

	// assign value from *mock.param to *dst
	// direct the pointer first
	valDst := reflect.ValueOf(dst)
	directDst := reflect.Indirect(valDst)

	valParam := reflect.ValueOf(mock.param)
	directParam := reflect.Indirect(valParam)

	// set the pointer with the payload from *mock.param
	directDst.Set(directParam)

	dm.trimMock()
	return nil
}

func (dm *DatastoreMock) checkExpectations() error {
	if len(dm.mocks) == 0 {
		return errors.New("No more expectation")
	}
	return nil
}

func (mock *MockAction) checkAction(action string) error {
	if mock.action != action {
		return fmt.Errorf("Action %s is not expected. Expected action is %s", mock.action, action)
	}
	return nil
}

func (mock *MockAction) checkNamespace(ctx context.Context) error {
	if ns := internal.GetNamespace(ctx); mock.namespace != ns {
		return fmt.Errorf("Expected to called with namespace %s but current namespace is %s", mock.namespace, ns)
	}
	return nil
}

func (mock *MockAction) checkKey(key *Key) error {
	if !reflect.DeepEqual(mock.key, key) {
		return errors.New("Key not equal")
	}
	return nil
}

func (dm *DatastoreMock) MockPut(key *Key, src interface{}) *MockAction {
	m := &MockAction{
		action: ActionPut,
		param:  src,
		key:    key,
	}
	dm.appendMock(m)
	return m
}

func (dm *DatastoreMock) MockGet(key *Key, dst interface{}) *MockAction {
	m := &MockAction{
		action: ActionGet,
		param:  dst,
		key:    key,
	}
	dm.appendMock(m)
	return m
}

func (dm *DatastoreMock) appendMock(m *MockAction) {
	if dm.mocks == nil {
		dm.mocks = make([]*MockAction, 0)
	}

	dm.mocks = append(dm.mocks, m)
}

func (dm *DatastoreMock) trimMock() {
	if len(dm.mocks) > 1 {
		dm.mocks = dm.mocks[1:]
	} else {
		dm.mocks = nil
	}
}

func (m *MockAction) WithNameSpace(ns string) *MockAction {
	m.namespace = ns
	return m
}

func (m *MockAction) WillReturnKeyErr(key *Key, err error) *MockAction {
	m.expect.key = key
	m.expect.err = err
	return m
}

func (m *MockAction) ExpectValue(val interface{}) *MockAction {
	m.expect.value = val
	return m
}

func (m *MockAction) WillReturnErr(err error) {
	m.expect.err = err
}

func isMock(ctx context.Context) (*DatastoreMock, bool) {
	m, ok := ctx.Value("datastore_mock").(*DatastoreMock)
	return m, ok
}
