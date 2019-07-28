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
	type testdata struct {
		Label     string
		Condition string
		Want      error
		Context   context.Context
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
	}

	var tt = []testdata {
		// success condition
		{Label: "Test should create Restaurant successfully", Condition: "success", Want: nil, Title: "White Canary", About: "WE sale yummy pancakes!", Phone: "01967584756", Location: "House #50, Road #89", City: "Dhaka", Area: "Gulshan", Country: "Bangladesh", OTime: "8AM", CTime: "11PM", Referral: ""},
		// error condition
		{Label: "Test should fail with title taken", Condition: "fail", Want: ErrTitleTaken, Title: "White Canary", About: "WE sale yummy pancakes!", Phone: "01967584756", Location: "House #50, Road #89", City: "Dhaka", Area: "Gulshan", Country: "Bangladesh", OTime: "8AM", CTime: "11PM", Referral: ""},
		{Label: "Test should fail with invalid phone number", Condition: "fail", Want: ErrInvalidPhone, Title: "Taste Bud", About: "WE sale yummy desserts!", Phone: "019675847560", Location: "House #50, Road #89", City: "Dhaka", Area: "Gulshan", Country: "Bangladesh", OTime: "8AM", CTime: "11PM", Referral: ""},
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

	_ = s.CreateUser(ctx, "johndoe@gmail.com", "John Snow", "ilovegolang")
	//user2, _ := s.UserLogin(ctx, "johndoe@gmail.com", "ilovegolang")
	//ctx2 := context.WithValue(ctx, KeyAuthFoodProviderID, user2.AuthUser.ID)
	//tt = append(tt, testdata{Label: "Test should fail to create a restaurant by a non food provider account", Condition: "fail", Want: ErrUnauthenticated, Context:ctx2, Title: "White Canary", About: "WE sale yummy pancakes!", Phone: "01967584756", Location: "House #50, Road #89", City: "Dhaka", Area: "Gulshan", Country: "Bangladesh", OTime: "8AM", CTime: "11PM", Referral: ""})
	//_ = s.CreateFoodProvider(ctx, "evil@gmail.com", "John Snow Evil", "01807584576", "ilovebeingevil")
	//user2, _ := s.FoodProviderLogin(ctx, "johndoe@gmail.com", "ilovebeingevil")
	//ctx2 := context.WithValue(ctx, KeyAuthFoodProviderID, user2.AuthUser.ID)
	//append(tt, )
	for _, test := range tt {
		t.Run(test.Label, func(t *testing.T) {
			Got := s.CreateRestaurant(ctx, test.Title, test.About, test.Phone, test.Location, test.City, test.Area, test.Country, test.OTime, test.CTime, test.Referral)
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
