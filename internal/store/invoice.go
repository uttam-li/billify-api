package store

import (
    "context"
    "database/sql"
    "errors"

    "github.com/google/uuid"
    "github.com/lib/pq"
)

var (
    ErrDuplicateInvoice = errors.New("invoice already exists for this customer on this date")
)

type InvoiceStore struct {
    db *sql.DB
}

func isUniqueViolation(err error) bool {
    if pqErr, ok := err.(*pq.Error); ok {
        return pqErr.Code == "23505"
    }
    return false
}

func (s *InvoiceStore) Create(ctx context.Context, invoice *Invoice) error {
    query := `
        INSERT INTO invoice (inv_no, buss_id, cust_id, total_amount, inv_date, due_date, is_paid, paid_date)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at
    `

    ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
    defer cancel()

    err := s.db.QueryRowContext(
        ctx,
        query,
        invoice.InvNo,
        invoice.BusID,
        invoice.CustID,
        invoice.TotalAmount,
        invoice.InvDate,
        invoice.DueDate,
        invoice.IsPaid,
        invoice.PaidDate,
    ).Scan(
        &invoice.ID,
        &invoice.CreatedAt,
    )
    if err != nil {
        if isUniqueViolation(err) {
            return ErrDuplicateInvoice
        }
        return err
    }

    return nil
}

func (s *InvoiceStore) GetNextInvoiceNumber(ctx context.Context, busID uuid.UUID) (int64, error) {
    query := `
        SELECT COALESCE(MAX(inv_no), 0) + 1
        FROM invoice
        WHERE buss_id = $1
    `

    ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
    defer cancel()

    var nextInvNo int64
    err := s.db.QueryRowContext(ctx, query, busID).Scan(&nextInvNo)
    if err != nil {
        return 0, err
    }

    return nextInvNo, nil
}

func (s *InvoiceStore) GetByID(ctx context.Context, invoiceID uuid.UUID) (*Invoice, error) {
    query := `
        SELECT id, inv_no, buss_id, cust_id, total_amount, inv_date, due_date, is_paid, paid_date, created_at
        FROM invoice
        WHERE id = $1
    `

    ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
    defer cancel()

    invoice := &Invoice{}
    err := s.db.QueryRowContext(
        ctx,
        query,
        invoiceID,
    ).Scan(
        &invoice.ID,
        &invoice.InvNo,
        &invoice.BusID,
        &invoice.CustID,
        &invoice.TotalAmount,
        &invoice.InvDate,
        &invoice.DueDate,
        &invoice.IsPaid,
        &invoice.PaidDate,
        &invoice.CreatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrNotFound
        }
        return nil, err
    }

    return invoice, nil
}

func (s *InvoiceStore) GetByBusID(ctx context.Context, busID uuid.UUID) ([]*Invoice, error) {
    query := `
        SELECT id, inv_no, buss_id, cust_id, total_amount, inv_date, due_date, is_paid, paid_date, created_at
        FROM invoice
        WHERE buss_id = $1
    `

    ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
    defer cancel()

    rows, err := s.db.QueryContext(ctx, query, busID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var invoices []*Invoice
    for rows.Next() {
        invoice := &Invoice{}
        err := rows.Scan(
            &invoice.ID,
            &invoice.InvNo,
            &invoice.BusID,
            &invoice.CustID,
            &invoice.TotalAmount,
            &invoice.InvDate,
            &invoice.DueDate,
            &invoice.IsPaid,
            &invoice.PaidDate,
            &invoice.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        invoices = append(invoices, invoice)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return invoices, nil
}

func (s *InvoiceStore) Update(ctx context.Context, invoice *Invoice) error {
    query := `
        UPDATE invoice
        SET cust_id = $2,
            total_amount = $3,
            inv_date = $4,
            due_date = $5,
            is_paid = $6,
            paid_date = $7
        WHERE id = $1
        RETURNING created_at
    `

    ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
    defer cancel()

    result, err := s.db.ExecContext(
        ctx,
        query,
        invoice.ID,
        invoice.CustID,
        invoice.TotalAmount,
        invoice.InvDate,
        invoice.DueDate,
        invoice.IsPaid,
        invoice.PaidDate,
    )
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("failed to update invoice")
    }

    return nil
}

func (s *InvoiceStore) UpdateStatus(ctx context.Context, invoiceID uuid.UUID) error {
    query := `
        UPDATE invoice
        SET is_paid = NOT is_paid,
            paid_date = CASE
                WHEN NOT is_paid THEN NOW()
                ELSE NULL
            END
        WHERE id = $1
    `

    ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
    defer cancel()

    result, err := s.db.ExecContext(ctx, query, invoiceID)
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

func (s *InvoiceStore) Delete(ctx context.Context, invoiceID uuid.UUID) error {
    query := `
        DELETE FROM invoice
        WHERE id = $1
    `

    ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
    defer cancel()

    result, err := s.db.ExecContext(ctx, query, invoiceID)
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

func (s *InvoiceItemStore) Delete(ctx context.Context, itemID uuid.UUID) error {
    query := `
        DELETE FROM invoice_item
        WHERE id = $1
    `

    ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
    defer cancel()

    result, err := s.db.ExecContext(ctx, query, itemID)
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