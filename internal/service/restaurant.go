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
	if err != nil {
		return fmt.Errorf("could not build sql query: %v", err)
	}
	// query = `
	//	INSERT INTO restaurant (title, owner_id, about, location, city, area, country, phone, opening_time, closing_time, ambassador_code)
	//	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	//	RETURNING id`
	var id string
	err = tx.QueryRowContext(ctx, query, args...).Scan(&id)
	fmt.Println("ID: ", id)
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

func (s *Service) UpdateRestaurant(ctx context.Context, id, about, phone, location, city, area string) error {
	_, auth := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !auth {
		return ErrUnauthenticated
	}

	if !rxUUID.MatchString(id) {
		return ErrInvalidRestaurantId
	}

	about = strings.TrimSpace(about)
	city = strings.TrimSpace(city)
	area = strings.TrimSpace(area)
	location = strings.TrimSpace(location)
	phone = strings.TrimSpace(phone)
	if phone != "" {
		if !rxPhone.MatchString(phone) {
			return ErrInvalidPhone
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	query, args, err := buildQuery(`
		UPDATE INTO restaurant SET 
		{{if .about}}
		about = @1 
		{{end}})
		{{if .city}}
		, city = @2
		{{end}})
		{{if .area}}
		, city = @3
		{{end}}
		{{if .phone}}
		, phone = @4
		{{end}}
		{{if .location}}
		, city = @5
		{{end}}
  		WHERE id = @6`, map[string]interface{}{
		"1": about,
		"2": city,
		"3": area,
		"4": phone,
		"5": location,
		"6": id,
	})
	if err != nil {
		return fmt.Errorf("could not build sql query: %v", err)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update restaurant: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to update restaurant: could not commit transaction: %v", err)
	}

	return nil
}

func (s *Service) UpdateRestaurantStatus(ctx context.Context, id string, closed bool) error {
	_, auth := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !auth {
		return ErrUnauthenticated
	}

	if !rxUUID.MatchString(id) {
		return ErrInvalidRestaurantId
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	query := "UPDATE INTO restaurant SET closed = $1 WHERE id = $2"
	_, err = tx.ExecContext(ctx, query, closed, id)
	if err != nil {
		return fmt.Errorf("failed to update restaurant: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to update restaurant: could not commit transaction: %v", err)
	}

	return nil
}

// UpdateRestaurantDisplayPicture of the authenticated restaurant returning the new avatar URL.
func (s *Service) UpdateRestaurantDisplayPicture(ctx context.Context, r io.Reader, id string) (string, error) {
	uid, ok := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !ok {
		return "", ErrUnauthenticated
	}

	if !rxUUID.MatchString(id) {
		return "", ErrInvalidRestaurantId
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

	displayPicturePath := path.Join(restaurantDir, id)
	if _, err := os.Stat(displayPicturePath); os.IsNotExist(err) {
		err = os.Mkdir(displayPicturePath, os.ModeDir)
		return "", fmt.Errorf("failed to create path for image: %v", err)
	}
	displayPicturePath = path.Join(displayPicturePath, dp)
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
		UPDATE restaurant SET avatar = $1 WHERE id = $2
		RETURNING (SELECT avatar FROM restaurant WHERE id = $2) AS old_dp`, dp, uid).
		Scan(&oldDp); err != nil {
		defer os.Remove(displayPicturePath)
		return "", fmt.Errorf("could not update dp: %v", err)
	}

	if oldDp.Valid {
		defer os.Remove(path.Join(restaurantDir, id, oldDp.String))
	}
	dpURL := s.origin
	dpURL.Path = "/img/restaurant/" + id + "/" + dp

	return dpURL.String(), nil
}

// UpdateRestaurantDisplayPicture of the authenticated restaurant returning the new avatar URL.
func (s *Service) UpdatePicture(ctx context.Context, r io.Reader, id, dir, query, urlPath string, uid int64, h, w int) (string, error) {
	if id != "" {
		if !rxUUID.MatchString(id) {
			return "", ErrInvalidRestaurantId
		}
	}

	r = io.LimitReader(r, MaxAvatarBytes)
	img, format, err := image.Decode(r)
	if err == image.ErrFormat {
		return "", ErrUnsupportedPictureFormat
	}

	if err != nil {
		return "", fmt.Errorf("could not read imageName: %v", err)
	}

	if format != "png" && format != "jpeg" {
		return "", ErrUnsupportedPictureFormat
	}

	imageName, err := gonanoid.Nanoid()
	if err != nil {
		return "", fmt.Errorf("could not generate imageName filename: %v", err)
	}

	if format == "png" {
		imageName += ".png"
	} else {
		imageName += ".jpg"
	}

	displayPicturePath := path.Join(dir, id)
	if _, err := os.Stat(displayPicturePath); os.IsNotExist(err) {
		err = os.Mkdir(displayPicturePath, os.ModeDir)
		return "", fmt.Errorf("failed to create path for image: %v", err)
	}
	displayPicturePath = path.Join(displayPicturePath, imageName)
	f, err := os.Create(displayPicturePath)
	if err != nil {
		return "", fmt.Errorf("could not create imageName file: %v", err)
	}

	defer f.Close()
	img = imaging.Fill(img, w, h, imaging.Center, imaging.CatmullRom)
	if format == "png" {
		err = png.Encode(f, img)
	} else {
		err = jpeg.Encode(f, img, nil)
	}
	if err != nil {
		return "", fmt.Errorf("could not write imageName to disk: %v", err)
	}

	var oldImg sql.NullString
	if err = s.db.QueryRowContext(ctx, query, imageName, uid).
		Scan(&oldImg); err != nil {
		defer os.Remove(displayPicturePath)
		return "", fmt.Errorf("could not update imageName: %v", err)
	}

	if oldImg.Valid {
		defer os.Remove(path.Join(dir, id, oldImg.String))
	}
	imgURL := s.origin
	imgURL.Path = urlPath + imageName

	return imgURL.String(), nil
}
