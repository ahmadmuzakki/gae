package gae

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"net/http"
)

func NewContext(r *http.Request) context.Context {

	return appengine.NewContext(r)
}
