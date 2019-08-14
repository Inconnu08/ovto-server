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
)

type Restaurant struct {
	Id     string  `json:"id"`
	Title  string  `json:"title"`
	About  string  `json:"about, omitempty"`
	Role   string  `json:"role, omitempty"`
	Rating float64 `json:"rating, omitempty"`
}

func (s *Service) CreateRestaurant(ctx context.Context, title, about, phone, location, city, area, country, openingTime, closingTime string) error {
	uid, ok := ctx.Value(KeyAuthFoodProviderID).(int64)
	println("Finally:", uid)
	if !ok {
		return ErrUnauthenticated
	}
	println(uid)

	title = strings.TrimSpace(title)
	about = strings.TrimSpace(about)
	city = strings.TrimSpace(city)
	area = strings.TrimSpace(area)
	location = strings.TrimSpace(location)
	openingTime = strings.TrimSpace(openingTime)
	closingTime = strings.TrimSpace(closingTime)
	if !rxPhone.MatchString(phone) {
		return ErrInvalidPhone
	}

	// for debug
	var n string
	var i int
	err := s.db.QueryRowContext(ctx, "SELECT id, fullname from foodprovider WHERE id = $1", uid).Scan(&i, &n)
	println("Owner", n, i)
	// debug end

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	//query, args, err := buildQuery(`
	//	INSERT INTO restaurant (title, owner_id, about, location, city, area, country, phone, opening_time, closing_time)
	//	VALUES (@1, @2, @3, @4, @5, @6, @7, @8, @9, @10)
	//	RETURNING id`, map[string]interface{}{
	//	"1":               title,
	//	"2":               uid,
	//	"3":               about,
	//	"4":               location,
	//	"5":               city,
	//	"6":               area,
	//	"7":               country,
	//	"8":               phone,
	//	"9":               openingTime,
	//	"10":              closingTime,
	//})
	//if err != nil {
	//	return fmt.Errorf("could not build sql query: %v", err)
	//}
	query := `
		INSERT INTO restaurant (title, owner_id, about, location, city, area, country, phone, opening_time, closing_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`
	var id string
	err = tx.QueryRowContext(ctx, query, title, uid, about, location, city, area, country, phone, openingTime, closingTime).Scan(&id)
	fmt.Println("[RESTAURANT]", err)
	unique := isUniqueViolation(err)
	if unique {
		return ErrTitleTaken
	}

	fk := isForeignKeyViolation(err)
	if fk {
		return ErrUserNotFound
	}

	query = `INSERT INTO permission (id, restaurant_id, restaurant, role) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, query, uid, id, title, Admin)
	if err != nil {
		return fmt.Errorf("[Permission] could not create restaurant: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("[Restaurant] could not commit transaction: %v", err)
	}

	if err != nil {
		return fmt.Errorf("could not create restaurant: %v", err)
	}
	fmt.Println("Restaurant:", title, "ID: ", id)
	return nil
}

func (s *Service) GetRestaurantByID(ctx context.Context, id string) (Restaurant, error) {
	_, ok := ctx.Value(KeyAuthUserID).(int64)
	var r Restaurant
	if !ok {
		return r, ErrUnauthenticated
	}

	if !rxUUID.MatchString(id) {
		return r, ErrInvalidRestaurantId
	}

	query := "SELECT title, about, rating FROM restaurant WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, id).Scan(&r.Title, &r.About, &r.Rating)
	if err == sql.ErrNoRows {
		return r, ErrRestaurantNotFound
	}

	if err != nil {
		return r, fmt.Errorf("could not query restaurant: %v", err)
	}

	return r, nil
}

func (s *Service) GetRestaurantsByFp(ctx context.Context) ([]Restaurant, error) {
	uid, ok := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !ok {
		return nil, ErrUnauthenticated
	}

	query := "SELECT id, title, about, rating FROM restaurant WHERE owner_id = $1"
	rows, err := s.db.QueryContext(ctx, query, uid)
	if err == sql.ErrNoRows {
		return nil, ErrRestaurantNotFound
	}

	defer rows.Close()
	uu := make([]Restaurant, 0, 1)
	for rows.Next() {
		var r Restaurant
		var rl int
		if err = rows.Scan(&r.Id, &r.Title, &rl); err != nil {
			fmt.Println(r)
			return nil, fmt.Errorf("could not get title: %v", err)
		}
		r.Role = getRole(rl)
		fmt.Println(r)
		uu = append(uu, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("could not iterate restaurants rows: %v", err)
	}

	return uu, nil
}

func (s *Service) CreateRestaurantByAmbassador(ctx context.Context, title, about, phone, location, city, area, country, openingTime, closingTime, referral string) error {
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
	fmt.Println(err)
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
		UPDATE restaurant SET 
		{{if .about}}
		about = @1 
		{{end}}
		{{if .city}}
		, city = @2
		{{end}}
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
		"1":        about,
		"2":        city,
		"3":        area,
		"4":        phone,
		"5":        location,
		"6":        id,
		"about":    about,
		"city":     city,
		"area":     area,
		"phone":    phone,
		"location": location,
		"id":       id,
	})
	if err != nil {
		return fmt.Errorf("could not build sql query: %v", err)
	}

	fmt.Println(query)
	fmt.Println(args)
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
func (s *Service) UpdateRestaurantDisplayPicture(ctx context.Context, r io.Reader, rid string) (string, error) {
	uid, ok := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !ok {
		return "", ErrUnauthenticated
	}

	if !rxUUID.MatchString(rid) {
		return "", ErrInvalidRestaurantId
	}

	if _, err := s.checkPermission(ctx, Manager, uid, rid); err != nil {
		fmt.Println("Permission Failed!")
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

	displayPicturePath := path.Join(restaurantDir, rid)
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
		RETURNING (SELECT avatar FROM restaurant WHERE id = $2) AS old_dp`, dp, rid).
		Scan(&oldDp); err != nil {
		defer os.Remove(displayPicturePath)
		return "", fmt.Errorf("could not update dp: %v", err)
	}

	if oldDp.Valid {
		defer os.Remove(path.Join(restaurantDir, rid, oldDp.String))
	}
	dpURL := s.origin
	dpURL.Path = "/img/restaurant/" + rid + "/" + dp

	return dpURL.String(), nil
}

