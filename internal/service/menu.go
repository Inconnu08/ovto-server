package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Item struct {
	Id                string  `json:"id"`
	Name              string  `json:"name"`
	Category          string  `json:"category"`
	CategoryAvailable string  `json:"category_available"`
	Description       string  `json:"description"`
	Price             float64 `json:"price"`
	Availability      bool    `json:"availability"`
}

type Category struct {
	Label string `json:"label"`
	Id    int64  `json:"id"`
}

func (s *Service) CreateCategory(ctx context.Context, rid, name string, availability bool) error {
	uid, auth := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !auth {
		return ErrUnauthenticated
	}

	if !rxUUID.MatchString(rid) {
		return ErrRestaurantNotFound
	}

	name = strings.TrimSpace(name)

	if _, err := s.checkPermission(ctx, Manager, uid, rid); err != nil {
		fmt.Println("[Permission Failed]:", err)
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	query := "INSERT INTO category(restaurant, label, availability) VALUES ($1, $2, $3)"
	_, err = tx.ExecContext(ctx, query, rid, name, availability)
	u := isUniqueViolation(err)
	if u {
		return ErrTitleTaken
	}

	if err != nil {
		return fmt.Errorf("failed to update restaurant: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to update restaurant: could not commit transaction: %v", err)
	}

	return nil
}

func (s *Service) GetCategoriesByRestaurant(ctx context.Context, rid string) ([]Category, error) {
	c := make([]Category, 0, 1)

	_, auth := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !auth {
		return c, ErrUnauthenticated
	}

	if !rxUUID.MatchString(rid) {
		return c, ErrRestaurantNotFound
	}

	query := `
		SELECT label, id
 		FROM category
		WHERE restaurant = $1`
	rows, err := s.db.QueryContext(ctx, query, rid)
	if err == sql.ErrNoRows {
		return c, nil
	}

	defer rows.Close()
	for rows.Next() {
		var i Category
		if err = rows.Scan(&i.Label, &i.Id); err != nil {
			fmt.Println(i)
			return nil, fmt.Errorf("could not get category: %v", err)
		}

		c = append(c, i)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("could not iterate category: %v", err)
	}

	return c, nil
}

//func (s *Service) GetCategories(ctx context.Context, ctxType, rid string) error {
//	//_, auth := ctx.Value(ctxType).(int64)
//	//if !auth {
//	//	return ErrUnauthenticated
//	//}
//	if !rxUUID.MatchString(rid) {
//		return ErrInvalidRestaurantId
//	}
//
//	query := "SELECT id, name, availability FROM category WHERE restaurant_id = $1"
//	rows, err := s.db.QueryContext(ctx, query, rid)
//	if err == sql.ErrNoRows {
//		return nil, ErrRestaurantNotFound
//	}
//
//	defer rows.Close()
//	uu := make([]Restaurant, 0, 1)
//	for rows.Next() {
//		var r Restaurant
//		var rl int
//		if err = rows.Scan(&r.Id, &r.Title, &rl); err != nil {
//			fmt.Println(r)
//			return nil, fmt.Errorf("could not get title: %v", err)
//		}
//		r.Role = getRole(rl)
//		fmt.Println(r)
//		uu = append(uu, r)
//	}
//
//	if err = rows.Err(); err != nil {
//		return nil, fmt.Errorf("could not iterate restaurants rows: %v", err)
//	}
//
//	return nil
//}

func (s *Service) CreateItem(ctx context.Context, rid string, cid int64, name, description string, price float64, available bool) error {
	uid, auth := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !auth {
		return ErrUnauthenticated
	}

	if !rxUUID.MatchString(rid) {
		return ErrRestaurantNotFound
	}

	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if price < 0 {
		return ErrInvalidPrice
	}

	if _, err := s.checkPermission(ctx, Manager, uid, rid); err != nil {
		return err
	}

	var id string
	query := "INSERT INTO item (restaurant_id, category_id, name, description, price, availability) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	err := s.db.QueryRowContext(ctx, query, rid, cid, name, description, price, available).Scan(&id)
	unique := isUniqueViolation(err)
	if unique {
		return ErrItemAlreadyExists
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetMenuForFp(ctx context.Context, rid string) (*[]Item, error) {
	_, auth := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !auth {
		return nil, ErrUnauthenticated
	}

	if !rxUUID.MatchString(rid) {
		return nil, ErrRestaurantNotFound
	}

	query := `
		SELECT i.id, c.label, c.availability, i.name, i.description, i.price, i.availability
 		FROM category c INNER JOIN item i ON c.id = i.category_id
  		WHERE i.restaurant_id = $1`
	rows, err := s.db.QueryContext(ctx, query, rid)
	if err == sql.ErrNoRows {
		fmt.Println("[NO ROWS]:", err)
		return nil, ErrMenuNotFound
	}

	m := make([]Item, 0)
	defer rows.Close()
	for rows.Next() {
		var i Item
		if err = rows.Scan(&i.Id, &i.Category, &i.CategoryAvailable, &i.Name, &i.Description, &i.Price, &i.Availability); err != nil {
			return nil, fmt.Errorf("could not get item: %v", err)
		}

		m = append(m, i)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("could not iterate menu: %v", err)
	}

	return &m, nil
}

//func (s *Service) GetMenuById(ctx context.Context, id string) (*[]Item, error) {
//	_, auth := ctx.Value(KeyAuthFoodProviderID).(int64)
//	if !auth {
//		return nil, ErrUnauthenticated
//	}
//
//	var d string
//	query := `SELECT name FROM item WHERE id = $1`
//	_ = s.db.QueryRowContext(ctx, query, id).Scan(&d)
//	fmt.Println("[GET MENU 0]:", d)
//
//	return d, nil
//}
