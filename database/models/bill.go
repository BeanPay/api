package models

import (
	"database/sql"
	"errors"
	"time"
)

type Bill struct {
	Id                string    `json:"id"`
	UserId            string    `json:"user_id"`
	Name              string    `json:"name"`
	PaymentURL        string    `json:"payment_url"`
	Frequency         string    `json:"frequency"`
	EstimatedTotalDue float64   `json:"estimated_total_due"`
	FirstDueDate      time.Time `json:"first_due_date"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (b *Bill) consumeRow(row *sql.Row) error {
	return row.Scan(
		&b.Id,
		&b.UserId,
		&b.Name,
		&b.PaymentURL,
		&b.Frequency,
		&b.EstimatedTotalDue,
		&b.FirstDueDate,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
}

type BillRepository struct {
	DB *sql.DB
}

func (r *BillRepository) fetch(query string, args ...interface{}) ([]*Bill, error) {
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	bills := make([]*Bill, 0)
	for rows.Next() {
		b := &Bill{}
		err := rows.Scan(
			&b.Id,
			&b.UserId,
			&b.Name,
			&b.PaymentURL,
			&b.Frequency,
			&b.EstimatedTotalDue,
			&b.FirstDueDate,
			&b.CreatedAt,
			&b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bills = append(bills, b)
	}
	return bills, nil
}

func (r *BillRepository) FetchAllUserBills(userId string) ([]*Bill, error) {
	return r.fetch("SELECT * FROM bills WHERE user_id = $1", userId)
}

func (r *BillRepository) FetchByID(id string) (*Bill, error) {
	row := r.DB.QueryRow(
		"SELECT * FROM bills WHERE id = $1;",
		id,
	)
	bill := &Bill{}
	err := bill.consumeRow(row)
	if err != nil {
		return nil, err
	}
	return bill, nil
}

func (r *BillRepository) Insert(bill *Bill) error {
	return bill.consumeRow(
		r.DB.QueryRow(
			`INSERT INTO bills(user_id, name, payment_url, frequency, estimated_total_due, first_due_date)
			VALUES($1, $2, $3, $4, $5, $6)
			RETURNING *;`,
			bill.UserId,
			bill.Name,
			bill.PaymentURL,
			bill.Frequency,
			bill.EstimatedTotalDue,
			bill.FirstDueDate,
		),
	)
}

func (r *BillRepository) Update(bill *Bill) error {
	return bill.consumeRow(
		r.DB.QueryRow(
			`UPDATE bills
			SET
				name=$1,
				payment_url=$2,
				frequency=$3,
				estimated_total_due=$4,
				first_due_date=$5
			WHERE id = $6
			RETURNING *;`,
			bill.Name,
			bill.PaymentURL,
			bill.Frequency,
			bill.EstimatedTotalDue,
			bill.FirstDueDate,
			bill.Id,
		),
	)
}

func (r *BillRepository) Delete(bill *Bill) error {
	res, err := r.DB.Exec(
		"DELETE FROM bills WHERE id=$1;",
		bill.Id,
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
