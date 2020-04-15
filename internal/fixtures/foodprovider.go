package fixtures

import (
	"context"
	"fmt"
	"ovto/internal/service"
)

const (
	fullname string = "Taufiq Rahman"
	email    string = "johndoe@gmail.com"
	phone    string = "01767586098"
	password string = "coolpass"

	title    string = "White Canary"
	about    string = "We sale yummy pancakes!"
	location string = "House #50, Road #89"
	city     string = "Dhaka"
	area     string = "Gulshan"
	country  string = "Bangladesh"
	oTime    string = "8AM"
	cTime    string = "11PM"
	referral string = ""
)

func PopulateFoodProvider(s *service.Service) {
	ctx := context.TODO()

	//if err := s.CreateFoodProvider(ctx, email, fullname, phone, password); err != nil {
	//	fmt.Println(err)
	//	return
	//}

	fp, err := s.FoodProviderLogin(ctx, phone, password)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(&fp)

	err = s.CreateRestaurant(ctx, title, about, phone, location, city, area, country, oTime, cTime)
	if err != nil {
		fmt.Println(err)
		return
	}
}
