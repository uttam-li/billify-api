package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("user not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Users interface {
		GetByID(context.Context, uuid.UUID) (*User, error)
		GetByEmail(context.Context, string) (*User, error)
		Create(context.Context, *User) error
		Update(context.Context, *User) error
		Delete(context.Context, uuid.UUID) error
	}
	OAuthProvider interface {
		CreateOrUpdate(context.Context, *OAuthProvider) error
	}
	Business interface {
		Create(context.Context, *Business) error
		GetByID(context.Context, uuid.UUID) (*Business, error)
		GetByUserID(context.Context, uuid.UUID) ([]*Business, error)
		GetDashboard(context.Context, uuid.UUID, time.Time, time.Time) (*Dashboard, error)
		Update(context.Context, *Business) error
		Delete(context.Context, uuid.UUID) error
	}
	Invoices interface {
		GetByID(context.Context, uuid.UUID) (*Invoice, error)
		Create(context.Context, *Invoice) error
		Update(context.Context, *Invoice) error
		UpdateStatus(context.Context, uuid.UUID) error
		Delete(context.Context, uuid.UUID) error
		GetByBusID(context.Context, uuid.UUID) ([]*Invoice, error)
		GetNextInvoiceNumber(context.Context, uuid.UUID) (int64, error)
	}
	InvoiceItems interface {
		GetByID(context.Context, uuid.UUID) (*InvoiceItem, error)
		GetByInvoiceID(context.Context, uuid.UUID) ([]*InvoiceItem, error)
		Create(context.Context, *InvoiceItem) error
		Update(context.Context, *InvoiceItem) error
		UpdateAll(context.Context, uuid.UUID, []*InvoiceItem) error
		Delete(context.Context, uuid.UUID) error
	}
	Customers interface {
		Create(context.Context, *Customer) error
		GetByID(context.Context, uuid.UUID) (*Customer, error)
		GetByBusID(context.Context, uuid.UUID) ([]*CustomerWithPendingAmount, error)
		Update(context.Context, *Customer) error
		Delete(context.Context, uuid.UUID) error
	}
	Products interface {
		GetByBusID(context.Context, uuid.UUID) ([]*Product, error)
		Create(context.Context, *Product) error
		Update(context.Context, *Product) error
		Delete(context.Context, uuid.UUID) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users:         &UserStore{db},
		OAuthProvider: &OAuthProviderStore{db},
		Business:      &BusinessStore{db},
		Invoices:      &InvoiceStore{db},
		InvoiceItems:  &InvoiceItemStore{db},
		Customers:     &CustomerStore{db},
		Products:      &ProductStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
