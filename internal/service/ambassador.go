package service

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

// Create Ambassador with the given details.
func (s *Service) CreateAmbassador(ctx context.Context, email, fullname, phone, fb, city, area, address, password string) error {
	email = strings.TrimSpace(email)
	if !rxEmail.MatchString(email) {
		return ErrInvalidEmail
	}

	fullname = strings.TrimSpace(fullname)
	if !rxFullname.MatchString(fullname) {
		return ErrInvalidFullname
	}

	phone = strings.TrimSpace(phone)
	if !rxPhone.MatchString(phone) {
		return ErrInvalidPhone
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	hPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	RETRY:
	var retry int
	query := "INSERT INTO Ambassador (email, fullname, phone, fb, city, area, address, password, referral_code)" +
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
	_, err = tx.ExecContext(ctx, query, email, fullname, phone, fb, city, area, address, hPassword, GetRandomName(retry))
	unique := isUniqueViolation(err)
	if unique {
		fmt.Println("UNIQUE CONTRAINT:", err)
		if strings.Contains(err.Error(), "email") {
			return ErrEmailTaken
		}
		if strings.Contains(err.Error(), "phone") {
			return ErrPhoneNumberTaken
		}
		if strings.Contains(err.Error(), "referral_code") {
			retry += 1
			goto RETRY
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	if err != nil {
		return fmt.Errorf("could not create food provider: %v", err)
	}

	return nil
}

func (s *Service) UpdateAmbassador(ctx context.Context, fb, city, area, address, bkash, rocket string) error {
	uid, ok := ctx.Value(KeyAuthAmbassadorID).(int64)
	if !ok {
		return ErrUnauthenticated
	}

	fb = strings.TrimSpace(fb)
	city = strings.TrimSpace(city)
	area = strings.TrimSpace(area)
	address = strings.TrimSpace(address)
	bkash = strings.TrimSpace(bkash)
	rocket = strings.TrimSpace(rocket)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	if fb != "" {
		query := "UPDATE Ambassador SET fb = $1 WHERE id = $2"
		_, err = tx.ExecContext(ctx, query, fb, uid)
	}
	if city != "" {
		query := "UPDATE Ambassador SET city = $1 WHERE id = $2"
		_, err = tx.ExecContext(ctx, query, city, uid)
	}
	if area != "" {
		query := "UPDATE Ambassador SET area = $1 WHERE id = $2"
		_, err = tx.ExecContext(ctx, query, area, uid)
	}
	if address != "" {
		query := "UPDATE Ambassador SET address = $1 WHERE id = $2"
		_, err = tx.ExecContext(ctx, query, address, uid)
	}
	if bkash != "" {
		query := "UPDATE Ambassador SET bkash = $1 WHERE id = $2"
		_, err = tx.ExecContext(ctx, query, bkash, uid)
	}
	if rocket != "" {
		query := "UPDATE Ambassador SET rocket = $1 WHERE id = $2"
		_, err = tx.ExecContext(ctx, query, rocket, uid)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not update user: %v", err)
	}

	return nil
}
