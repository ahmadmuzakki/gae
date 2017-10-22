package datastore

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"reflect"
	"strings"
)

type MockQuery struct {
	mocks   []*MockQueryAction
	cursors []Cursor
}

type storage struct {
	action MockQueryAction
}

type MockQueryAction struct {
	query       *Query
	expectation []QueryExpectation
}

type QueryExpectation struct {
	Key   *Key
	Value interface{}
}

func NewMockQuery(ctx context.Context) (context.Context, *MockQuery) {
	mock := &MockQuery{}
	ctx = context.WithValue(ctx, "mock_query", mock)
	return ctx, mock
}

func (mq *MockQuery) ExpectQuery(kind string) *MockQueryAction {
	mock := &MockQueryAction{
		query: &Query{
			kind:  kind,
			limit: -1,
		},
	}
	mq.mocks = append(mq.mocks, mock)
	return mock
}

func (a *MockQueryAction) ExpectResult(results ...QueryExpectation) {
	a.expectation = results
}

func (action *MockQueryAction) Ancestor(ancestor *Key) *MockQueryAction {
	q := action.query.clone()
	q.ancestor = ancestor
	action.query = q
	return action
}

func (action *MockQueryAction) Filter(filterStr string, value interface{}) *MockQueryAction {
	q := action.query.clone()
	q.filter = append(q.filter, filter{Field: filterStr, Value: value})
	action.query = q
	return action
}

func (action *MockQueryAction) Order(fieldName string) *MockQueryAction {
	q := action.query.clone()
	fieldName = strings.TrimSpace(fieldName)
	q.order = append(q.order, fieldName)
	action.query = q
	return action
}

func (action *MockQueryAction) Project(fieldNames ...string) *MockQueryAction {
	q := action.query.clone()
	q.projection = append([]string(nil), fieldNames...)
	action.query = q
	return action
}

func (action *MockQueryAction) Distinct() *MockQueryAction {
	q := action.query.clone()
	q.distinct = true
	action.query = q
	return action
}

// KeysOnly returns a derivative query that yields only keys, not keys and
// entities. It cannot be used with projection queries.
func (action *MockQueryAction) KeysOnly() *MockQueryAction {
	q := action.query.clone()
	q.keysOnly = true
	action.query = q
	return action
}

// Limit returns a derivative query that has a limit on the number of results
// returned. A negative value means unlimited.
func (action *MockQueryAction) Limit(limit int) *MockQueryAction {
	q := action.query.clone()
	q.limit = int32(limit)
	action.query = q
	return action
}

// Offset returns a derivative query that has an offset of how many keys to
// skip over before returning results. A negative value is invalid.
func (action *MockQueryAction) Offset(offset int) *MockQueryAction {
	q := action.query.clone()
	q.offset = int32(offset)
	action.query = q
	return action
}

// Start returns a derivative query with the given start point.
func (action *MockQueryAction) Start(c Cursor) *MockQueryAction {
	q := action.query.clone()
	q.cursor = c

	action.query = q
	return action
}

// End returns a derivative query with the given end point.
func (action *MockQueryAction) End(c Cursor) *MockQueryAction {
	q := action.query.clone()
	q.cursor = c

	action.query = q
	return action
}

func (action *MockQueryAction) Count(c context.Context) (int, error) {
	/*// intercept for mock
	action.query = q
	return action.queryDs.Count(c)*/
	return 0, nil
}

func (action *MockQueryAction) GetAll(c context.Context, dst interface{}) ([]*Key, error) {
	return nil, nil
}

func (mq *MockQuery) run(ctx context.Context, q *Query) *Iterator {
	mock := mq.mocks[0]

	var err error
	if !reflect.DeepEqual(mock.query, q) {
		err = fmt.Errorf("Query %+v did not match with expected %+v", q, mock.query)
	}

	ctx = mq.setValue(ctx, mock.expectation)
	return &Iterator{
		c:   ctx,
		err: err,
	}
}

func (mq *MockQuery) setValue(ctx context.Context, value []QueryExpectation) context.Context {
	return context.WithValue(ctx, "expected_value", value)
}

func (mq *MockQuery) getValue(ctx context.Context) []QueryExpectation {
	return ctx.Value("expected_value").([]QueryExpectation)
}

func isMockQuery(ctx context.Context) (*MockQuery, bool) {
	if mock, ok := ctx.Value("mock_query").(*MockQuery); ok {
		return mock, ok
	}
	return nil, false
}

func (mq *MockQuery) next(i *Iterator, dst interface{}) (*Key, error) {
	values := mq.getValue(i.c)

	if i.index == len(values) {
		return nil, datastore.Done
	}

	dir := reflect.Indirect(reflect.ValueOf(dst))
	expect := values[i.index]
	value := reflect.ValueOf(expect.Value)
	directValue := reflect.Indirect(value)
	dir.Set(directValue)
	i.index += 1
	return expect.Key, nil
}

func (mq *MockQuery) MockCursor(str string) {
	c := Cursor{
		cursorStr: str,
	}
	mq.cursors = append(mq.cursors, c)
}

func (mq *MockQuery) decodeCursor(ctx context.Context, str string) (Cursor, error) {

	if len(mq.cursors) == 0 {
		return Cursor{}, fmt.Errorf("Cursor with string %s is not expected", str)
	}

	c := mq.cursors[0]
	mq.cursors = mq.cursors[1:]
	return c, nil
}

var Done = datastore.Done