// UpdatePicture is a utility function returning the new image URL.
func (s *Service) UpdatePicture(ctx context.Context, r io.Reader, rid, dir, query, urlPath string, uid int64, h, w int) (string, error) {
	if rid != "" {
		if !rxUUID.MatchString(rid) {
			return "", ErrInvalidRestaurantId
		}
	}

	r = io.LimitReader(r, MaxImageBytes)
	img, format, err := image.Decode(r)
	if err == image.ErrFormat {
		return "", ErrUnsupportedImageFormat
	}

	if err != nil {
		return "", fmt.Errorf("could not read image name: %v", err)
	}

	if format != "png" && format != "jpeg" {
		return "", ErrUnsupportedImageFormat
	}

	if _, err := s.checkPermission(ctx, Manager, uid, rid); err != nil {
		fmt.Println("Permission Failed!")
		return "", ErrUnauthenticated
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

	picturePath := path.Join(dir, rid)
	if _, err := os.Stat(picturePath); os.IsNotExist(err) {
		err = os.Mkdir(picturePath, os.ModeDir)
		return "", fmt.Errorf("failed to create path for image: %v", err)
	}
	picturePath = path.Join(picturePath, imageName)
	f, err := os.Create(picturePath)
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
		defer os.Remove(picturePath)
		return "", fmt.Errorf("could not update imageName: %v", err)
	}

	if oldImg.Valid {
		defer os.Remove(path.Join(dir, rid, oldImg.String))
	}
	imgURL := s.origin
	imgURL.Path = urlPath + imageName

	return imgURL.String(), nil
}

func (s *Service) UpdateRestaurantCoverPicture(ctx context.Context, r io.Reader, rid string) (string, error) {
	uid, ok := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !ok {
		return "", ErrUnauthenticated
	}

	if _, err := s.checkPermission(ctx, Manager, uid, rid); err != nil {
		fmt.Println("Permission Failed!")
		return "", ErrUnauthenticated
	}

	return s.UpdatePicture(ctx, r, rid, "cover", `
		UPDATE restaurant SET cover = $1 WHERE id = $2
		RETURNING (SELECT cover FROM restaurant WHERE id = $2) AS old_dp`,
		"/img/restaurant/"+rid+"/",
		uid, 300, 800)
}

func (s *Service) UpdateRestaurantGallery(ctx context.Context, r io.Reader, rid string) (string, error) {
	uid, ok := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !ok {
		return "", ErrUnauthenticated
	}

	if !rxUUID.MatchString(rid) {
		return "", ErrInvalidRestaurantId
	}

	if _, err := s.checkPermission(ctx, Manager, uid, rid); err != nil {
		fmt.Println("Permission Failed!")
		return "", ErrUnauthenticated
	}

	r = io.LimitReader(r, MaxImageBytes)
	img, format, err := image.Decode(r)
	if err == image.ErrFormat {
		return "", ErrUnsupportedImageFormat
	}

	if err != nil {
		return "", fmt.Errorf("could not read imageName: %v", err)
	}

	if format != "png" && format != "jpeg" {
		return "", ErrUnsupportedImageFormat
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

	displayPicturePath := path.Join(restaurantDir, rid, "gallery")
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

	img = imaging.Fill(img, 400, 400, imaging.Center, imaging.CatmullRom)
	if format == "png" {
		err = png.Encode(f, img)
	} else {
		err = jpeg.Encode(f, img, nil)
	}
	if err != nil {
		return "", fmt.Errorf("could not write imageName to disk: %v", err)
	}

	if _, err = s.db.QueryContext(ctx, `INSERT INTO restaurant_gallery(restaurant_id, image) VALUES ($1, $2)`, rid, imageName) ; err != nil {
		return "", fmt.Errorf("could not update imageName: %v", err)
	}

	dpURL := s.origin
	dpURL.Path = "/img/restaurant/" + rid + "/gallery/" + imageName

	return dpURL.String(), nil
}