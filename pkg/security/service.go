package security

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)
type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Auth(login, password string) (ok bool) {
	err := s.pool.QueryRow(context.Background(), `SELECT login FROM managers WHERE login = $1 and password = $2`, login, password).Scan(&login)
	if err != nil {
		return false
	}
	return true
}