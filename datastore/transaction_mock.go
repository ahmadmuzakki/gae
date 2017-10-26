package datastore

import (
	"fmt"
	"golang.org/x/net/context"
	"reflect"
)

func MockRunInTransaction(ctx context.Context, opts *TransactionOptions) context.Context {
	return context.WithValue(ctx, "mock_transaction", Transaction{opts})
}

func shouldRunInTransaction(ctx context.Context) error {
	tx, ok := ctx.Value("transaction").(Transaction)

	txmock, okmock := ctx.Value("mock_transaction").(Transaction)

	if ok && !okmock {
		return fmt.Errorf("Not expecting operation runs in transaction")
	}

	if !ok && okmock {
		return fmt.Errorf("Expecting operation to be run in transaction but it's not.")
	}

	if !reflect.DeepEqual(tx, txmock) {
		return fmt.Errorf("Transaction Options %+v is not match with %+v", tx.opts, txmock.opts)
	}
	return nil
}
