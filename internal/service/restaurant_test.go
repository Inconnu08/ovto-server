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

func TestCreateRestaurant(t *testing.T) {
	var tt = []struct {
		Label     string
		Condition string
		Want      error
		Title     string
		About     string
		Phone     string
		Location  string
		City      string
		Area      string
		Country   string
		OTime     string
		CTime     string
		Referral  string
	}{
		// success condition
		{Label: "Test should create Restaurant successfully", Condition: "success", Want: nil, Title: "White Canary", About: "WE sale yummy pancakes!", Phone: "01967584756", Location: "House #50, Road #89", City: "Dhaka", Area: "Gulshan", Country: "Bangladesh", OTime: "8AM", CTime: "11PM", Referral: ""},
		// error condition
		{Label: "Test should fail if phone number taken", Condition: "fail", Want: ErrPhoneNumberTaken},
		{Label: "Test should fail if email number taken", Condition: "fail", Want: ErrEmailTaken},
		{Label: "Test should fail for invalid email", Condition: "fail", Want: ErrInvalidEmail},
		{Label: "Test should fail for invalid phone number", Condition: "fail", Want: ErrInvalidPhone},
		{Label: "Test should fail for invalid full name", Condition: "fail", Want: ErrInvalidFullname},
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

	_ = s.CreateFoodProvider(ctx, "johndoe@gmail.com", "John Snow", "01867584576", "ilovegolang")
	user, _ := s.FoodProviderLogin(ctx, "johndoe@gmail.com", "ilovegolang")
	ctx = context.WithValue(ctx, KeyAuthFoodProviderID, user.AuthUser.ID)

	for _, test := range tt {
		t.Run(test.Label, func(t *testing.T) {

			Got := s.CreateRestaurant(ctx, test.Title, test.About, test.Phone, test.Location, test.City, test.Area, test.Country, test.OTime, test.CTime, test.Referral)
			if test.Condition == "success" {
				if Got != nil {
					t.Error("Got:", Got, "| Want:", test.Want)
				}
			} else {
				if !cmp.Equal(err.Error(), test.Want.Error()) {
					t.Error("Got:", err, "| Want:", test.Want)
				}
			}
		})
	}
}
