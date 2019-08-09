package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type key string

const (
	TokenLifeSpan = time.Hour * 1
	// These are used in context.
	KeyAuthUserID         key = "auth_user_id"
	KeyAuthFoodProviderID key = "auth_food_provider_id"
	KeyAuthAmbassadorID   key = "auth_ambassador_id"
)

var (
	// ErrUnimplemented denotes a not implemented functionality.
	ErrUnimplemented = errors.New("unimplemented")
	// ErrUnauthenticated denotes no authenticated user in context.
	ErrUnauthenticated = errors.New("unauthenticated")
	// ErrInvalidRedirectURI denotes an invalid redirect uri.
	ErrInvalidRedirectURI = errors.New("invalid redirect uri")
	// ErrInvalidVerificationCode denotes an invalid verification code.
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	// ErrVerificationCodeNotFound denotes a not found verification code.
	ErrVerificationCodeNotFound = errors.New("verification code not found")
	// ErrEmptyValue denotes a input value cannot be empty.
	ErrEmptyValue = errors.New("value cannot be empty")
)

type LoginOutput struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	AuthUser  User      `json:"authUser"`
}

type FPLoginOutput struct {
	Token       string              `json:"token"`
	ExpiresAt   time.Time           `json:"expiresAt"`
	AuthUser    FoodProviderProfile `json:"authUser"`
	Restaurants *[]Restaurant        `json:"restaurants, omitempty"`
}

type GoogleAuthOutput struct {
	Name    string         `json:"name"`
	Email   string         `json:"Email"`
	Picture ProfilePicture `json:"picture"`
}

type FacebookAuthOutput struct {
	Id       string         `json:"id"`
	Name     string         `json:"name"`
	Email    string         `json:"Email"`
	Birthday string         `json:"birthday"`
	Picture  ProfilePicture `json:"picture"`
}

type ProfilePicture struct {
	Data PictureData `json:"data"`
}

type PictureData struct {
	Height       int    `json:"height"`
	IsSilhouette bool   `json:"is_silhouette"`
	Url          string `json:"url"`
	Width        int    `json:"width"`
}

type ThirdPartyProfile interface {
}

// AuthUserID is used to decode token and get the id
func (s *Service) AuthUserID(token string) (int64, error) {
	str, err := s.uCodec.DecodeToString(token)
	if err != nil {
		return 0, fmt.Errorf("could not decode token: %v", err)
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse auth user id from token: %v", err)
	}

	return i, nil
}

func (s *Service) AuthFpID(token string) (int64, error) {
	str, err := s.fpCodec.DecodeToString(token)
	if err != nil {
		return 0, fmt.Errorf("could not decode token: %v", err)
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse auth user id from token: %v", err)
	}

	return i, nil
}

func (s *Service) AuthAmbassadorID(token string) (int64, error) {
	str, err := s.aCodec.DecodeToString(token)
	if err != nil {
		return 0, fmt.Errorf("could not decode token: %v", err)
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse auth user id from token: %v", err)
	}

	return i, nil
}

// AuthUser is the current authenticated user.
func (s *Service) AuthUser(ctx context.Context) (User, error) {
	var u User
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return u, ErrUnauthenticated
	}

	//return s.userByID(ctx, uid)

	query := "SELECT Fullname FROM users WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, uid).Scan(&u.Fullname)
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not query selected auth user: %v", err)
	}

	u.ID = uid
	return u, nil
}

// AuthUser is the current authenticated user.
func (s *Service) AuthFp(ctx context.Context) (User, error) {
	var u User
	uid, ok := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !ok {
		return u, ErrUnauthenticated
	}

	query := "SELECT Fullname FROM foodprovider WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, uid).Scan(&u.Fullname)
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not query selected auth user: %v", err)
	}

	u.ID = uid
	return u, nil
}

