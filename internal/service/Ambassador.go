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

	query := "INSERT INTO Ambassador (email, fullname, phone, fb, city, area, address, password)" +
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	_, err = tx.ExecContext(ctx, query, email, fullname, phone, fb, city, area, address, hPassword)
	unique := isUniqueViolation(err)
	if unique {
		if strings.Contains(err.Error(), "Email") {
			return ErrEmailTaken
		} else {
			return ErrPhoneNumberTaken
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