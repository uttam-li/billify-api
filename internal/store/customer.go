package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrDuplicateCustomer = errors.New("customer already exists")
)

type CustomerStore struct {
	db *sql.DB
}

func (s *CustomerStore) Create(ctx context.Context, customer *Customer) error {
	query := `
        INSERT INTO customer (buss_id, name, gstno, email, phone, baddress, saddress)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		customer.BusID,
		customer.Name,
		customer.GSTNo,
		customer.Email,
		customer.Phone,
		customer.BAddress,
		customer.SAddress,
	).Scan(
		&customer.ID,
		&customer.CreatedAt,
	)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return ErrNotFound
		case err.Error() == `pq: duplicate key value violates unique constraint "customer_buss_id_email_phone_key"`:
			return ErrDuplicateCustomer
		default:
			return err
		}
	}

	return nil
}

func (s *CustomerStore) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	query := `
		SELECT id, buss_id, name, gstno, email, phone, baddress, saddress, created_at
		FROM customer
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	customer := &Customer{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&customer.ID,
		&customer.BusID,
		&customer.Name,
		&customer.GSTNo,
		&customer.Email,
		&customer.Phone,
		&customer.BAddress,
		&customer.SAddress,
		&customer.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return customer, nil
}

func (s *CustomerStore) GetByBusID(ctx context.Context, busID uuid.UUID) ([]*CustomerWithPendingAmount, error) {
	query := `
        SELECT 
            c.id, 
            c.buss_id, 
            c.name, 
            c.gstno, 
            c.email, 
            c.phone, 
            c.baddress, 
            c.saddress, 
            c.created_at, 
            COALESCE(SUM(i.total_amount), 0) AS pending_amount,
            COUNT(i.id) AS total_invoices
        FROM customer c
        LEFT JOIN invoice i ON c.id = i.cust_id
        WHERE c.buss_id = $1
        GROUP BY c.id
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, busID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []*CustomerWithPendingAmount
	for rows.Next() {
		customer := &CustomerWithPendingAmount{}
		err := rows.Scan(
			&customer.ID,
			&customer.BusID,
			&customer.Name,
			&customer.GSTNo,
			&customer.Email,
			&customer.Phone,
			&customer.BAddress,
			&customer.SAddress,
			&customer.CreatedAt,
			&customer.PendingAmount,
			&customer.TotalInvoices,
		)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return customers, nil
}

func (s *CustomerStore) Update(ctx context.Context, customer *Customer) error {
	query := `
        UPDATE customer
        SET name = $2,
            gstno = $3,
            email = $4,
            phone = $5,
            baddress = $6,
            saddress = $7
        WHERE id = $1
        RETURNING created_at
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		customer.ID,
		customer.Name,
		customer.GSTNo,
		customer.Email,
		customer.Phone,
		customer.BAddress,
		customer.SAddress,
	).Scan(
		&customer.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *CustomerStore) Delete(ctx context.Context, customerID uuid.UUID) error {
	query := `
        DELETE FROM customer
        WHERE id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, customerID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
