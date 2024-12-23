package service

import (
	"database/sql"
	"fmt"
)

type ResponseService struct {
    db *sql.DB
}

func NewResponseService(db *sql.DB) *ResponseService {
    return &ResponseService{db: db}
}

func (s *ResponseService) CreateOrderResponse(orderID string, executorID int) error {
    query := `
        INSERT INTO responses(
            order_id,
            user_id,
            created_at
        ) VALUES ($1, $2, NOW())
        RETURNING id`

    var responseID int
    err := s.db.QueryRow(query, orderID, executorID).Scan(&responseID)
    if err != nil {
        return fmt.Errorf("error creating order response: %v", err)
    }

    return nil
}
