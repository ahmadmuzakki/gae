package mock

import (
	"golang.org/x/net/context"
	"log"
)

type mock struct {
}

func InitMock() context.Context {
	m := &mock{}
	ctx := context.WithValue(context.Background(), "gae_mock", m)

	return ctx
}

func ValidateContext(ctx context.Context) {
	if _, ok := ctx.Value("gae_mock").(*mock); !ok {
		log.Fatal("Not Gae Mock Context")
	}
}

func GetMock(ctx context.Context) *mock {
	if m, ok := ctx.Value("gae_mock").(*mock); ok {
		return m
	}
	return nil
}
