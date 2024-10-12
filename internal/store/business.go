package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type BusinessStore struct {
	db *sql.DB
}

func (s *BusinessStore) Create(ctx context.Context, business *Business) error {
	query := `
        INSERT INTO business (user_id, name, gstno, company_email, company_phone, address, city, zip_code, state, country, bank_name, account_no, ifsc, bank_branch)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
        RETURNING buss_id, created_at
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		business.UserID,
		business.Name,
		business.GSTNo,
		business.CompanyEmail,
		business.CompanyPhone,
		business.Address,
		business.City,
		business.ZipCode,
		business.State,
		business.Country,
		business.BankName,
		business.AccountNo,
		business.IFSC,
		business.BankBranch,
	).Scan(
		&business.ID,
		&business.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *BusinessStore) GetByID(ctx context.Context, businessID uuid.UUID) (*Business, error) {
	query := `
        SELECT buss_id, user_id, name, gstno, company_email, company_phone, address, city, zip_code, state, country, bank_name, account_no, ifsc, bank_branch, created_at, updated_at
        FROM business
        WHERE buss_id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	business := &Business{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		businessID,
	).Scan(
		&business.ID,
		&business.UserID,
		&business.Name,
		&business.GSTNo,
		&business.CompanyEmail,
		&business.CompanyPhone,
		&business.Address,
		&business.City,
		&business.ZipCode,
		&business.State,
		&business.Country,
		&business.BankName,
		&business.AccountNo,
		&business.IFSC,
		&business.BankBranch,
		&business.CreatedAt,
		&business.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return business, nil
}

func (s *BusinessStore) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Business, error) {
	query := `
        SELECT buss_id, user_id, name, gstno
        FROM business
        WHERE user_id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businesses []*Business
	for rows.Next() {
		business := &Business{}
		err := rows.Scan(
			&business.ID,
			&business.UserID,
			&business.Name,
			&business.GSTNo,
		)
		if err != nil {
			return nil, err
		}
		businesses = append(businesses, business)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return businesses, nil
}

func (s *BusinessStore) Update(ctx context.Context, business *Business) error {
	query := `
        UPDATE business
        SET user_id = $2,
            name = $3,
            gstno = $4,
            company_email = $5,
            company_phone = $6,
            address = $7,
            city = $8,
            zip_code = $9,
            state = $10,
            country = $11,
            bank_name = $12,
            account_no = $13,
            ifsc = $14,
            bank_branch = $15,
            updated_at = $16
        WHERE buss_id = $1
        RETURNING created_at, updated_at
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		business.ID,
		business.UserID,
		business.Name,
		business.GSTNo,
		business.CompanyEmail,
		business.CompanyPhone,
		business.Address,
		business.City,
		business.ZipCode,
		business.State,
		business.Country,
		business.BankName,
		business.AccountNo,
		business.IFSC,
		business.BankBranch,
		time.Now(),
	).Scan(
		&business.CreatedAt,
		&business.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *BusinessStore) Delete(ctx context.Context, businessID uuid.UUID) error {
	query := `
		DELETE FROM business
		WHERE buss_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, businessID)
	if err != nil {
		return err
	}

	return nil
}

func (s *BusinessStore) GetDashboard(ctx context.Context, businessID uuid.UUID, from time.Time, to time.Time) (*Dashboard, error) {
    query := `
        WITH filtered_invoices AS (
            SELECT 
                is_paid, 
                total_amount 
            FROM invoice 
            WHERE buSs_id = $1 AND inv_date >= $2 AND inv_date <= $3
        ),
        recent_invoices AS (
            SELECT 
                inv_no,
                inv_date,
                total_amount,
                is_paid
            FROM invoice
            WHERE buss_id = $1
            ORDER BY inv_date DESC
            LIMIT 4
        )
        SELECT
            COUNT(*) AS total_invoices,
            COUNT(*) FILTER (WHERE is_paid = false) AS pending_invoices,
            SUM(total_amount) AS total_revenue,
            SUM(total_amount) FILTER (WHERE is_paid = false) AS unpaid_amount,
            (
                SELECT json_agg(recent_invoices)
                FROM recent_invoices
            ) AS recent_invoices
        FROM filtered_invoices
    `

    dashboard := &Dashboard{}
	
	var recentInvoicesJSON []byte
	var totalRevenue sql.NullFloat64
	var unpaidAmount sql.NullFloat64
    err := s.db.QueryRowContext(
        ctx,
        query,
        businessID,
        from,
        to,
    ).Scan(
        &dashboard.TotalInvoices,
        &dashboard.PendingInvoices,
        &totalRevenue,
        &unpaidAmount,
        &recentInvoicesJSON,
    )
    if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	if !totalRevenue.Valid {
        dashboard.TotalRevenue = 0
    }
    if !unpaidAmount.Valid {
        dashboard.UnpaidAmount = 0
    }

    var recentInvoices []Invoice
    if len(recentInvoicesJSON) > 0 {
        if err := json.Unmarshal(recentInvoicesJSON, &recentInvoices); err != nil {
            return nil, err
        }
    }
    dashboard.RecentInvoices = recentInvoices

    return dashboard, nil
}
