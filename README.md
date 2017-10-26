gae: A Golang AppEngine SDK wrapper
===================
*designed for testable Go Appengine SDK*

This package will give you support for mockable services in GAE. 

List of supported services:
- Datastore


### Install

`go get -u github.com/ahmadmuzakki/gae`

### Datastore

file *user.go*
```go
package main 

import (
	"fmt"
	"github.com/ahmadmuzakki/gae"
	"github.com/ahmadmuzakki/gae/datastore"
	"golang.org/x/net/context"
	"log"
	"net/http"
)

type User struct {
	Key     *datastore.Key
	Name    string
	Address string
}

func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := gae.NewContext(r)

	newkey, err := createUser(ctx)

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(newkey.String()))
}

func createUser(ctx context.Context) (*datastore.Key, error) {
	user := &User{
		Name:    "Jeki",
		Address: "Sidoarjo",
	}
	k := datastore.NewKey(ctx, "User", user.Name, 0, nil)
	return datastore.Put(ctx, k, user)
}
```

Let's mock it up! :D

file *user_test.go*
```go
package main

import (
	"fmt"
	"github.com/ahmadmuzakki/gae/datastore"
	gaemock "github.com/ahmadmuzakki/gae/mock"
	"github.com/stretchr/testify/assert"
	"testing"
	"errors"
)


func TestCreateUser(t *testing.T) {
	ctx := gaemock.InitMock()
	ctx, mock := datastore.Mock(ctx)

	testCases := []struct {
		mock        func()
		expectKey   *datastore.Key
		expectError error
	}{
		{
			mock: func() {
				user := User{
					Name:    "Jeki",
					Address: "Jakarta",
				}

				k := mock.MockKey(ctx, "User", user.Name, 0, nil)

				mock.MockPut(k, &user).WillReturnKeyErr(k, nil)
			},
			expectKey:   mock.MockKey(ctx, "User", "Jeki", 0, nil),
			expectError: nil,
		},
		{
			mock: func() {
				user := User{
					Name:    "Jeki",
					Address: "Jakarta",
				}
				k := mock.MockKey(ctx, "User", user.Name, 0, nil)

				mock.MockPut(k, &user).WillReturnKeyErr(nil, fmt.Errorf("failed to create user"))
			},
			expectKey:   nil,
			expectError: fmt.Errorf("failed to create user"),
		},
	}

	for _, test := range testCases {
		test.mock()
		newk, err := createUser(ctx)
		assert.Equal(t, test.expectError, err)
		assert.Equal(t, test.expectKey, newk)
	}

}

```
