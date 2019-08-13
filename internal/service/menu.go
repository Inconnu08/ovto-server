package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Item struct {
	Id 				  string  `json:"id"`
	Name              string  `json:"name,omitempty"`
	Category          string  `json:"category,omitempty"`
	CategoryAvailable string  `json:"category_available,omitempty"`
	Description       string  `json:"description,omitempty"`
	Price             float64 `json:"price,omitempty"`
	Availability      bool    `json:"availability,omitempty"`
}

type Category struct {
	Label string `json:"label"`
	Id    string `json:"id"`
}

//type Category struct {
//	Menu *[]Item `json:"Menu,omitempty"`
//}

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
	fmt.Println("ROWS:", err)
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

		fmt.Println(i)
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

func (s *Service) CreateItem(ctx context.Context, rid, cid, name, description string, price float64, available bool) error {
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
		fmt.Println("Permission Failed!")
		return ErrUnauthenticated
	}

	query := "INSERT INTO item (restaurant_id , category_id, name, description, price, availability) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := s.db.QueryContext(ctx, query, rid, cid, name, description, price, available)
	//unique := isUniqueViolation(err)
	//if !unique && err != nil {
	//	return err
	//}
	//if unique {
	//	return ErrEmailTaken
	//}
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetMenuForFp(ctx context.Context, rid string) ([]Item, error) {
	m := make([]Item, 0, 1)

	_, auth := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !auth {
		return m, ErrUnauthenticated
	}

	if !rxUUID.MatchString(rid) {
		return m, ErrRestaurantNotFound
	}
//, c.label, i.name, c.availability, i.description, i.price, i.availability 		INNER JOIN item i ON c.id = i.category_id
	query := `SELECT name FROM item WHERE restaurant_id = $1`
	rows, err := s.db.QueryContext(ctx, query, rid)
	fmt.Println("[GET MENU]:", err)
	fmt.Println("[GET MENU]:", rows)
	if err == sql.ErrNoRows {
		fmt.Println("[NO ROWS]:", err)
		return m, nil
	}

	defer rows.Close()
	fmt.Println("[NEXT]:", rows.Next())
	for rows.Next() {
		fmt.Println("[ENTER LOOP]")
		var i Item
		if err = rows.Scan(&i.Id, &i.Category, &i.Name, &i.CategoryAvailable, &i.Description, &i.Price, &i.Availability); err != nil {
			fmt.Println("[ITEM]", i)
			return nil, fmt.Errorf("could not get item: %v", err)
		}

		fmt.Println(i)
		m = append(m, i)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("could not iterate menu: %v", err)
	}

	return m, nil
}
