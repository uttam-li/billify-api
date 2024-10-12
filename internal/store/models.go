package store

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Password  password  `json:"-"`
	Email     string    `json:"email"`
	CreatedAt string    `json:"created_at"`
}

type OAuthProvider struct {
	ID           string    `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Provider     string    `json:"provider"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type password struct {
	text *string
	hash []byte
}

type Business struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	GSTNo        string    `json:"gstno"`
	CompanyEmail string    `json:"company_email"`
	CompanyPhone string    `json:"company_phone"`
	Address      string    `json:"address"`
	City         string    `json:"city"`
	ZipCode      string    `json:"zip_code"`
	State        string    `json:"state"`
	Country      string    `json:"country"`
	BankName     string    `json:"bank_name"`
	AccountNo    string    `json:"account_no"`
	IFSC         string    `json:"ifsc"`
	BankBranch   string    `json:"bank_branch"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Invoice struct {
	ID          uuid.UUID  `json:"id"`
	InvNo       int64      `json:"inv_no"`
	BusID       uuid.UUID  `json:"bus_id"`
	CustID      uuid.UUID  `json:"cust_id"`
	TotalAmount float64    `json:"total_amount"`
	InvDate     time.Time  `json:"inv_date"`
	DueDate     time.Time  `json:"due_date"`
	IsPaid      bool       `json:"is_paid"`
	PaidDate    *time.Time `json:"paid_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type InvoiceItem struct {
	ID        uuid.UUID `json:"id"`
	InvID     uuid.UUID `json:"inv_id"`
	ProdID    uuid.UUID `json:"prod_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
}

type Customer struct {
	ID        uuid.UUID `json:"id"`
	BusID     uuid.UUID `json:"bus_id"`
	Name      string    `json:"name"`
	GSTNo     string    `json:"gstno"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	BAddress  string    `json:"b_address"`
	SAddress  string    `json:"s_address"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	ID        uuid.UUID `json:"id"`
	BusID     uuid.UUID `json:"bus_id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	TaxRate   float64   `json:"tax_rate"`
	Unit      string    `json:"unit"`
	HSNCode   string    `json:"hsn_code"`
	CreatedAt time.Time `json:"created_at"`
}

type Dashboard struct {
	TotalInvoices   int       `json:"total_invoices"`
	PendingInvoices int       `json:"pending_invoices"`
	TotalRevenue    float64   `json:"total_revenue"`
	UnpaidAmount    float64   `json:"unpaid_amount"`
	RecentInvoices  []Invoice `json:"recent_invoices"`
}

type CustomerWithPendingAmount struct {
    ID            uuid.UUID `json:"id"`
    BusID         uuid.UUID `json:"bus_id"`
    Name          string    `json:"name"`
    GSTNo         string    `json:"gstno"`
    Email         string    `json:"email"`
    Phone         string    `json:"phone"`
    BAddress      string    `json:"b_address"`
    SAddress      string    `json:"s_address"`
    CreatedAt     time.Time `json:"created_at"`
    PendingAmount float64   `json:"pending_amount"`
    TotalInvoices int       `json:"total_invoices"`
}

type InvoiceResponse struct {
    ID          uuid.UUID      `json:"id"`
    InvNo       int64          `json:"inv_no"`
    BusID       uuid.UUID      `json:"bus_id"`
    CustID      uuid.UUID      `json:"cust_id"`
    TotalAmount float64        `json:"total_amount"`
    InvDate     time.Time      `json:"inv_date"`
    DueDate     time.Time      `json:"due_date"`
    IsPaid      bool           `json:"is_paid"`
    PaidDate    *time.Time     `json:"paid_date,omitempty"`
    CreatedAt   time.Time      `json:"created_at"`
    Items       []*InvoiceItem  `json:"items"`
}