package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/disintegration/imaging"
	gonanoid "github.com/matoous/go-nanoid"
	"golang.org/x/crypto/bcrypt"
)

// MaxAvatarBytes to read.
const MaxAvatarBytes = 5 << 20 // 5MB

var (
	rxEmail    = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	rxFullname = regexp.MustCompile(`^[a-zA-Z ]{0,20}$`)
	rxPhone    = regexp.MustCompile(`(^([+]{1}[8]{2}|0088)?(01){1}[5-9]{1}\d{8})$`)
	rxUUID     = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")

	userDpDir     = path.Join("web", "static", "img", "user", "dp")
	restaurantDir = path.Join("web", "static", "img", "restaurant")

	// ErrUserNotFound denotes a not found user.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidEmail denotes an invalid Email address.
	ErrInvalidEmail = errors.New("invalid Email")
	// ErrInvalidFullname denotes an invalid username.
	ErrInvalidFullname = errors.New("invalid Fullname")
	// ErrEmailTaken denotes an Email already taken.
	ErrEmailTaken = errors.New("email taken")
	// ErrTitleTaken denotes a restaurant with that title already exists.
	ErrTitleTaken = errors.New("title taken")
	// ErrUnauthenticated denotes no authenticated user in context.
	ErrInvalidRestaurantId = errors.New("invalid restaurant ID")
	// ErrForbiddenFollow denotes a forbidden follow. Like following yourself.
	ErrForbiddenFollow = errors.New("forbidden follow")
	// ErrUnsupportedPictureFormat denotes an unsupported avatar image format.
	ErrUnsupportedPictureFormat = errors.New("unsupported picture format")
	// ErrInvalidPassword denotes an invalid Password which could not be hashed.
	ErrInvalidPassword = errors.New("invalid Password")
	// ErrInvalidEmail denotes an invalid phone number.
	ErrInvalidPhone = errors.New("invalid phone number")
	// ErrPhoneNumberTaken denotes the phone number provided is taken
	ErrPhoneNumberTaken = errors.New("phone number taken")
)

// User model.
type User struct {
	ID       int64  `json:"id,omitempty"`
	Fullname string `json:"Fullname"`
	//AvatarURL *string `json:"avatarURL"`
}

// UserProfile model.
type UserProfile struct {
	User
	Email          string `json:"Email,omitempty"`
	FollowersCount int    `json:"followersCount"`
	FolloweesCount int    `json:"followeesCount"`
	Me             bool   `json:"me"`
	Following      bool   `json:"following"`
	Followeed      bool   `json:"followeed"`
}

// CreateUser with the given Email and name.
func (s *Service) CreateUser(ctx context.Context, email, fullname, password string) error {
	email = strings.TrimSpace(email)
	if !rxEmail.MatchString(email) {
		return ErrInvalidEmail
	}

	fullname = strings.TrimSpace(fullname)
	if !rxFullname.MatchString(fullname) {
		return ErrInvalidFullname
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	var id int
	query := "INSERT INTO users (Email, Fullname) VALUES ($1, $2) RETURNING id"
	err = tx.QueryRowContext(ctx, query, email, fullname).Scan(&id)
	unique := isUniqueViolation(err)
	if !unique && err != nil {
		return err
	}
	if unique {
		return ErrEmailTaken
	}

	hPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query = "INSERT INTO credentials (user_id , Password) VALUES ($1, $2)"
	if _, err = tx.ExecContext(ctx, query, id, hPassword); err != nil {
		return fmt.Errorf("failed to save Password: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	if err != nil {
		return fmt.Errorf("could not create user: %v", err)
	}

	return nil
}

func (s *Service) UpdateUser(ctx context.Context, address, phone string) error {
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return ErrUnauthenticated
	}

	address = strings.TrimSpace(address)
	phone = strings.TrimSpace(phone)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	if address != "" {
		query := "UPDATE users SET address = $1 WHERE id = $2"
		_, err = tx.ExecContext(ctx, query, address, uid)
	}

	if phone != "" {
		if !rxPhone.MatchString(phone) {
			return ErrInvalidPhone
		}
		query := "UPDATE users SET phone = $1 WHERE id = $2"
		_, err = tx.ExecContext(ctx, query, phone, uid)
	}

	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	if err != nil {
		return fmt.Errorf("could not update user: %v", err)
	}

	return nil
}

func (s *Service) DeleteUser(ctx context.Context) error {
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return ErrUnauthenticated
	}

	query := "DELETE users WHERE id = $1"
	_, err := s.db.ExecContext(ctx, query, uid)
	if err != nil {
		return fmt.Errorf("could not delete user: %v", err)
	}

	return nil
}

// UpdateDisplayPicture of the authenticated user returning the new avatar URL.
func (s *Service) UpdateDisplayPicture(ctx context.Context, r io.Reader) (string, error) {
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return "", ErrUnauthenticated
	}

	r = io.LimitReader(r, MaxAvatarBytes)
	img, format, err := image.Decode(r)
	if err == image.ErrFormat {
		return "", ErrUnsupportedPictureFormat
	}

	if err != nil {
		return "", fmt.Errorf("could not read dp: %v", err)
	}

	if format != "png" && format != "jpeg" {
		return "", ErrUnsupportedPictureFormat
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
	dpURL.Path = "/img/user/dp/" + dp

	return dpURL.String(), nil
}

func (s *Service) ChangeUserPassword(ctx context.Context, oldPassword, newPassword string) error {
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return ErrUnauthenticated
	}

	oldPassword = strings.TrimSpace(oldPassword)
	newPassword = strings.TrimSpace(newPassword)

	var hPassword []byte
	query := "SELECT Password FROM users WHERE id = $1"
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

	query = "UPDATE credentials password = $2 WHERE user_id = $1"
	_, err = s.db.ExecContext(ctx, query, uid, hPassword)
	if err != nil {
		return fmt.Errorf("could not update password: %v", err)
	}

	return nil
}
