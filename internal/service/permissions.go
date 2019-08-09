package service

import (
	"context"
	"fmt"
)

type permission int

const (
	Admin permission = iota * 5
	Owner
	Manager
	Supervisor
	Waiter
)

func (s *Service) checkPermission(ctx context.Context, level permission, userId int64, RestaurantId string) error {
	var role permission
	query := `SELECT role FROM permission WHERE id = $1 AND restaurant_id = $2`
	err := s.db.QueryRowContext(ctx, query, userId, RestaurantId).Scan(&role)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to read foodprovider's role: %v", err)
	}

	fmt.Println("ROLE:", role)
	fmt.Println("LEVEL:", level)
	if role >= level {
		return ErrUnauthenticated
	}

	return nil
}