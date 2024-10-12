package store

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type ProductStore struct {
	db *sql.DB
}

func (s *ProductStore) Create(ctx context.Context, product *Product) error {
	query := `
        INSERT INTO product (buss_id, name, price, tax_rate, unit, hsn_code)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		product.BusID,
		product.Name,
		product.Price,
		product.TaxRate,
		product.Unit,
		product.HSNCode,
	).Scan(
		&product.ID,
		&product.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *ProductStore) GetByBusID(ctx context.Context, busID uuid.UUID) ([]*Product, error) {
	query := `
        SELECT id, buss_id, name, price, tax_rate, unit, hsn_code, created_at
        FROM product
        WHERE buss_id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, busID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		product := &Product{}
		err := rows.Scan(
			&product.ID,
			&product.BusID,
			&product.Name,
			&product.Price,
			&product.TaxRate,
			&product.Unit,
			&product.HSNCode,
			&product.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ProductStore) Update(ctx context.Context, product *Product) error {
	query := `
        UPDATE product
        SET name = $2,
            price = $3,
            tax_rate = $4,
            unit = $5,
            hsn_code = $6
        WHERE id = $1
        RETURNING created_at
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		product.ID,
		product.Name,
		product.Price,
		product.TaxRate,
		product.Unit,
		product.HSNCode,
	).Scan(
		&product.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *ProductStore) Delete(ctx context.Context, productID uuid.UUID) error {
	query := `
        DELETE FROM product
        WHERE id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, productID)
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
