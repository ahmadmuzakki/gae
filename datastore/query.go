package datastore

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"strings"
)

type filter struct {
	Field string
	Value interface{}
}

func NewQuery(ctx context.Context, kind string) *Query {
	query := &Query{
		kind:  kind,
		limit: -1,
	}
	if _, ok := isMockQuery(ctx); !ok {
		query.queryDs = datastore.NewQuery(kind)
	}
	return query
}

// Query represents a datastore query.
type Query struct {
	kind       string
	ancestor   *Key
	filter     []filter
	order      []string
	projection []string

	distinct bool
	keysOnly bool
	eventual bool
	limit    int32
	offset   int32
	start    string
	end      string

	cursor Cursor

	err     error
	queryDs *datastore.Query
}

func (q *Query) clone() *Query {
	x := *q
	// Copy the contents of the slice-typed fields to a new backing store.
	if len(q.filter) > 0 {
		x.filter = make([]filter, len(q.filter))
		copy(x.filter, q.filter)
	}
	if len(q.order) > 0 {
		x.order = make([]string, len(q.order))
		copy(x.order, q.order)
	}
	return &x
}

func (q *Query) Ancestor(ancestor *Key) *Query {
	q = q.clone()
	q.ancestor = ancestor
	return q
}

func (q *Query) Filter(filterStr string, value interface{}) *Query {
	q = q.clone()
	q.filter = append(q.filter, filter{Field: filterStr, Value: value})
	q.queryDs = q.queryDs.Filter(filterStr, value)
	return q
}

func (q *Query) Order(fieldName string) *Query {
	q = q.clone()
	fieldName = strings.TrimSpace(fieldName)
	q.order = append(q.order, fieldName)
	q.queryDs = q.queryDs.Order(fieldName)
	return q
}

func (q *Query) Project(fieldNames ...string) *Query {
	q = q.clone()
	q.projection = append([]string(nil), fieldNames...)
	q.queryDs = q.queryDs.Project(fieldNames...)
	return q
}

// Distinct returns a derivative query that yields de-duplicated entities with
// respect to the set of projected fields. It is only used for projection
// queries.
func (q *Query) Distinct() *Query {
	q = q.clone()
	q.distinct = true
	q.queryDs = q.queryDs.Distinct()
	return q
}

// KeysOnly returns a derivative query that yields only keys, not keys and
// entities. It cannot be used with projection queries.
func (q *Query) KeysOnly() *Query {
	q = q.clone()
	q.keysOnly = true
	q.queryDs = q.queryDs.KeysOnly()
	return q
}

// Limit returns a derivative query that has a limit on the number of results
// returned. A negative value means unlimited.
func (q *Query) Limit(limit int) *Query {
	q = q.clone()
	q.limit = int32(limit)
	q.queryDs = q.queryDs.Limit(limit)
	return q
}

// Offset returns a derivative query that has an offset of how many keys to
// skip over before returning results. A negative value is invalid.
func (q *Query) Offset(offset int) *Query {
	q = q.clone()
	q.offset = int32(offset)
	q.queryDs = q.queryDs.Offset(offset)
	return q
}

// Start returns a derivative query with the given start point.
func (q *Query) Start(c Cursor) *Query {
	q = q.clone()
	q.cursor = c

	q.queryDs = q.queryDs.Start(c.dsCursor)
	return q
}

// End returns a derivative query with the given end point.
func (q *Query) End(c Cursor) *Query {
	q = q.clone()
	q.cursor = c

	q.queryDs = q.queryDs.End(c.dsCursor)
	return q
}

func (q *Query) preRun(ctx context.Context) {
	if q.ancestor != nil {
		dsKey := convertKeyToDsKey(ctx, q.ancestor)
		q.queryDs = q.queryDs.Ancestor(dsKey)
	}

}

func (q *Query) Count(c context.Context) (int, error) {
	q.preRun(c)
	// intercept for mock
	return q.queryDs.Count(c)
}

func (q *Query) GetAll(c context.Context, dst interface{}) ([]*Key, error) {
	q.preRun(c)
	dskeys, err := q.queryDs.GetAll(c, dst)
	keys := convertDsKeysToKeys(c, dskeys)
	return keys, err
}

// Run runs the query in the given context.
func (q *Query) Run(ctx context.Context) *Iterator {
	q.preRun(ctx)

	if mock, ok := isMockQuery(ctx); ok {
		return mock.run(ctx, q)
	}

	it := q.queryDs.Run(ctx)
	return &Iterator{
		iter:   it,
		cursor: q.cursor,
		c:      ctx,
	}
}

type Cursor struct {
	ctx       context.Context
	cursorStr string
	dsCursor  datastore.Cursor
}

func (c *Cursor) String() string {
	return c.dsCursor.String()
}

func DecodeCursor(ctx context.Context, s string) (Cursor, error) {
	if mock, ok := isMockQuery(ctx); ok {
		return mock.decodeCursor(ctx, s)
	}
	cursor, err := datastore.DecodeCursor(s)
	return Cursor{
		ctx:       ctx,
		dsCursor:  cursor,
		cursorStr: s,
	}, err
}

// Iterator is the result of running a query.
type Iterator struct {
	c      context.Context
	cursor Cursor
	// iter is the original iterator which yielded this iterator.
	iter *datastore.Iterator

	err error

	// current row of iterator
	index int
}

func (i *Iterator) Next(dst interface{}) (*Key, error) {
	if i.err != nil {
		return nil, i.err
	}

	if mock, ok := isMockQuery(i.c); ok {
		return mock.next(i, dst)
	}

	k, err := i.iter.Next(dst)
	if err != nil {
		return nil, err
	}
	return convertDsKeyToKey(i.c, k), nil
}

func (i *Iterator) Cursor() (Cursor, error) {
	c, err := i.iter.Cursor()
	i.cursor.dsCursor = c
	return i.cursor, err
}
