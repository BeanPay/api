package models

import (
	"database/sql"
	"errors"
	"time"
)

type RefreshToken struct {
	Id        string    `json:"id"`
	ChainId   string    `json:"chain_id"`
	UserId    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *RefreshToken) consumeRow(row *sql.Row) error {
	return row.Scan(
		&r.Id,
		&r.ChainId,
		&r.UserId,
		&r.CreatedAt,
	)
}

type RefreshTokenRepository struct {
	DB *sql.DB
}

func (r *RefreshTokenRepository) FetchByID(id string) (*RefreshToken, error) {
	row := r.DB.QueryRow(
		"SELECT * FROM refresh_tokens WHERE id = $1;",
		id,
	)
	refreshToken := &RefreshToken{}
	err := refreshToken.consumeRow(row)
	if err != nil {
		return nil, err
	}
	return refreshToken, nil
}

func (r *RefreshTokenRepository) FetchMostRecentInChain(chainId string) (*RefreshToken, error) {
	row := r.DB.QueryRow(
		"SELECT * FROM refresh_tokens WHERE chain_id=$1 ORDER BY created_at DESC LIMIT 1;",
		chainId,
	)
	refreshToken := &RefreshToken{}
	err := refreshToken.consumeRow(row)
	if err != nil {
		return nil, err
	}
	return refreshToken, nil
}

func (r *RefreshTokenRepository) DeleteChain(chainId string) error {
	res, err := r.DB.Exec(
		"DELETE FROM refresh_tokens WHERE chain_id=$1;",
		chainId,
	)
	if err != nil {
		return err
	}
	numRows, _ := res.RowsAffected()
	if numRows < 1 {
		return errors.New("Nothing was deleted.")
	}
	return nil
}

func (r *RefreshTokenRepository) Insert(refreshToken *RefreshToken) error {
	return refreshToken.consumeRow(
		r.DB.QueryRow(
			"INSERT INTO refresh_tokens(chain_id, user_id) VALUES($1, $2) RETURNING *;",
			refreshToken.ChainId,
			refreshToken.UserId,
		),
	)
}