// AuthFoodProvider is the current authenticated food provider.
func (s *Service) AuthFoodProvider(ctx context.Context) (User, error) {
	var u User
	uid, ok := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !ok {
		return u, ErrUnauthenticated
	}

	//return s.userByID(ctx, uid)

	query := "SELECT Fullname FROM foodprovider WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, uid).Scan(&u.Fullname)
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not query selected auth food provider: %v", err)
	}

	u.ID = uid
	return u, nil
}

func (s *Service) UserLogin(ctx context.Context, phone string, password string) (LoginOutput, error) {
	var out LoginOutput

	password = strings.TrimSpace(password)
	phone = strings.TrimSpace(phone)
	if !rxPhone.MatchString(phone) {
		return out, ErrInvalidPhone
	}

	//var avatar sql.NullString
	query := "SELECT id, Fullname FROM users WHERE phone = $1"
	err := s.db.QueryRowContext(ctx, query, phone).Scan(&out.AuthUser.ID, &out.AuthUser.Fullname)
	if err == sql.ErrNoRows {
		return out, ErrUserNotFound
	}

	var hPassword []byte
	query = "SELECT Password FROM credentials WHERE user_id = $1"
	err = s.db.QueryRowContext(ctx, query, out.AuthUser.ID).Scan(&hPassword)
	if err == sql.ErrNoRows {
		return out, ErrUserNotFound
	}

	if err != nil {
		return out, fmt.Errorf("could not query user %v\n", err)
	}

	if err = bcrypt.CompareHashAndPassword(hPassword, []byte(password)); err != nil {
		return out, ErrInvalidPassword
	}

	//out.AuthUser.AvatarURL = s.avatarURL(avatar)

	out.Token, err = s.uCodec.EncodeToString(strconv.FormatInt(out.AuthUser.ID, 10))
	if err != nil {
		return out, fmt.Errorf("could not generate token: %v", err)
	}

	out.ExpiresAt = time.Now().Add(TokenLifeSpan)

	return out, nil
}

func (s *Service) FoodProviderLogin(ctx context.Context, phone string, password string) (FPLoginOutput, error) {
	var out FPLoginOutput

	password = strings.TrimSpace(password)
	phone = strings.TrimSpace(phone)
	if !rxPhone.MatchString(phone) {
		return out, ErrInvalidPhone
	}

	//var avatar sql.NullString
	var hPassword []byte
	query := "SELECT id, Fullname, Email, Phone, Password FROM foodprovider WHERE phone = $1"
	err := s.db.QueryRowContext(ctx, query, phone).Scan(&out.AuthUser.ID, &out.AuthUser.Fullname, &out.AuthUser.Email, &out.AuthUser.Phone, &hPassword)
	if err == sql.ErrNoRows {
		return out, ErrUserNotFound
	}

	if err = bcrypt.CompareHashAndPassword(hPassword, []byte(password)); err != nil {
		return out, ErrInvalidPassword
	}

	// TODO
	query = "SELECT restaurant_id, role, restaurant FROM permission WHERE id = $1"
	rows, err := s.db.QueryContext(ctx, query, &out.AuthUser.ID)
	fmt.Println(err)
	if err != sql.ErrNoRows {
		defer rows.Close()
		restaurantList := make([]Restaurant, 0)
		for rows.Next() {
			var r Restaurant
			var rl int
			if err = rows.Scan(&r.Id, &rl, &r.Title); err != nil {
				fmt.Println(r)
				return out, fmt.Errorf("could not iterate properly: %v", err)
			}
			fmt.Println(r)
			r.Role = getRole(rl)
			restaurantList = append(restaurantList, r)
		}

		if err = rows.Err(); err != nil {
			return out, fmt.Errorf("could not iterate restaurants: %v", err)
		}

		out.Restaurants = &restaurantList
	}

	//out.AuthUser.AvatarURL = s.avatarURL(avatar)

	out.Token, err = s.fpCodec.EncodeToString(strconv.FormatInt(out.AuthUser.ID, 10))
	if err != nil {
		return out, fmt.Errorf("could not generate token: %v", err)
	}

	out.ExpiresAt = time.Now().Add(TokenLifeSpan)

	return out, nil
}

