package service

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/aidosgal/lenshub/internal/model"
)

type OrderService struct {
    db *sql.DB
}

func NewOrderService(db *sql.DB) *OrderService {
    return &OrderService{db: db}
}

func (s *OrderService) CreateOrder(order model.Order) (model.Order, error) {
    query := `
        WITH inserted_order AS (
            INSERT INTO orders (
                title, 
                description, 
                location, 
                user_id, 
                specialization, 
                created_at
            ) VALUES ($1, $2, $3, $4, $5, $6)
            RETURNING id, title, description, location, specialization, created_at, user_id
        )
        SELECT 
            o.id,
            o.title,
            o.description,
            o.location,
            o.specialization,
            o.created_at,
            u.id as user_id,
            u.name,
            u.user_name,
            u.chat_id,
            u.role,
            u.portfolio_url,
            u.specialization as user_specialization
        FROM inserted_order o
        JOIN users u ON o.user_id = u.id`

    var createdOrder model.Order
    var user model.User
    
    err := s.db.QueryRow(
        query,
        order.Title,
        order.Description,
        order.Location,
        order.User.Id,
        order.Specialization,
        order.CreatedAt,
    ).Scan(
        &createdOrder.ID,
        &createdOrder.Title,
        &createdOrder.Description,
        &createdOrder.Location,
        &createdOrder.Specialization,
        &createdOrder.CreatedAt,
        &user.Id,
        &user.Name,
        &user.UserName,
        &user.ChatId,
        &user.Role,
        &user.Portfolio,
        &user.Specialization,
    )

    if err != nil {
        return model.Order{}, fmt.Errorf("error creating order: %v", err)
    }

    createdOrder.User = user
    return createdOrder, nil
}

func (s *OrderService) GetOrderByID(orderID string) (model.Order, error) {
    order_id, err := strconv.Atoi(orderID)
    if err != nil {
        return model.Order{}, err
    }

    query := `
        SELECT 
            o.id,
            o.title,
            o.description,
            o.location,
            o.specialization,
            o.created_at,
            u.id as user_id,
            u.name,
            u.user_name,
            u.chat_id,
            u.role,
            u.portfolio_url,
            u.specialization as user_specialization
        FROM orders o
        JOIN users u ON o.user_id = u.id
        WHERE o.id = $1`

    var order model.Order
    var user model.User

    err = s.db.QueryRow(query, order_id).Scan(
        &order.ID,
        &order.Title,
        &order.Description,
        &order.Location,
        &order.Specialization,
        &order.CreatedAt,
        &user.Id,
        &user.Name,
        &user.UserName,
        &user.ChatId,
        &user.Role,
        &user.Portfolio,
        &user.Specialization,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return model.Order{}, fmt.Errorf("order not found")
        }
        return model.Order{}, fmt.Errorf("error getting order: %v", err)
    }

    order.User = user
    return order, nil
}
