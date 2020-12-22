package managers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/khiki1995/crud/pkg/customers"

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

type Auth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Manager struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Salary      string    `json:"salary"`
	Plan        string    `json:"plan"`
	Boss_id     int64     `json:"boss_id"`
	Departament string    `json:"departament"`
	Phone       string    `json:"phone"`
	Password    string    `json:"password"`
	Roles       []string  `json:"roles"`
	Active      bool      `json:"active"`
	Created     time.Time `json:"created"`
}

type Registration struct {
	ID    int64    `json:"id"`
	Name  string   `json:"name"`
	Phone string   `json:"phone"`
	Roles []string `json:"roles"`
}

type Product struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Price   int       `json:"price"`
	Qty     int       `json:"qty"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}
type SalePosition struct {
	ID         int64 `json:"id"`
	Product_id int64 `json:"product_id"`
	Qty        int   `json:"qty"`
	Price      int   `json:"price"`
}
type Sale struct {
	ID          int64           `json:"id"`
	Manager_id  int64           `json:"manager_id"`
	Customer_id int64           `json:"customer_id"`
	Created     time.Time       `json:"created"`
	Positions   []*SalePosition `json:"positions"`
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) GetToken(ctx context.Context, phone string, password string) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(ctx, `SELECT id, password FROM managers WHERE phone = $1`, phone).Scan(&id, &hash)

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
	_, err = s.pool.Exec(ctx, `INSERT INTO managers_tokens(token, manager_id) VALUES($1, $2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}
	return token, nil
}

func (s *Service) IDByToken(ctx context.Context, token string) (id int64, err error) {
	expireTime := time.Now()
	err = s.pool.QueryRow(ctx, `
		SELECT manager_id, expire FROM managers_tokens WHERE token = $1
	`, token).Scan(&id, &expireTime)

	if err == pgx.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, ErrInternal
	}
	if time.Now().After(expireTime) {
		return 0, ErrTokenExpired
	}

	return id, nil
}

func (s *Service) Register(ctx context.Context, reg *Registration) (string, error) {
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
	_, err = s.pool.Exec(ctx, `INSERT INTO managers_tokens(token, manager_id) VALUES($1, $2)`, token, item.ID)
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

func (s *Service) SaveProduct(ctx context.Context, product *Product) (*Product, error) {
	if product.ID == 0 {
		err := s.pool.QueryRow(ctx, `
			INSERT INTO products (name, price, qty) VALUES ($1, $2, $3)
			RETURNING id, active, created
			`, product.Name, product.Price, product.Qty).Scan(&product.ID, &product.Active, &product.Created)
		if err != nil {
			return nil, ErrInternal
		}
		return product, nil
	}

	err := s.pool.QueryRow(ctx, `
		UPDATE  products SET name = $1, price = $2, qty = $3
		WHERE id = $4 RETURNING id, active, created`,
		product.Name, product.Price, product.Qty, product.ID).Scan(&product.ID, &product.Active, &product.Created)
	if err != nil {
		return nil, ErrInternal
	}
	return product, nil
}

func (s *Service) MakeSale(ctx context.Context, sale *Sale) (*Sale, error) {
	active := false
	qty := 0
	err := s.pool.QueryRow(ctx, `
		INSERT INTO sales (manager_id, customer_id) VALUES ($1, $2)
		RETURNING id, created
		`, sale.Manager_id, sale.Customer_id).Scan(&sale.ID, &sale.Created)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	positionsQuery := "INSERT INTO sales_positions(sale_id, product_id, qty, price) VALUES "
	for _, v := range sale.Positions {
		if err := s.pool.QueryRow(ctx, `SELECT qty, active from products where id = $1`, v.Product_id).Scan(&qty, &active); err != nil {
			return nil, ErrInternal
		}
		if qty < v.Qty || !active {
			return nil, ErrInternal
		}
		if _, err := s.pool.Exec(ctx, `UPDATE products set qty = $1 where id = $2`, qty-v.Qty, v.Product_id); err != nil {
			return nil, ErrInternal
		}
		positionsQuery += "(" + strconv.FormatInt(sale.ID, 10) + "," + strconv.FormatInt(v.Product_id, 10) + "," + strconv.Itoa(v.Qty) + "," + strconv.Itoa(v.Price) + "),"
	}
	positionsQuery = positionsQuery[:len(positionsQuery)-1] + ";"
	_, err = s.pool.Exec(ctx, positionsQuery)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return sale, nil
}

func (s *Service) GetSales(ctx context.Context, id int64) (total int, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(sp.price * sp.qty),0) total
		FROM managers m
		LEFT JOIN sales s on s.manager_id = $1
		LEFT JOIN sales_positions sp ON sp.sale_id = s.id
		GROUP BY m.id
	`, id).Scan(&total)
	if err != nil {
		log.Print(err)
		return 0, ErrInternal
	}
	return total, nil
}

func (s *Service) GetProducts(ctx context.Context) ([]*Product, error) {
	var products []*Product
	rows, err := s.pool.Query(ctx, `SELECT id, name, price, qty, active, created FROM products`)
	if err != nil {
		return nil, ErrInternal
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		product := &Product{}
		err = rows.Scan(&product.ID, &product.Name, &product.Price, &product.Qty, &product.Active, &product.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		products = append(products, product)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return products, nil
}

func (s *Service) RemoveProductByID(ctx context.Context, id int64) (*Product, error) {
	product := &Product{}
	err := s.pool.QueryRow(ctx, `
		DELETE FROM products WHERE id = $1 RETURNING id, name, price, qty
	`, id).Scan(&product.ID, &product.Name, &product.Price, &product.Qty)
	if err != nil {
		return nil, ErrInternal
	}

	return product, nil
}

func (s *Service) ChangeCustomer(ctx context.Context, item *customers.Customer) (*customers.Customer, error) {
	customer := &customers.Customer{}
	err := s.pool.QueryRow(ctx, `
		UPDATE customers SET name = $1, phone = $2 WHERE id = $3
		RETURNING id, name, phone, active, created
	`, item.Name, item.Phone, item.ID).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Active, &customer.Created)
	if err != nil {
		return nil, ErrInternal
	}
	return item, nil
}

func (s *Service) GetCustomers(ctx context.Context) ([]*customers.Customer, error) {
	var items []*customers.Customer
	rows, err := s.pool.Query(ctx, `SELECT id, name, phone, active, created FROM customers`)
	if err != nil {
		return nil, ErrInternal
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		customer := &customers.Customer{}
		err = rows.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Active, &customer.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, customer)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return items, nil
}

func (s *Service) RemoveCustomerByID(ctx context.Context, id int64) (*customers.Customer, error) {
	customer := &customers.Customer{}
	err := s.pool.QueryRow(ctx, `
		DELETE FROM customers WHERE id = $1 RETURNING id, phone
	`, id).Scan(&customer.ID, &customer.Phone)
	if err != nil {
		return nil, ErrInternal
	}

	return customer, nil
}

func (s *Service) IsAdmin(ctx context.Context, id int64) bool {
	err := s.pool.QueryRow(ctx, `
		select id from managers where 'ADMIN' =  any (roles) and id = $1
	`, id).Scan(&id)
	if err != nil {
		return false
	}
	return true
}
