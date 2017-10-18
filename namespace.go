package gae

import (
	"github.com/ahmadmuzakki/gae/internal"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

func Namespace(ctx context.Context, ns string) (context.Context, error) {
	ctx = internal.WithNamespace(ctx, ns)
	return appengine.Namespace(ctx, ns)
}