func (s *Service) AmbassadorLogin(ctx context.Context, phone string, password string) (LoginOutput, error) {
	var out LoginOutput

	password = strings.TrimSpace(password)
	phone = strings.TrimSpace(phone)
	if !rxPhone.MatchString(phone) {
		return out, ErrInvalidPhone
	}

	//var avatar sql.NullString
	var hPassword []byte
	query := "SELECT id, Fullname, Password FROM Ambassador WHERE phone = $1"
	err := s.db.QueryRowContext(ctx, query, phone).Scan(&out.AuthUser.ID, &out.AuthUser.Fullname, &hPassword)
	if err == sql.ErrNoRows {
		return out, ErrUserNotFound
	}

	if err = bcrypt.CompareHashAndPassword(hPassword, []byte(password)); err != nil {
		return out, ErrInvalidPassword
	}

	//out.AuthUser.AvatarURL = s.avatarURL(avatar)

	out.Token, err = s.aCodec.EncodeToString(strconv.FormatInt(out.AuthUser.ID, 10))
	if err != nil {
		return out, fmt.Errorf("could not generate token: %v", err)
	}

	out.ExpiresAt = time.Now().Add(TokenLifeSpan)

	return out, nil
}

func (s *Service) FacebookAuth(ctx context.Context, profile FacebookAuthOutput) (LoginOutput, error) {
	var out LoginOutput

	//var avatar sql.NullString
	query := "SELECT id, Fullname FROM users WHERE Email = $1"
	err := s.db.QueryRowContext(ctx, query, profile.Email).Scan(&out.AuthUser.ID, &out.AuthUser.Fullname)
	if err == sql.ErrNoRows {
		query := "INSERT INTO users (Email, Fullname) VALUES ($1, $2) RETURNING id"
		err = s.db.QueryRowContext(ctx, query, profile.Email, profile.Name).Scan(&out.AuthUser.ID)
		unique := isUniqueViolation(err)
		if unique && strings.Contains(err.Error(), "Email") {
			return out, ErrEmailTaken
		}

		if err != nil {
			return out, fmt.Errorf("could not create user: %v", err)
		}
		out.AuthUser.Fullname = profile.Name
	}

	out.Token, err = s.uCodec.EncodeToString(strconv.FormatInt(out.AuthUser.ID, 10))
	if err != nil {
		return out, fmt.Errorf("could not generate token: %v", err)
	}

	out.ExpiresAt = time.Now().Add(TokenLifeSpan)

	return out, nil
}

func (s *Service) GoogleAuth(ctx context.Context, profile GoogleAuthOutput) (LoginOutput, error) {
	var out LoginOutput

	//var avatar sql.NullString
	query := "SELECT id, Fullname FROM users WHERE Email = $1"
	err := s.db.QueryRowContext(ctx, query, profile.Email).Scan(&out.AuthUser.ID, &out.AuthUser.Fullname)
	if err == sql.ErrNoRows {
		query := "INSERT INTO users (Email, Fullname) VALUES ($1, $2) RETURNING id"
		err = s.db.QueryRowContext(ctx, query, profile.Email, profile.Name).Scan(&out.AuthUser.ID)
		unique := isUniqueViolation(err)
		if unique && strings.Contains(err.Error(), "Email") {
			return out, ErrEmailTaken
		}

		if err != nil {
			return out, fmt.Errorf("could not create user: %v", err)
		}
		out.AuthUser.Fullname = profile.Name
	}

	out.Token, err = s.uCodec.EncodeToString(strconv.FormatInt(out.AuthUser.ID, 10))
	if err != nil {
		return out, fmt.Errorf("could not generate token: %v", err)
	}

	out.ExpiresAt = time.Now().Add(TokenLifeSpan)

	return out, nil
}
