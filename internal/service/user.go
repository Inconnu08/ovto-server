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
	"log"
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
	dpDir      = path.Join("web", "static", "img", "dp")

	// ErrUserNotFound denotes a not found user.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidEmail denotes an invalid email address.
	ErrInvalidEmail = errors.New("invalid email")
	// ErrInvalidFullname denotes an invalid username.
	ErrInvalidFullname = errors.New("invalid fullname")
	// ErrEmailTaken denotes an email already taken.
	ErrEmailTaken = errors.New("email taken")
	// ErrForbiddenFollow denotes a forbidden follow. Like following yourself.
	ErrForbiddenFollow = errors.New("forbidden follow")
	// ErrUnsupportedDisplayPictureFormat denotes an unsupported avatar image format.
	ErrUnsupportedDisplayPictureFormat = errors.New("unsupported display picture format")
	// ErrInvalidPassword denotes an invalid password which could not be hashed.
	ErrInvalidPassword = errors.New("invalid password")
	// ErrInvalidEmail denotes an invalid phone number.
	ErrInvalidPhone = errors.New("invalid phone number")
)

// User model.
type User struct {
	ID       int64  `json:"id,omitempty"`
	Fullname string `json:"fullname"`
	//AvatarURL *string `json:"avatarURL"`
}

// UserProfile model.
type UserProfile struct {
	User
	Email          string `json:"email,omitempty"`
	FollowersCount int    `json:"followersCount"`
	FolloweesCount int    `json:"followeesCount"`
	Me             bool   `json:"me"`
	Following      bool   `json:"following"`
	Followeed      bool   `json:"followeed"`
}

// CreateUser with the given email and name.
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
	query := "INSERT INTO users (email, fullname) VALUES ($1, $2) RETURNING id"
	err = tx.QueryRowContext(ctx, query, email, fullname).Scan(&id)
	log.Print("Returned user id: ", id)
	unique := isUniqueViolation(err)
	if unique && strings.Contains(err.Error(), "email") {
		return ErrEmailTaken
	}

	hPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	log.Println(hPassword)
	log.Println(string(hPassword))
	log.Println([]byte(string(hPassword)))
	if err != nil {
		return err
	}

	query = "INSERT INTO credentials (user_id , password) VALUES ($1, $2)"
	if _, err = tx.ExecContext(ctx, query, id, hPassword); err != nil {
		return fmt.Errorf("failed to save password: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	if err != nil {
		return fmt.Errorf("could not create user: %v", err)
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
