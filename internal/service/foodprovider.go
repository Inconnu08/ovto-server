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
	ID       int64  `json:"id,omitempty"`
	Fullname string `json:"fullname"`
	//AvatarURL *string `json:"avatarURL"`
}

// FoodProvider profile model
type FoodProviderProfile struct {
	FoodProvider
	Email          string `json:"email,omitempty"`
	FollowersCount int    `json:"followersCount"`
	FolloweesCount int    `json:"followeesCount"`
	Me             bool   `json:"me"`
	Following      bool   `json:"following"`
	Followeed      bool   `json:"followeed"`
}

// CreateUser with the given email and name.
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

	query := "INSERT INTO foodprovider (email, fullname, phone, password) VALUES ($1, $2, $3, $4) RETURNING id"
	_, err = tx.ExecContext(ctx, query, email, fullname, phone, hPassword)
	unique := isUniqueViolation(err)
	if unique && strings.Contains(err.Error(), "email") {
		return ErrEmailTaken
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	if err != nil {
		return fmt.Errorf("could not create user: %v", err)
	}

	return nil
}

// UpdateDisplayPicture of the authenticated user returning the new avatar URL.
func (s *Service) UpdateFPDisplayPicture(ctx context.Context, r io.Reader) (string, error) {
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return "", ErrUnauthenticated
	}

	r = io.LimitReader(r, MaxAvatarBytes)
	img, format, err := image.Decode(r)
	if err == image.ErrFormat {
		return "", ErrUnsupportedDisplayPictureFormat
	}

	if err != nil {
		return "", fmt.Errorf("could not read dp: %v", err)
	}

	if format != "png" && format != "jpeg" {
		return "", ErrUnsupportedDisplayPictureFormat
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

	displayPicturePath := path.Join(dpDir, dp)
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
		defer os.Remove(path.Join(dpDir, oldDp.String))
	}
	dpURL := s.origin
	dpURL.Path = "/img/dp/" + dp

	return dpURL.String(), nil
}