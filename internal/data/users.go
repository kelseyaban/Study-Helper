package data

import (
	"context"
	"database/sql"
	"github.com/abankelsey/study_helper/internal/validator"
	"time"
)

// represents a users entry in the sytem
type Users struct {
	User_id       int64     `json:"user_id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Password_hash string    `json:"password_hash"`
	Created_at    time.Time `json:"created_at"`
}

// validates the fields of the users struct
func ValidateUsers(v *validator.Validator, users *Users) {
	v.Check(validator.NotBlank(users.Username), "username", "This field cannot be left blank")
	v.Check(validator.MaxLength(users.Username, 50), "username", "must not be more than 50 bytes long")
	v.Check(validator.IsValidEmail(users.Email), "email", "This field cannot be left blank")
	v.Check(validator.MaxLength(users.Email, 100), "email", "must not be more than 100 bytes long")
	v.Check(validator.NotBlank(users.Password_hash), "password_hash", "This field cannot be left blank")
	v.Check(validator.MaxLength(users.Password_hash, 50), "password_hash", "must not be more than 50 bytes long")

}

// TodoModel struct handles database operations related to todo
type UsersModel struct {
	DB *sql.DB
}

// Adds new todo entry into the database
func (m *UsersModel) Insert(users *Users) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING user_id, created_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(
		ctx,
		query,
		users.Username,
		users.Email,
		users.Password_hash,
	).Scan(&users.User_id, &users.Created_at)
}
