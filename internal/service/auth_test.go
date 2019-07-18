package service

import (
	"context"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hako/branca"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
)

func TestService_UserLogin(t *testing.T) {
	var tt = []struct {
		Label       string
		Condition   string
		Want        error
		Email       string
		LookupEmail string
		Password    string
	}{
		// success Condition
		{Label: "Test should login user successfully", Condition: "success", Want: nil, Email: "johndoe@gmail.com", LookupEmail: "johndoe@gmail.com", Password: "Demogorgan"},
		// error Condition
		{Label: "Test should fail with user not found", Condition: "fail", Want: ErrUserNotFound, Email: "johndoe@gmail.com", LookupEmail: "Demogorgan@gmail.com", Password: "Demogorgan"},
		{Label: "Test should fail with invalid password", Condition: "fail", Want: ErrInvalidPassword, Email: "johndoe@gmail.com", LookupEmail: "johndoe@gmail.com", Password: "something"},
	}

	tearDown := SetupTest()
	defer tearDown()

	ctx := context.TODO()

	codec := branca.NewBranca("supersecretkeyyoushouldnotcommit")
	codec.SetTTL(uint32(TokenLifeSpan.Seconds()))

	c, err := pgx.ParseURI(pgURL.String())
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := stdlib.OpenDB(c)

	if err := ValidateSchema(db); err != nil {
		log.Fatalf("failed to validate schema: %v\n", err)
	}

	s := New(db, codec, url.URL{})

	for _, test := range tt {
		t.Run(test.Label, func(t *testing.T) {
			_ = s.CreateUser(ctx, test.Email, "Test User", test.Password)
			user, got := s.UserLogin(ctx, test.LookupEmail, test.Password)
			if test.Condition == "success" {
				if user.Token == "" {
					t.Error("Got:", got, "| Want:", test.Want)
				}
			} else {
				if !cmp.Equal(got.Error(), test.Want.Error()) {
					t.Error("Got:", got, "| Want:", test.Want)
				}
			}
		})
	}
}
