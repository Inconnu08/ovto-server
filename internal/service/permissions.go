package service

import (
	"context"
	"database/sql"
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

func (s *Service) checkPermission(ctx context.Context, level permission, userId int64, RestaurantId string) (string, error) {
	var role permission
	var title string
	query := `SELECT restaurant, role FROM permission WHERE id = $1 AND restaurant_id = $2`
	err := s.db.QueryRowContext(ctx, query, userId, RestaurantId).Scan(&title, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrRestaurantNotFound
		}
		return "", fmt.Errorf("failed to read foodprovider's role: %v", err)
	}

	fmt.Println("ROLE:", role)
	fmt.Println("LEVEL:", level)
	if role >= level {
		return "", ErrUnauthenticated
	}

	return title, nil
}

func getRole(role int) string {
	switch role {
	case 0:
		return "Admin"
	case 5:
		return "Owner"
	case 10:
		return "Manager"
	case 15:
		return "Supervisor"
	case 20:
		return "Waiter"
	}

	return "Waiter"
}

func getRoleLevel(role string) (permission, error) {
	switch role {
	case "Admin":
		return Admin, nil
	case "Owner":
		return Owner, nil
	case "Manager":
		return Manager, nil
	case "Supervisor":
		return Supervisor, nil
	case "Waiter":
		return Waiter, nil
	}

	return 0, fmt.Errorf("invalid role level")
}
