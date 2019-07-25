package service

import (
	"context"
	"fmt"
	"strings"
)

func (s *Service) CreateRestaurant(ctx context.Context, title, about, phone, location, city, area, country, openingTime, closingTime, referral string) error {
	uid, ok := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !ok {
		return ErrUnauthenticated
	}

	title = strings.TrimSpace(title)
	about = strings.TrimSpace(about)
	city = strings.TrimSpace(city)
	area = strings.TrimSpace(area)
	location = strings.TrimSpace(location)
	openingTime = strings.TrimSpace(openingTime)
	closingTime = strings.TrimSpace(closingTime)
	referral = strings.TrimSpace(referral)
	if !rxPhone.MatchString(phone) {
		return ErrInvalidPhone
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	query, args, err := buildQuery(`
		INSERT INTO restaurant (title, owner_id, about, location, city, area, country, phone, opening_time, closing_time
		{{if .ambassador_code}}
		, ambassador_code
		{{end}})
		VALUES (@1, @2, @3, @4, @5, @6, @7, @8, @9, @10
		{{if .ambassador_code}}
		, @11
		{{end}})
  		RETURNING id`, map[string]interface{}{
		"1":               title,
		"2":               uid,
		"3":               about,
		"4":               location,
		"5":               city,
		"6":               area,
		"7":               country,
		"8":               phone,
		"9":               openingTime,
		"10":              closingTime,
		"11":              referral,
		"ambassador_code": referral,
	})

	//query = `
	//	INSERT INTO restaurant (title, owner_id, about, location, city, area, country, phone, opening_time, closing_time, ambassador_code)
	//	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	//	RETURNING id`
	_, err = tx.ExecContext(ctx, query, args...)
	unique := isUniqueViolation(err)
	if unique {
		return ErrTitleTaken
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	if err != nil {
		return fmt.Errorf("could not create restaurant: %v", err)
	}

	return nil
}
