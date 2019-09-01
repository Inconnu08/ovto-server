package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/disintegration/imaging"
	gonanoid "github.com/matoous/go-nanoid"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Ambassador struct {
	ID       int64  `json:"id,omitempty"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Facebook string `json:"facebook"`
	City     string `json:"city"`
	Area     string `json:"area"`
	Address  string `json:"address"`
	Password string `json:"password"`
}

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
	defer func() { _ = tx.Rollback() }()

	hPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

RETRY:
	var retry int
	query := `
		INSERT INTO Ambassador (email, fullname, phone, fb, city, area, address, password, referral_code)
 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
  		RETURNING id`
	_, err = tx.ExecContext(ctx, query, email, fullname, phone, fb, city, area, address, hPassword, GetRandomName(retry))
	unique := isUniqueViolation(err)
	if unique {
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
		return fmt.Errorf("could not create ambassador: %v", err)
	}

	return nil
}

// UpdateAmbassador updates ambassador's profile.
func (s *Service) UpdateAmbassador(ctx context.Context, fb, city, area, address string) error {
	uid, ok := ctx.Value(KeyAuthAmbassadorID).(int64)
	if !ok {
		return ErrUnauthenticated
	}

	fb = strings.TrimSpace(fb)
	city = strings.TrimSpace(city)
	area = strings.TrimSpace(area)
	address = strings.TrimSpace(address)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

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

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not update user: %v", err)
	}

	return nil
}

// GetAmbassadorById queries an Ambassador by their ID.
func (s *Service) GetAmbassadorById(ctx context.Context, id int64) (Ambassador, error) {
	var u Ambassador
	//var avatar sql.NullString
	query := "SELECT fullname, phone, fb, city, area, address FROM Ambassador WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, id).Scan()
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not query selected user: %v", err)
	}

	u.ID = id

	return u, nil
}

// GetAmbassadorByName queries an Ambassador by their name.
func (s *Service) GetAmbassadorByName(ctx context.Context, name string) (Ambassador, error) {
	var u Ambassador
	//var avatar sql.NullString
	query := "SELECT phone, fb, city, area, address FROM Ambassador WHERE fullname = $1"
	err := s.db.QueryRowContext(ctx, query, name).Scan()
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not query selected user: %v", err)
	}

	u.Fullname = name

	return u, nil
}

// AddPaymentMethod is a generic function for adding any sort of payment method for example bKash, rocket etc.
func (s *Service) AddPaymentMethod(ctx context.Context, password, method, number, remove string) error {
	uid, ok := ctx.Value(KeyAuthAmbassadorID).(int64)
	if !ok {
		return ErrUnauthenticated
	}

	number = strings.TrimSpace(number)
	if number == "" {
		return ErrEmptyValue
	}

	var hPassword []byte
	query := "SELECT Password FROM Ambassador WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, uid).Scan(&hPassword)
	if err == sql.ErrNoRows {
		return ErrUserNotFound
	}

	if err = bcrypt.CompareHashAndPassword(hPassword, []byte(password)); err != nil {
		return ErrInvalidPassword
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	query = `UPDATE Ambassador SET $1 = $2, $3 = NULL WHERE id = $4`
	_, err = tx.ExecContext(ctx, query, method, number, remove, uid)

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not update user: %v", err)
	}

	return nil
}

// Adds Bkash number for receiving payments.
func (s *Service) AddBKashNumber(ctx context.Context, password, bKash string) error {
	return s.AddPaymentMethod(ctx, password, "bkash", bKash, "rocket")
}

// Adds Rocket number for receiving payments
func (s *Service) AddRocketNumber(ctx context.Context, password, rocket string) error {
	return s.AddPaymentMethod(ctx, password, "rocket", rocket, "bkash")
}

// ChangeAmbassadorPassword updates the ambassador's password.
func (s *Service) ChangeAmbassadorPassword(ctx context.Context, oldPassword, newPassword string) error {
	uid, ok := ctx.Value(KeyAuthAmbassadorID).(int64)
	if !ok {
		return ErrUnauthenticated
	}

	oldPassword = strings.TrimSpace(oldPassword)
	newPassword = strings.TrimSpace(newPassword)

	var hPassword []byte
	query := "SELECT Password FROM Ambassador WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, uid).Scan(&hPassword)
	if err == sql.ErrNoRows {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword(hPassword, []byte(oldPassword)); err != nil {
		return ErrInvalidPassword
	}

	hPassword, err = bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query = "UPDATE credentials password = $2 WHERE id = $1"
	_, err = s.db.ExecContext(ctx, query, uid, hPassword)
	if err != nil {
		return fmt.Errorf("could not update password: %v", err)
	}

	return nil
}

// UpdateAmbassadorDisplayPicture of the authenticated user returning the new avatar URL.
func (s *Service) UpdateAmbassadorDisplayPicture(ctx context.Context, r io.Reader) (string, error) {
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return "", ErrUnauthenticated
	}

	r = io.LimitReader(r, MaxImageBytes)
	img, format, err := image.Decode(r)
	if err == image.ErrFormat {
		return "", ErrUnsupportedImageFormat
	}

	if err != nil {
		return "", fmt.Errorf("could not read dp: %v", err)
	}

	if format != "png" && format != "jpeg" {
		return "", ErrUnsupportedImageFormat
	}

	dp, err := gonanoid.Nanoid()
	if err != nil {
		return "", fmt.Errorf("could not generate dp filename: %v", err)
	}

	if format == "png" {
		dp += ".png"
	} else {
		dp += ".jpg"
	}

	displayPicturePath := path.Join(userDpDir, dp)
	f, err := os.Create(displayPicturePath)
	if err != nil {
		return "", fmt.Errorf("could not create dp file: %v", err)
	}

	defer f.Close()
	img = imaging.Fill(img, 400, 400, imaging.Center, imaging.CatmullRom)
	if format == "png" {
		err = png.Encode(f, img)
	} else {
		err = jpeg.Encode(f, img, nil)
	}
	if err != nil {
		return "", fmt.Errorf("could not write dp to disk: %v", err)
	}

	var oldDp sql.NullString
	if err = s.db.QueryRowContext(ctx, `
		UPDATE users SET avatar = $1 WHERE id = $2
		RETURNING (SELECT avatar FROM users WHERE id = $2) AS old_dp`, dp, uid).
		Scan(&oldDp); err != nil {
		defer os.Remove(displayPicturePath)
		return "", fmt.Errorf("could not update dp: %v", err)
	}

	if oldDp.Valid {
		defer os.Remove(path.Join(userDpDir, oldDp.String))
	}
	dpURL := s.origin
	dpURL.Path = "/img/ambassador/dp/" + dp

	return dpURL.String(), nil
}
