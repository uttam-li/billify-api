package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type InvoiceItemStore struct {
	db *sql.DB
}

func (s *InvoiceItemStore) Create(ctx context.Context, item *InvoiceItem) error {
	query := `
        INSERT INTO invoice_item (inv_id, prod_id, quantity, unit_price)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		item.InvID,
		item.ProdID,
		item.Quantity,
		item.UnitPrice,
	).Scan(
		&item.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *InvoiceItemStore) GetByID(ctx context.Context, itemID uuid.UUID) (*InvoiceItem, error) {
	query := `
        SELECT id, inv_id, prod_id, quantity, unit_price, created_at
        FROM invoice_item
        WHERE id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	item := &InvoiceItem{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		itemID,
	).Scan(
		&item.ID,
		&item.InvID,
		&item.ProdID,
		&item.Quantity,
		&item.UnitPrice,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return item, nil
}

func (s *InvoiceItemStore) GetByInvoiceID(ctx context.Context, invID uuid.UUID) ([]*InvoiceItem, error) {
	query := `
        SELECT id, inv_id, prod_id, quantity, unit_price
        FROM invoice_item
        WHERE inv_id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, invID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InvoiceItem
	for rows.Next() {
		item := &InvoiceItem{}
		err := rows.Scan(
			&item.ID,
			&item.InvID,
			&item.ProdID,
			&item.Quantity,
			&item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *InvoiceItemStore) Update(ctx context.Context, item *InvoiceItem) error {
	query := `
        UPDATE invoice_item
        SET inv_id = $2,
            prod_id = $3,
            quantity = $4,
            unit_price = $5
        WHERE id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.InvID,
		item.ProdID,
		item.Quantity,
		item.UnitPrice,
	)
	if err != nil {
		return err
	}

	return nil
}
func (s *InvoiceItemStore) UpdateAll(ctx context.Context, invoiceID uuid.UUID, items []*InvoiceItem) error {
	if len(items) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Delete all old items
	_, err = tx.ExecContext(ctx, `
        DELETE FROM invoice_item
        WHERE inv_id = $1
    `, invoiceID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert new items
	for _, item := range items {
		if item.ProdID == uuid.Nil {
			tx.Rollback()
			return fmt.Errorf("prod_id cannot be null")
		}

		_, err := tx.ExecContext(ctx, `
            INSERT INTO invoice_item (inv_id, prod_id, quantity, unit_price)
            VALUES ($1, $2, $3, $4)
        `, invoiceID, item.ProdID, item.Quantity, item.UnitPrice)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
