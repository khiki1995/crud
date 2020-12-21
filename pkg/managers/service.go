package managers

import (
	"context"
	"errors"
	"time"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrInternal = errors.New("Internal error")
var ErrUserNotFound = errors.New("no such user")
var ErrPhoneUsed = errors.New("phone already registered")
var ErrTokenExpired = errors.New("expired")
var ErrTokenNotFound = errors.New("expired")
var ErrPasswordInvalid = errors.New("invalid password")

type Service struct {
	pool *pgxpool.Pool
}

type Manager struct {
	ID      	int64     `json:"id"`
	Name    	string    `json:"name"`
	Salary  	string    `json:"salary"`
	Plan		string	  `json:"plan"`
	Boss_id 	int64	  `json:"boss_id"`
	Departament string	  `json:"departament"`
	Login		string	  `json:"login"`
	Passwor 	string	  `json:"password"`
	Active      bool      `json:"active"`
	Created     time.Time `json:"created"`
}

func (s *Service) IDByToken(ctx context.Context, token string) (id int64, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT manager_id FROM managers_tokens WHERE token = $1
	`, token).Scan(&id)

	if err == pgx.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, ErrInternal
	}

	return id, nil
}