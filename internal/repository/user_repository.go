package repository

import (
	"database/sql"

	"github.com/aidosgal/lenshub/internal/model"
)

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func(r *UserRepository) CreateUser(user model.User) error {
    query := `
        INSERT INTO users (name, user_name, chat_id, role, portfolio_url, specialization)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

    _, err := r.db.Exec(query, user.Name, user.UserName, user.ChatId, user.Role, user.Portfolio, user.Specialization)
    if err != nil {
        return err
    }
    return nil
}

func (r *UserRepository) GetUserByChatID(chatID string) (*model.User, error) {
    query := `
        SELECT id, name, user_name, chat_id, role, portfolio_url, specialization 
        FROM users 
        WHERE chat_id = $1
    `
    
    user := &model.User{}
    err := r.db.QueryRow(query, chatID).Scan(
        &user.Id,
        &user.Name,
        &user.UserName,
        &user.ChatId,
        &user.Role,
        &user.Portfolio,
        &user.Specialization,
    )

    if err == sql.ErrNoRows {
        return nil, nil // User not found
    }

    if err != nil {
        return nil, err // Database error
    }

    return user, nil
}

func (r *UserRepository) GetUsersBySpecialization(specialization string) ([]model.User, error) {
    query := `
        SELECT name, user_name, chat_id, role, portfolio_url, specialization 
        FROM users 
        WHERE specialization = $1 AND role = 'Исполнитель'
    `
    
    rows, err := r.db.Query(query, specialization)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []model.User
    for rows.Next() {
        var user model.User
        if err := rows.Scan(
            &user.Name,
            &user.UserName,
            &user.ChatId,
            &user.Role,
            &user.Portfolio,
            &user.Specialization,
        ); err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    return users, nil
}
