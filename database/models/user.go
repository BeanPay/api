package models

import (
	"database/sql"
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

func (u *User) consumeNextRow(rows *sql.Rows) error {
	return rows.Scan(
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

func (p *UserRepository) Insert(user User) (*User, error) {
	row := p.DB.QueryRow(
		"INSERT INTO users(email, password) VALUES($1, $2) RETURNING *;",
		user.Email,
		user.Password,
	)
	newUser := &User{}
	err := newUser.consumeRow(row)
	if err != nil {
		return nil, err
	}
	return newUser, nil
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

func (p *UserRepository) Update(user User) error {
	return nil
}

func (p *UserRepository) Delete(user User) error {
	return nil
}
