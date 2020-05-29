package models

import (
	"database/sql"
	"errors"
	"time"
)

type Payment struct {
	Id        string    `json:"id"`
	BillId    string    `json:"bill_id"`
	DueDate   time.Time `json:"due_date"`
	TotalPaid float64   `json:"total_paid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Payment) consumeRow(row *sql.Row) error {
	return row.Scan(
		&p.Id,
		&p.BillId,
		&p.DueDate,
		&p.TotalPaid,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
}

type PaymentRepository struct {
	DB *sql.DB
}

func (r *PaymentRepository) fetch(query string, args ...interface{}) ([]*Payment, error) {
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	payments := make([]*Payment, 0)
	for rows.Next() {
		p := &Payment{}
		err := rows.Scan(
			&p.Id,
			&p.BillId,
			&p.DueDate,
			&p.TotalPaid,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

// Returns all payments made by a specific between the dates 'from' (inclusive) and 'to' (exclusive).
func (r *PaymentRepository) FetchAllUserPayments(userId string, from time.Time, to time.Time) ([]*Payment, error) {
	return r.fetch(
		`SELECT *
		FROM payments
		WHERE bill_id IN (
			SELECT id
			FROM bills
			WHERE user_id = $1
		)
		AND due_date >= $2 AND due_date < $3;`,
		userId,
		from,
		to,
	)
}

func (r *PaymentRepository) FetchByID(id string) (*Payment, error) {
	row := r.DB.QueryRow(
		"SELECT * FROM payments WHERE id = $1;",
		id,
	)
	payment := &Payment{}
	err := payment.consumeRow(row)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (r *PaymentRepository) Insert(payment *Payment) error {
	return payment.consumeRow(
		r.DB.QueryRow(
			`INSERT INTO payments(bill_id, due_date, total_paid)
			VALUES($1, $2, $3)
			RETURNING *;`,
			payment.BillId,
			payment.DueDate,
			payment.TotalPaid,
		),
	)
}

func (r *PaymentRepository) Delete(payment *Payment) error {
	res, err := r.DB.Exec(
		"DELETE FROM payments WHERE id=$1;",
		payment.Id,
	)
	if err != nil {
		return err
	}
	numRows, _ := res.RowsAffected()
	if numRows != 1 {
		return errors.New("Nothing was deleted.")
	}
	return nil
}
