package service

import (
	"context"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hako/branca"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	_ "github.com/jackc/pgx/stdlib"
)

func TestService_CreateUser(t *testing.T) {
	var tt = []struct {
		Label     string
		Condition string
		Want      error
		Email     string
		Fullname  string
		Password  string
	}{
		// success Condition
		{Label: "Test should create user successfully", Condition: "success", Want: nil, Email: "johndoe@gmail.com", Fullname: "John Snow", Password: "ShesMyQueen"},
		// error Condition
		{Label: "Test should fail if email taken", Condition: "fail", Want: ErrEmailTaken, Email: "johndoe@gmail.com", Fullname: "John Snow", Password: "ShesMyQueen"},
		{Label: "Test should fail for invalid email", Condition: "fail", Want: ErrInvalidEmail, Email: "johndoegmail.com", Fullname: "John Snow", Password: "ShesMyQueen"},
		{Label: "Test should fail for invalid full name", Condition: "fail", Want: ErrInvalidFullname, Email: "johndoe@gmail.com", Fullname: "007John Snow", Password: "ShesMyQueen"},
	}

	tearDown := SetupTest()
	defer tearDown()

	ctx := context.TODO()

	codec := branca.NewBranca("supersecretkeyyoushouldnotcommit")
	codec.SetTTL(uint32(TokenLifeSpan.Seconds()))

	fpCodec := branca.NewBranca("supersecretkeyyoushouldcommitnot")
	fpCodec.SetTTL(uint32(TokenLifeSpan.Seconds()))

	c, err := pgx.ParseURI(pgURL.String())
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := stdlib.OpenDB(c)

	if err := ValidateSchema(db); err != nil {
		log.Fatalf("failed to validate schema: %v\n", err)
	}

	s := New(db, codec, fpCodec, nil, url.URL{})

	for _, test := range tt {
		t.Run(test.Label, func(t *testing.T) {
			Got := s.CreateUser(ctx, test.Email, test.Fullname, test.Password)
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

func TestService_UpdateUser(t *testing.T) {
	var tt = []struct {
		Label     string
		Condition string
		Want      error
		Email     string
		Fullname  string
		Password  string
		Phone     string
		Address   string
	}{
		// success Condition
		{Label: "Test should update user successfully", Condition: "success", Want: nil, Email: "johndoe@gmail.com", Fullname: "John Snow", Password: "ShesMyQueen", Phone: "01748596758", Address: "Intergalactic, House #1234, Road #456, Cybertron"},
		{Label: "Test should update user with just phone successfully", Condition: "success", Want: nil, Email: "johndoe@gmail.com", Fullname: "John Snow", Password: "ShesMyQueen", Phone: "01748596758"},
		{Label: "Test should update user with just address successfully", Condition: "success", Want: nil, Email: "johndoe@gmail.com", Fullname: "John Snow", Password: "ShesMyQueen", Address: "House name, House #124, Road #456, Mexico"},
		// error Condition
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

	s := New(db, codec, nil,nil, url.URL{})

	for _, test := range tt {
		t.Run(test.Label, func(t *testing.T) {
			_ = s.CreateUser(ctx, test.Email, test.Fullname, test.Password)
			user, _ := s.UserLogin(ctx, test.Email, test.Password)
			ctx = context.WithValue(ctx, KeyAuthUserID, user.AuthUser.ID)
			Got := s.UpdateUser(ctx, test.Address, test.Phone)
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
