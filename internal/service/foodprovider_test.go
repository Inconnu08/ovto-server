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

func TestCreateFoodProvider(t *testing.T) {
	var tt = []struct {
		Label     string
		Condition string
		Want      error
		Fullname  string
		Email     string
		Phone 	  string
		Password  string
	}{
		// success condition
		{Label: "Test should login Food Provider successfully", Condition: "success", Want: nil, Fullname: "Taufiq Rahman", Email: "johndoe@gmail.com", Phone:"01767586798", Password: "coolpass"},
		// error condition
		{Label: "Test should login should fail to create with invalid phone", Condition: "fail", Want: ErrInvalidPhone, Fullname: "Taufiq Rahman", Email: "johndoe@gmail.com", Phone:"01767586798c",  Password: "ilovegolang"},
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
			err = s.CreateFoodProvider(ctx, test.Email, test.Fullname, test.Phone, test.Password)
			if test.Condition == "success" {
				if err != nil {
					t.Error("Got:", err, "| Want:", test.Want)
				}
			} else {
				if !cmp.Equal(err.Error(), test.Want.Error()) {
					t.Error("Got:", err, "| Want:", test.Want)
				}
			}
		})
	}
}
