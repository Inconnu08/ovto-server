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

func TestCategory(t *testing.T) {
	type testdata struct {
		Label     string
		Condition string
		Want      error
		Context   context.Context
		Name      string
	}

	var tt = []testdata{
		// success condition
		{Label: "Test should create Category successfully", Condition: "success", Want: nil, Name: "Breakfast"},
		// error condition
		{Label: "Test should not create Category that already exists in Restaurant", Condition: "Fail", Want: ErrTitleTaken, Name: "Breakfast"},
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

	s := New(db, nil, codec, fpCodec, nil, url.URL{})

	_ = s.CreateFoodProvider(ctx, "johndoe@gmail.com", "John Snow", "01867584576", "ilovegolang")
	user, _ := s.FoodProviderLogin(ctx, "01867584576", "ilovegolang")
	ctx = context.WithValue(ctx, KeyAuthFoodProviderID, user.AuthUser.ID)
	_ = s.CreateRestaurant(ctx, "test.Title", "test.About", "01616534596", "test.Location", "test.City", "test.Area", "test.Country", "11AM", "10PM")
	user, _ = s.FoodProviderLogin(ctx, "01867584576", "ilovegolang")
	r := user.Restaurants
	rid := (*r)[0].Id

	for _, test := range tt {
		t.Run(test.Label, func(t *testing.T) {
			Got := s.CreateCategory(ctx, rid, test.Name, true)
			if test.Condition == "success" {
				if Got != nil {
					t.Error("Got:", Got, "| Want:", test.Want)
				}
				c, err := s.GetCategoriesByRestaurant(ctx, rid)
				if c == nil {
					t.Error("Got:", err, "| Want:", test.Want)
				}
			} else {
				if !cmp.Equal(Got.Error(), test.Want.Error()) {
					t.Error("Got:", Got, "| Want:", test.Want)
				}
			}
		})
	}
}

func TestMenu(t *testing.T) {
	type testdata struct {
		Label     string
		Condition string
		Want      error
		Context   context.Context
		Name      string
		Desc      string
		Price     float64
		Category  int
	}

	var tt = []testdata{
		// success condition
		{Label: "Test should create Menu successfully", Condition: "success", Want: nil, Name: "Cheesy Sandwich", Desc: "Very cheesy?!", Price: 150, Category: 2},
		{Label: "Test should create Menu successfully", Condition: "success", Want: nil, Name: "Chicken Sandwich", Desc: "Very cheesy?!", Price: 130, Category: 1},
		// error condition
		{Label: "Test should fail to create Menu item that already exists", Condition: "fail", Want: ErrItemAlreadyExists, Name: "Cheesy Sandwich", Desc: "Very cheesy?!", Price: 150, Category: 1},
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

	_ = s.CreateFoodProvider(ctx, "johndoe@gmail.com", "John Snow", "01867584576", "ilovegolang")
	user, _ := s.FoodProviderLogin(ctx, "01867584576", "ilovegolang")
	log.Println("AUTH:", user)
	ctx = context.WithValue(ctx, KeyAuthFoodProviderID, user.AuthUser.ID)
	_ = s.CreateRestaurant(ctx, "test.Title", "test.About", "01616534596", "test.Location", "test.City", "test.Area", "test.Country", "11AM", "10PM")
	user, _ = s.FoodProviderLogin(ctx, "01867584576", "ilovegolang")
	//ctx = context.WithValue(ctx, KeyAuthFoodProviderID, user.AuthUser.ID)
	r := user.Restaurants
	println("R: ", r)
	rid := (*r)[0].Id

	_ = s.CreateCategory(ctx, rid, "Breakfast", true)
	_ = s.CreateCategory(ctx, rid, "Dinner", true)

	cid, _ := s.GetCategoriesByRestaurant(ctx, rid)
	cid1 := cid[0].Id
	cid2 := cid[1].Id
	for _, test := range tt {
		t.Run(test.Label, func(t *testing.T) {
			if test.Category == 1 {
				cid1 = cid2
			}
			Got := s.CreateItem(ctx, rid, cid1, test.Name, test.Desc, test.Price, true)
			if test.Condition == "success" {
				if Got != nil {
					t.Error("Got:", Got, "| Want:", test.Want)
				}
				c, err := s.GetMenuForFp(ctx, rid)
				if c == nil {
					t.Error("Got:", err, "| Want:", test.Want)
				}
			} else {
				Got := s.CreateItem(ctx, rid, cid1, test.Name, test.Desc, test.Price, true)
				if !cmp.Equal(Got.Error(), test.Want.Error()) {
					t.Error("Got:", Got, "| Want:", test.Want)
				}
			}
		})
	}
}
