package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/abankelsey/study_helper/internal/validator"
)

// represents a session entry in the sytem
type Sessions struct {
	Session_id   int64     `json:"session_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Subject      string    `json:"subject"`
	Start_date   time.Time `json:"start_date"`
	End_date     time.Time `json:"end_date"`
	Is_completed bool      `json:"is_completed"`
	Created_at   time.Time `json:"created_at"`
}

// validates the fields of the sessions struct
func ValidateSessions(v *validator.Validator, sessions *Sessions) {
	v.Check(validator.NotBlank(sessions.Title), "title", "This field cannot be left blank")
	v.Check(validator.MaxLength(sessions.Title, 50), "title", "must not be more than 50 bytes long")
	v.Check(validator.NotBlank(sessions.Description), "description", "This field cannot be left blank")
	v.Check(validator.MaxLength(sessions.Description, 50), "description", "must not be more than 50 bytes long")
	v.Check(validator.NotBlank(sessions.Subject), "subject", "This field cannot be left blank")
	v.Check(validator.MaxLength(sessions.Subject, 50), "subject", "must not be more than 50 bytes long")
}

type SessionsModel struct {
	DB *sql.DB
}

// Adds new todo entry into the database
func (m *SessionsModel) Insert(sessions *Sessions) error {
	query := `
		INSERT INTO study_sessions (title, description, subject, start_date, end_date, is_completed)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING session_id, created_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(
		ctx,
		query,
		sessions.Title,
		sessions.Description,
		sessions.Subject,
		sessions.Start_date,
		sessions.End_date,
		sessions.Is_completed,
	).Scan(&sessions.Session_id, &sessions.Created_at)
}

// Retrieve list of all session entries from the database
func (m *SessionsModel) SessionList() ([]*Sessions, error) {
	query := `
        SELECT session_id, title, description, subject, start_date, end_date, is_completed
        FROM study_sessions
        ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Sessions

	for rows.Next() {
		s := &Sessions{}
		err := rows.Scan(&s.Session_id, &s.Title, &s.Description, &s.Subject, &s.Start_date, &s.End_date, &s.Is_completed)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// DeleteSession removes a session entry from the database using its ID
func (m *SessionsModel) DeleteSession(sessionID int64) error {
	query := `
	DELETE FROM study_sessions WHERE session_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, sessionID)
	return err
}

// Get the session info based on the session
func (m *SessionsModel) GetSessionByID(id int64) (*Sessions, error) {
	stmt := `
	SELECT session_id, title, description, subject, start_date, end_date, is_completed FROM study_sessions WHERE session_id = $1`
	row := m.DB.QueryRow(stmt, id)

	var s Sessions
	err := row.Scan(&s.Session_id, &s.Title, &s.Description, &s.Subject, &s.Start_date, &s.End_date, &s.Is_completed)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// Edits an entry session into the database
func (m *SessionsModel) EditSession(session *Sessions) error {
	query := `
        UPDATE study_sessions
        SET title = $1,
			description = $2,
			subject = $3,
			start_date = $4,
			end_date = $5,
            is_completed = $6
        WHERE session_id = $7`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(
		ctx,
		query,
		session.Title,
		session.Description,
		session.Subject,
		session.Start_date,
		session.End_date,
		session.Is_completed,
		session.Session_id,
	)
	return err
}
