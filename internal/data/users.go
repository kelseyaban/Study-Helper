package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/abankelsey/study_helper/internal/validator"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// represents a users entry in the sytem
type Users struct {
	User_id       int64     `json:"user_id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Password_hash []byte    `json:"password_hash"`
	Activated     bool      `json:"activated"`
	Created_at    time.Time `json:"created_at"`
}

// validates the fields of the users struct
func ValidateUsers(v *validator.Validator, users *Users, password string) {
	v.Check(validator.NotBlank(users.Name), "name", "This field cannot be left blank")
	v.Check(validator.MaxLength(users.Name, 50), "name", "Must not be more than 50 characters long")
	v.Check(validator.NotBlank(users.Email), "email", "This field cannot be left blank")
	v.Check(validator.IsValidEmail(users.Email), "email", "Must be a valid email address")
	v.Check(validator.MaxLength(users.Email, 100), "email", "Must not be more than 100 characters long")

	v.Check(validator.NotBlank(password), "password", "This field cannot be left blank")
	v.Check(validator.MinLength(password, 8), "password", "Password must be at least 8 characters long")
	v.Check(validator.MaxLength(password, 72), "password", "Password must not be more than 72 characters long") // bcrypt max
	v.Check(validator.HasNumber(password), "password", "Password must contain at least one number")
	v.Check(validator.HasUpper(password), "password", "Password must contain at least one uppercase letter")
	v.Check(validator.HasSymbol(password), "password", "Password must contain at least one special character (!@#$ etc.)")
}

// TodoModel struct handles database operations related to todo
type UsersModel struct {
	DB *sql.DB
}

var ErrInvalidCredentials = errors.New("invalid credentials")

// Insert a new user into the database with hashed password
func (m *UsersModel) Insert(users *Users, plainPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), 12)
	if err != nil {
		return err
	}

	users.Password_hash = hashedPassword

	query := `
       INSERT INTO users (name, email, password_hash, activated)
       VALUES ($1, $2, $3, $4)
       RETURNING user_id, created_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(
		ctx, query,
		users.Name, users.Email, users.Password_hash, users.Activated,
	).Scan(&users.User_id, &users.Created_at)
}

// Authenticate checks if a user exists and the password is correct
func (m *UsersModel) Authenticate(email, plainPassword string) (*Users, error) {
	var user Users

	query := `
        SELECT user_id, password_hash
        FROM users
        WHERE email = $1
		AND activated = TRUE`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.User_id,
		&user.Password_hash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check hashed password
	err = bcrypt.CompareHashAndPassword(user.Password_hash, []byte(plainPassword))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

// GetUser fetches a user by ID
func (m *UsersModel) GetUser(userID int64) (*Users, error) {
	var user Users

	query := `
        SELECT user_id, username, email, password_hash, activated, created_at
        FROM users
        WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.User_id,
		&user.Name,
		&user.Email,
		&user.Password_hash,
		&user.Activated,
		&user.Created_at,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &user, nil
}
