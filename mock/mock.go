package mock

import (
	"github.com/ahmadmuzakki/gae/datastore"
	"golang.org/x/net/context"
	"log"
)

type mock struct{}

func InitMock() context.Context {
	m := &mock{}
	ctx := context.WithValue(context.Background(), "gae_mock", m)

	return ctx
}

func validateContext(ctx context.Context) {
	if _, ok := ctx.Value("gae_mock").(*mock); !ok {
		log.Fatal("Not Gae Mock Context")
	}
}

func DatastoreMock(ctx context.Context) (context.Context, *datastore.DatastoreMock) {
	validateContext(ctx)
	mock := &datastore.DatastoreMock{}
	ctx = context.WithValue(ctx, "datastore_mock", mock)
	return ctx, mock
}

func IsMock(ctx context.Context) (*mock, bool) {
	m, ok := ctx.Value("gae_mock").(*mock)
	return m, ok
}
