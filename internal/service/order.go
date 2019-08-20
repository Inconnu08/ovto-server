package service

import (
	"context"
	"fmt"
)

func (s *Service) CreateOrder(ctx context.Context, rid, cid, status string, total float64, items map[int64]int64) error {
	uid, auth := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !auth {
		return ErrUnauthenticated
	}

	if !rxUUID.MatchString(rid) {
		return ErrInvalidRestaurantId
	}

	if _, err := s.checkPermission(ctx, Waiter, uid, rid); err != nil {
		fmt.Println("[Permission Failed]:", err)
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	var orderId int64
	query := "INSERT INTO orders(cust_id, restaurant_id, status) VALUES ($1, $2, $3) RETURNING id"
	err = tx.QueryRowContext(ctx, query, cid, rid, 1).Scan(&orderId)
	fk := isForeignKeyViolation(err)
	if fk {
		fmt.Println("[FK] ", err)
		return ErrRestaurantNotFound
	}

	fmt.Println("[ORDER ID] ", orderId)

	for item, quantity := range items {
		_, err = tx.ExecContext(ctx, "INSERT INTO order_item(order_id, item_id, quantity) VALUES ($1, $2, $3)", orderId, item, quantity)
		if err != nil {
			fmt.Println("[ORDER ITEMS] ", err)
			return err
		}
	}

	if err != nil {
		return fmt.Errorf("failed to create order: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to create order: could not commit transaction: %v", err)
	}

	return nil
}