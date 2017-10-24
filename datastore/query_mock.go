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

func (mq *MockQuery) run(ctx context.Context, q *Query) (it *Iterator) {
	if len(mq.mocks) == 0 {
		it.err = fmt.Errorf("No more expectations")
		return
	}

	mock := mq.mocks[0]

	if !reflect.DeepEqual(mock.query, q) {
		it.err = fmt.Errorf("Query %+v did not match with expected %+v", q, mock.query)
		return
	}

	ctx = mq.setValue(ctx, mock.expectation)
	it.c = ctx

	mq.mocks = mq.mocks[1:]
	return
}

func (mq *MockQuery) getAll(ctx context.Context, q *Query, dst interface{}) ([]*Key, error) {
	if reflect.TypeOf(dst).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("%s", "Destination should be pointer")
	}

	sliceDest := reflect.Indirect(reflect.ValueOf(dst))
	if sliceDest.Type().Kind() != reflect.Slice {
		return nil, fmt.Errorf("%s", "Destination is not array")
	}

	if len(mq.mocks) == 0 {
		return nil, fmt.Errorf("No more expectations")
	}

	mock := mq.mocks[0]
	for _, expect := range mock.expectation {
		// get the slice item Type
		itemType := sliceDest.Type().Elem()
		// new row of slice element
		// we use indirect because itemType is pointer
		newRow := reflect.Indirect(reflect.New(itemType))
		val := reflect.ValueOf(expect.Value)

		if val.Type().Kind() != reflect.Ptr {
			return nil, fmt.Errorf("Expected value should be pointer")
		}

		newRow.Set(reflect.Indirect(val))

		sliceDest.Set(reflect.Append(sliceDest, newRow))
	}

	return nil, nil
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

type MockQueryAction struct {
	query       *Query
	expectation []QueryExpectation
}

type QueryExpectation struct {
	Key   *Key
	Value interface{}
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

func (action *MockQueryAction) KeysOnly() *MockQueryAction {
	q := action.query.clone()
	q.keysOnly = true
	action.query = q
	return action
}

func (action *MockQueryAction) Limit(limit int) *MockQueryAction {
	q := action.query.clone()
	q.limit = int32(limit)
	action.query = q
	return action
}

func (action *MockQueryAction) Offset(offset int) *MockQueryAction {
	q := action.query.clone()
	q.offset = int32(offset)
	action.query = q
	return action
}

func (action *MockQueryAction) Start(c Cursor) *MockQueryAction {
	q := action.query.clone()
	q.cursor = c

	action.query = q
	return action
}

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
