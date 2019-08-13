package service

import (
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"strings"

	"github.com/disintegration/imaging"
	gonanoid "github.com/matoous/go-nanoid"
	"golang.org/x/crypto/bcrypt"
)

// FoodProvider model
type FoodProvider struct {
	ID       int64   `json:"id,omitempty"`
	Fullname string  `json:"fullname"`
	Avatar   *string `json:"avatar,omitempty"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
}

// FoodProvider profile model
type FoodProviderProfile struct {
	FoodProvider
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
	Me    bool   `json:"me,omitempty"`
}

// CreateUser with the given Email and name.
func (s *Service) CreateFoodProvider(ctx context.Context, email, fullname, phone, password string) error {
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

	var id int
	query := "INSERT INTO foodprovider (Email, Fullname, phone, Password) VALUES ($1, $2, $3, $4) RETURNING id"
	err = tx.QueryRowContext(ctx, query, email, fullname, phone, hPassword).Scan(&id)
	fmt.Println("UNIQUE ID: ", id)
	unique := isUniqueViolation(err)
	if unique {
		if strings.Contains(err.Error(), "Email") || strings.Contains(err.Error(), "_email") {
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
func (s *Service) CreateRole(ctx context.Context, rid, fullname, phone, password, role string) error {
	uid, ok := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !ok {
		return ErrUnauthenticated
	}

	fullname = strings.TrimSpace(fullname)
	if !rxFullname.MatchString(fullname) {
		return ErrInvalidFullname
	}

	phone = strings.TrimSpace(phone)
	if !rxPhone.MatchString(phone) {
		return ErrInvalidPhone
	}

	roleLevel, err := getRoleLevel(role)
	if err != nil {
		return ErrUnauthenticated
	}

	restaurant, err := s.checkPermission(ctx, Admin, uid, rid)
	if err != nil {
		fmt.Println("Permission Failed!")
		return ErrUnauthenticated
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

	var id int
	query := "INSERT INTO foodprovider (fullname, phone, password) VALUES ($1, $2, $3) RETURNING id"
	err = tx.QueryRowContext(ctx, query, fullname, phone, hPassword).Scan(&id)
	unique := isUniqueViolation(err)
	if unique {
		if strings.Contains(err.Error(), "Email") || strings.Contains(err.Error(), "_email") {
			return ErrEmailTaken
		} else {
			return ErrPhoneNumberTaken
		}
	}

	query = `INSERT INTO permission (id, restaurant_id, restaurant, role) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, query, id, rid, restaurant, roleLevel)
	if err != nil {
		return fmt.Errorf("[Permission] could not create restaurant: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	if err != nil {
		return fmt.Errorf("could not create food provider: %v", err)
	}

	return nil
}

// UpdateDisplayPicture of the authenticated user returning the new avatar URL.
func (s *Service) UpdateFPDisplayPicture(ctx context.Context, r io.Reader) (string, error) {
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
	dpURL.Path = "/img/dp/" + dp

	return dpURL.String(), nil
}

func (s *Service) ChangeFoodProviderPassword(ctx context.Context, oldPassword, newPassword string) error {
	uid, ok := ctx.Value(KeyAuthAmbassadorID).(int64)
	if !ok {
		return ErrUnauthenticated
	}

	oldPassword = strings.TrimSpace(oldPassword)
	newPassword = strings.TrimSpace(newPassword)

	var hPassword []byte
	query := "SELECT Password FROM foodprovider WHERE id = $1"
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

	query = "UPDATE foodprovider password = $2 WHERE id = $1"
	_, err = s.db.ExecContext(ctx, query, uid, hPassword)
	if err != nil {
		return fmt.Errorf("could not update password: %v", err)
	}

	return nil
}
