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
		INSERT INTO quotes (content)
		VALUES ($1)
		RETURNING quote_id, created_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(
		ctx,
		query,
		quotes.Content,
	).Scan(&quotes.Quote_id, &quotes.Created_at)
}

// Retrieve list of all quote entries from the database
func (m *QuotesModel) QuoteList() ([]*Quotes, error) {
	query := `
        SELECT quote_id, content
        FROM quotes
        ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []*Quotes

	for rows.Next() {
		q := &Quotes{}
		err := rows.Scan(&q.Quote_id, &q.Content)
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

// DeleteQuote removes a quote entry from the database using its ID
func (m *QuotesModel) DeleteQuote(quoteID int64) error {
	query := `
	DELETE FROM quotes WHERE quote_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, quoteID)
	return err
}
