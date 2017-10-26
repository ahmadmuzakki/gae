package datastore

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type Transaction struct {
	opts *TransactionOptions
}

func RunInTransaction(ctx context.Context, f func(tc context.Context) error, opts *TransactionOptions) error {
	ctx = context.WithValue(ctx, "transaction", Transaction{opts})
	o := datastore.TransactionOptions(*opts)
	return datastore.RunInTransaction(ctx, f, &o)
}

type TransactionOptions datastore.TransactionOptions
