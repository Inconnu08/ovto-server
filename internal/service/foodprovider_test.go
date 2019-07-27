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
		Email     string
		Fullname  string
		Phone 	  string
		Password  string
	}{
		// success condition
		{Label: "Test should create Food Provider successfully", Condition: "success", Want: nil, Email: "johndoe@gmail.com", Fullname: "John Snow", Phone: "01867584576", Password: "ilovegolang"},
		// error condition
		{Label: "Test should fail if phone number taken", Condition: "fail", Want: ErrPhoneNumberTaken, Email: "chesterb@gmail.com", Fullname: "John Snow", Phone: "01867584576", Password: "ilovegolang"},
		{Label: "Test should fail if email number taken", Condition: "fail", Want: ErrEmailTaken, Email: "johndoe@gmail.com", Fullname: "John Snow", Phone: "01860584576", Password: "ilovegolang"},
		{Label: "Test should fail for invalid email", Condition: "fail", Want: ErrInvalidEmail, Email: "johndoegmail.com", Fullname: "John Snow", Phone: "01867584576", Password: "ilovegolang"},
		{Label: "Test should fail for invalid phone number", Condition: "fail", Want: ErrInvalidPhone, Email: "johndoe@gmail.com", Fullname: "John Targaeryen", Phone: "018675845760", Password: "ilovegolang"},
		{Label: "Test should fail for invalid full name", Condition: "fail", Want: ErrInvalidFullname, Email: "johndoe@gmail.com", Fullname: "007John Snow", Phone: "01867584576", Password: "ilovegolang"},
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
			Got := s.CreateFoodProvider(ctx, test.Email, test.Fullname, test.Phone, test.Password)
			if test.Condition == "success" {
				if Got != nil {
					t.Error("Got:", Got, "| Want:", test.Want)
				}
			} else {
				if !cmp.Equal(Got.Error(), test.Want.Error()) {
					t.Error("Got:", Got, "| Want:", test.Want)
				}
			}
		})
	}
}
