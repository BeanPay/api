package models

import (
	"database/sql"
	"errors"
	"time"
)

type User struct {
	Id        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) consumeRow(row *sql.Row) error {
	return row.Scan(
		&u.Id,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
}

type UserRepository struct {
	DB *sql.DB
}

func (p *UserRepository) FetchByEmail(email string) (*User, error) {
	row := p.DB.QueryRow(
		"SELECT * FROM users WHERE email = $1;",
		email,
	)
	user := &User{}
	err := user.consumeRow(row)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (p *UserRepository) FetchByID(id string) (*User, error) {
	row := p.DB.QueryRow(
		"SELECT * FROM users WHERE id = $1;",
		id,
	)
	user := &User{}
	err := user.consumeRow(row)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (p *UserRepository) Insert(user *User) error {
	return user.consumeRow(
		p.DB.QueryRow(
			"INSERT INTO users(email, password) VALUES($1, $2) RETURNING *;",
			user.Email,
			user.Password,
		),
	)
}

func (p *UserRepository) Update(user *User) error {
	return user.consumeRow(
		p.DB.QueryRow(
			"UPDATE users SET email=$1, password=$2 WHERE id=$3 RETURNING *;",
			user.Email,
			user.Password,
			user.Id,
		),
	)
}

func (p *UserRepository) Delete(user *User) error {
	res, err := p.DB.Exec(
		"DELETE FROM users WHERE id=$1;",
		user.Id,
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
