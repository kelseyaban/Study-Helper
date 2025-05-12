package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/abankelsey/study_helper/internal/validator"
)

// represents a quote entry in the sytem
type Quotes struct {
	Quote_id   int64     `json:"quote_id"`
	User_id    int64     `json:"user_id"`
	Content    string    `json:"content"`
	Created_at time.Time `json:"created_at"`
}

// validates the fields of the quotes struct
func ValidateQuotes(v *validator.Validator, quotes *Quotes) {
	v.Check(validator.NotBlank(quotes.Content), "content", "This field cannot be left blank")
	v.Check(validator.MaxLength(quotes.Content, 50), "content", "must not be more than 50 bytes long")

}

// QuotesModel struct handles database operations related to todo
type QuotesModel struct {
	DB *sql.DB
}

// Adds new todo entry into the database
func (m *QuotesModel) Insert(quotes *Quotes) error {
	query := `
		INSERT INTO quotes (content, user_id)
		VALUES ($1,$2)
		RETURNING quote_id, created_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(
		ctx,
		query,
		quotes.Content,
		quotes.User_id,
	).Scan(&quotes.Quote_id, &quotes.Created_at)
}

// Retrieve list of all quote entries from the database
func (m *QuotesModel) QuoteList(userID int64) ([]*Quotes, error) {
	query := `
        SELECT quote_id, content, user_id, created_at
        FROM quotes
        WHERE user_id = $1
        ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []*Quotes

	for rows.Next() {
		q := &Quotes{}
		err := rows.Scan(&q.Quote_id, &q.Content, &q.User_id, &q.Created_at)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, q)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return quotes, nil
}

func (m *QuotesModel) DeleteQuote(quoteID int64, userID int64) error {
	query := `
    DELETE FROM quotes WHERE quote_id = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, quoteID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
