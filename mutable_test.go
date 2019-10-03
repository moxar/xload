package xload_test

import (
	"context"
	"fmt"
	"time"

	"github.com/moxar/xload"
)

func ExampleMutable() {

	// auto incremented id, starting at 5
	var serial = 5

	type User struct {
		ID   int
		Name string
	}

	// saveUser does not return the created users. Instead, it sets the users id with the current serial value.
	saveUsers := func(ctx context.Context, us ...*User) error {
		for _, u := range us {
			u.ID = serial
			serial++
		}
		// store user
		return nil
	}

	// the operation returns no input, because the underlying call (saveUsers) does not.
	// Instead, it wraps the users into mutable so they satisfy the Fragment interface.
	operation := func(ctx context.Context, fs ...xload.Fragment) (interface{}, error) {
		var users []*User
		for _, f := range fs {
			users = append(users, f.(xload.Mutable).Value.([]*User)...)
		}

		err := saveUsers(ctx, users...)
		return nil, err
	}

	// This is the collection of users to save.
	var in = []*User{
		{Name: "Batman"},
		{Name: "Superman"},
	}

	// Declare the buffer, and use it.
	buffer := xload.NewBuffer(context.TODO(), operation, 100, time.Millisecond*20)
	if _, err := buffer.Do(xload.NewMutable(in)); err != nil {
		panic(err)
	}

	for _, u := range in {
		fmt.Println(u.ID, u.Name)
	}
	// output:
	// 5 Batman
	// 6 Superman
}
