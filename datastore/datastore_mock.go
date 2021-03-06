package datastore

import (
	"errors"
	"fmt"
	"github.com/ahmadmuzakki/gae/internal"
	gaemock "github.com/ahmadmuzakki/gae/mock"
	"golang.org/x/net/context"
	"log"
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

func (k *MockKey) Encode() string {
	return k.String()
}

func NewMock(ctx context.Context) (context.Context, *DatastoreMock) {
	gaemock.ValidateContext(ctx)
	mock := &DatastoreMock{}
	ctx = context.WithValue(ctx, "datastore_mock", mock)
	return ctx, mock
}

func (dm *DatastoreMock) MockIncompleteKey(ctx context.Context, kind string, parent *Key) *Key {
	return dm.ExpectKey(ctx, kind, "", 0, parent)
}

func (dm *DatastoreMock) ExpectKey(ctx context.Context, kind string, stringID string, intID int64, parent *Key) *Key {
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

	if kind != k.kind {
		log.Printf("[ERROR] Mismatch Kind. Expected %s vs Actual %s \n", k.kind, kind)
		return nil
	}

	if stringID != k.stringID {
		log.Printf("[ERROR] Mismatch StringID. Expected %s vs Actual %s \n", k.stringID, stringID)
		return nil
	}

	if intID != k.intID {
		log.Printf("[ERROR] Mismatch IntID. Expected %d vs Actual %d \n", k.intID, intID)
		return nil
	}

	if !reflect.DeepEqual(parent, k.parent) {
		log.Printf("[ERROR] Mismatch Parent. Expected %+v vs Actual %+v \n", k.parent, parent)
		return nil
	}

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
	if err := dm.checkExpectations(); err != nil {
		return nil, err
	}

	mock := dm.mocks[0]
	defer dm.trimMock()

	if err := mock.checkAction(ActionPut); err != nil {
		return nil, err
	}

	if err := mock.checkKey(key); err != nil {
		return nil, err
	}

	if err := mock.checkNamespace(ctx); err != nil {
		return nil, err
	}

	if err := shouldRunInTransaction(ctx); err != nil {
		return nil, err
	}

	if err := mock.checkValue(ctx, src); err != nil {
		return nil, err
	}

	return mock.expect.key, mock.expect.err
}

func (dm *DatastoreMock) get(ctx context.Context, key *Key, dst interface{}) error {
	if err := dm.checkExpectations(); err != nil {
		return err
	}

	mock := dm.mocks[0]
	defer dm.trimMock()

	if err := mock.checkAction(ActionGet); err != nil {
		return err
	}

	if err := mock.checkKey(key); err != nil {
		return err
	}

	if err := mock.checkNamespace(ctx); err != nil {
		return err
	}

	if err := shouldRunInTransaction(ctx); err != nil {
		return err
	}

	// only check the param since dst is empty struct
	typeDest := reflect.TypeOf(dst)
	typeParam := reflect.TypeOf(mock.param)
	if !reflect.DeepEqual(typeDest, typeParam) {
		return fmt.Errorf("Destination %+v doesn't match with %+v", typeDest, typeParam)
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
		return fmt.Errorf("Key %+v doesn't match with %+v", mock.key, key)
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

func (m *MockAction) checkValue(ctx context.Context, param interface{}) error {
	if !reflect.DeepEqual(m.param, param) {
		return fmt.Errorf("Param %+v doens't equal with expectation %+v", param, m.param)
	}
	return nil
}

func isMock(ctx context.Context) (*DatastoreMock, bool) {
	m, ok := ctx.Value("datastore_mock").(*DatastoreMock)
	return m, ok
}
