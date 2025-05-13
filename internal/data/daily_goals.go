package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/abankelsey/study_helper/internal/validator"
)

// represents a goals entry in the sytem
type Goals struct {
	Goal_id      int64     `json:"goal_id"`
	User_id      int64     `json:"user_id"`
	Goal_text    string    `json:"goal_text"`
	Is_completed bool      `json:"is_completed"`
	Target_date  time.Time `json:"target_date"`
	Created_at   time.Time `json:"created_at"`
}

// validates the fields of the goals struct
func ValidateGoals(v *validator.Validator, goals *Goals) {
	v.Check(validator.NotBlank(goals.Goal_text), "goal_text", "This field cannot be left blank")
	v.Check(validator.MaxLength(goals.Goal_text, 50), "goal_text", "must not be more than 50 bytes long")

}

// GoalsModel struct handles database operations related to todo
type GoalsModel struct {
	DB *sql.DB
}

// Adds new todo entry into the database
func (m *GoalsModel) Insert(goals *Goals) error {
	query := `
        INSERT INTO daily_goals (user_id, goal_text, is_completed, target_date)
        VALUES ($1, $2, $3, $4)
        RETURNING goal_id, created_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(
		ctx,
		query,
		goals.User_id,
		goals.Goal_text,
		goals.Is_completed,
		goals.Target_date,
	).Scan(&goals.Goal_id, &goals.Created_at)
}

// Retrieve list of all daily goal entries from the database
func (m *GoalsModel) GoalList(userID int64) ([]*Goals, error) {
	query := `
        SELECT goal_id, user_id, goal_text, target_date, is_completed, created_at
        FROM daily_goals
        WHERE user_id = $1
        ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []*Goals

	for rows.Next() {
		g := &Goals{}
		err := rows.Scan(&g.Goal_id, &g.User_id, &g.Goal_text, &g.Target_date, &g.Is_completed, &g.Created_at)
		if err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return goals, nil
}

// DeleteGoal removes a goal entry from the database using its ID
func (m *GoalsModel) DeleteGoal(goalID int64, userID int64) error {
	query := `
	DELETE FROM daily_goals WHERE goal_id = $1 and user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, goalID, userID)
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

// Get the goal info based on the goal
func (m *GoalsModel) GetGoalByID(id int64) (*Goals, error) {
	stmt := `
    SELECT goal_id, user_id, goal_text, is_completed, target_date, created_at 
    FROM daily_goals 
    WHERE goal_id = $1`

	row := m.DB.QueryRow(stmt, id)

	var g Goals
	err := row.Scan(&g.Goal_id, &g.User_id, &g.Goal_text, &g.Is_completed, &g.Target_date, &g.Created_at)
	if err != nil {
		return nil, err
	}

	return &g, nil
}

// Edits an entry goal into the database
func (m *GoalsModel) EditGoal(goal *Goals) error {
	query := `
        UPDATE daily_goals
        SET goal_text = $1,
            is_completed = $2,
            target_date = $3
        WHERE goal_id = $4`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(
		ctx,
		query,
		goal.Goal_text,
		goal.Is_completed,
		goal.Target_date,
		goal.Goal_id,
	)
	return err
}
