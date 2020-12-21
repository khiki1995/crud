package customers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
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

type Customer struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Phone   string    `json:"phone"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}

type Registration struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type Auth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Product struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Qty   int    `json:"qty"`
}

type Purchase struct {
	Date     time.Time  `json:"date"`
	Products []*Product `json:"products"`
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
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
	_, err = s.pool.Exec(ctx, `INSERT INTO customers_tokens(token, customer_id) VALUES($1, $2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}
	return token, nil
}

func (s *Service) IDByToken(ctx context.Context, token string) (id int64, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT customer_id FROM customers_tokens WHERE token = $1
	`, token).Scan(&id)

	if err == pgx.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, ErrInternal
	}

	return id, nil
}

func (s *Service) AuthentificateCustomer(ctx context.Context, token string) (id int64, err error) {
	expireTime := time.Now()
	err = s.pool.QueryRow(ctx, `SELECT customer_id, expire FROM customers_tokens WHERE token = $1`, token).Scan(&id, &expireTime)

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

func (s *Service) Register(ctx context.Context, reg *Registration) (*Customer, error) {
	item := &Customer{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO customers (name, phone, password) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (phone) DO NOTHING 
		RETURNING id, name, phone, active, created
	`, reg.Name, reg.Phone, reg.Password).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return item, nil
}

func (s *Service) Products(ctx context.Context) ([]*Product, error) {
	items := make([]*Product, 0)
	rows, err := s.pool.Query(ctx, `
	SELECT id, name, price, qty FROM products WHERE active ORDER BY id LIMIT 500
	`)
	if errors.Is(err, pgx.ErrNoRows) {
		return items, nil
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &Product{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price, &item.Qty)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return items, nil
}

func (s *Service) Purchases(ctx context.Context, id int64) ([]*Purchase, error) {
	items := make([]*Purchase, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT s.created as Date, sp.product_id as ID, sp.name as Name, sp.price as Price, sp.qty as Qty
		FROM sales s
		INNER JOIN sale_positions sp ON sp.sale_id = s.id and s.customer_id = $1
		GROUP BY s.id, sp.product_id, sp.name, sp.price, sp.qty
		ORDER BY s.created
	`, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return items, nil
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		purchase := &Purchase{}
		product := &Product{}
		err = rows.Scan(&purchase.Date, &product.ID, &product.Name, &product.Price, &product.Qty)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		found := false
		for _, p := range items {
			if p.Date == purchase.Date {
				p.Products = append(p.Products, product)
				found = true
				break
			}
		}
		if !found {
			purchase.Products = append(purchase.Products, product)
			items = append(items, purchase)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return items, nil
}
