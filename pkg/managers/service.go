package managers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
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

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Manager struct {
	ID      	int64     `json:"id"`
	Name    	string    `json:"name"`
	Salary  	string    `json:"salary"`
	Plan		string	  `json:"plan"`
	Boss_id 	int64	  `json:"boss_id"`
	Departament string	  `json:"departament"`
	Phone		string	  `json:"login"`
	Password	string	  `json:"password"`
	Roles		[]string  `json:"roles"`
	Active      bool      `json:"active"`
	Created     time.Time `json:"created"`
}

type Registration struct {
	ID		 int64 	`json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Roles    string `json:"roles"`
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

func (s *Service) Register(ctx context.Context, reg *Registration) ( string, error) {
	item := &Manager{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO managers (name, phone, roles) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (phone) DO NOTHING 
		RETURNING id, name, phone, active, created
	`, reg.Name, reg.Phone, reg.Roles).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
	if err != nil {
		log.Print(err)
		return "", ErrInternal
	}

	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token := hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO managers_tokens(token, customer_id) VALUES($1, $2)`, token, item.ID)
	if err != nil {
		return "", ErrInternal
	}
	return token, nil
}

func (s *Service) Token(ctx context.Context, phone string, password string) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(ctx, `SELECT id, password FROM customers WHERE phone = $1`, phone).Scan(&id, &hash)

	if err == pgx.ErrNoRows {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", ErrInternal
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return "", ErrPasswordInvalid
	}

	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token = hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO managers_token(token, manager_id) VALUES($1, $2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}
	return token, nil
}

func (s *Service) AuthentificateManager(ctx context.Context, token string) (id int64, err error) {
	expireTime := time.Now()
	err = s.pool.QueryRow(ctx, `SELECT manager_id, expire FROM managers_tokens WHERE token = $1`, token).Scan(&id, &expireTime)

	if err == pgx.ErrNoRows {
		return 0, ErrUserNotFound
	}
	if err != nil {
		return 0, ErrInternal
	}
	if time.Now().After(expireTime) {
		return 0, ErrTokenExpired
	}
	return id, nil
}