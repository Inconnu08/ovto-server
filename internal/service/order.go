package service

import (
	"context"
	"database/sql"
	"fmt"
)

type Order struct {
	Id     int64            `json:"id"`
	CId    int64            `json:"cid"`
	RId    string           `json:"rid"`
	Status int64            `json:"status"`
	Items  map[string]int64 `json:"items"`
}

type orderClient struct {
	orders       chan Order
	restaurantID string
}

func (s *Service) CreateOrder(ctx context.Context, rid string, cid, status int64, items map[string]int64) error {
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
	defer func() { _ = tx.Rollback() }()

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

	var o Order
	o.Id = orderId
	o.CId = cid
	go s.orderCreated(o)

	return nil
}

func (s *Service) CreateUserOrder(ctx context.Context, rid string, status int64, items map[string]int64) error {
	uid, auth := ctx.Value(KeyAuthUserID).(int64)
	if !auth {
		return ErrUnauthenticated
	}

	if !rxUUID.MatchString(rid) {
		return ErrInvalidRestaurantId
	}

	var orderId int64
	query := "INSERT INTO orders(cust_id, restaurant_id, status) VALUES ($1, $2, $3) RETURNING id"
	err := s.db.QueryRowContext(ctx, query, uid, rid, 1).Scan(&orderId)
	fk := isForeignKeyViolation(err)
	if fk {
		fmt.Println("[FK] ", err)
		return ErrRestaurantNotFound
	}

	fmt.Println("[ORDER ID] ", orderId)

	var o Order
	o.Id = orderId
	o.CId = uid
	o.Items = items
	o.RId = rid
	go s.orderCreated(o)

	return nil
}

func (s *Service) orderCreated(o Order) {
	//u, err := s.userByID(context.Background(), o.CId)
	//if err != nil {
	//	log.Printf("could not fetch comment user: %v\n", err)
	//	return
	//}
	//
	//c.User = &u
	//c.Mine = false
	//
	// go s.notifyOrder(c)
	go s.broadcastOrder(o)
}

// SubscribeToOrders to receive orders in realtime.
func (s *Service) OrdersStream(ctx context.Context, rID string) (<-chan Order, error) {
	oo := make(chan Order)
	client := &orderClient{orders: oo, restaurantID: rID}
	//if uid, auth := ctx.Value(KeyAuthUserID).(string); auth {
	//	client.userID = &uid
	//}
	s.orderClients.Store(client, nil)

	go func() {
		<-ctx.Done()
		s.orderClients.Delete(client)
		close(oo)
	}()

	return oo, nil
}

func (s *Service) GetOrders(ctx context.Context, rid string) (*[]Order, error) {
	uid, auth := ctx.Value(KeyAuthFoodProviderID).(int64)
	if !auth {
		return nil, ErrUnauthenticated
	}

	if !rxUUID.MatchString(rid) {
		return nil, ErrInvalidRestaurantId
	}

	if _, err := s.checkPermission(ctx, Waiter, uid, rid); err != nil {
		fmt.Println("[Permission Failed]:", err)
		return nil, err
	}

	o := make([]Order, 0, 1)
	query := `
		SELECT id, cust_id
 		FROM orders
		WHERE restaurant = $1 AND status NOT 5`
	rows, err := s.db.QueryContext(ctx, query, rid)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	defer rows.Close()
	for rows.Next() {
		var i Order
		if err = rows.Scan(&i.CId); err != nil {
			fmt.Println(i)
			return nil, fmt.Errorf("could not get order: %v", err)
		}

		o = append(o, i)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("could not iterate orders: %v", err)
	}

	return &o, nil
}

func (s *Service) broadcastOrder(o Order) {
	s.orderClients.Range(func(key, _ interface{}) bool {
		client, ok := key.(*orderClient)
		if !ok {
			log.Println("broadcast order: no client type")
			return false
		}

		if client.restaurantID == o.RId {
			client.orders <- o
		}

		return true
	})
}
