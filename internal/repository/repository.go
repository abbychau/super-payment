package repository

import (
	"database/sql"
	"fmt"
	"super-payment/internal/models"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Repository interface defines the contract for data access
type Repository interface {
	// User operations
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)

	// Company operations
	CreateCompany(company *models.Company) error
	GetCompanyByID(id uint) (*models.Company, error)

	// Business Partner operations
	CreateBusinessPartner(partner *models.BusinessPartner) error
	GetBusinessPartnerByID(id uint) (*models.BusinessPartner, error)
	GetBusinessPartnersByCompanyID(companyID uint) ([]*models.BusinessPartner, error)

	// Invoice operations
	CreateInvoice(invoice *models.Invoice) error
	GetInvoiceByID(id uint) (*models.Invoice, error)
	GetInvoicesByCompanyID(companyID uint, req *models.GetInvoicesRequest) ([]*models.Invoice, error)
	UpdateInvoiceStatus(id uint, status models.InvoiceStatus) error
}

// MySQLRepository implements Repository interface
type MySQLRepository struct {
	db *sql.DB
}

// NewMySQLRepository creates a new MySQL repository
func NewMySQLRepository(dsn string) (*MySQLRepository, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MySQLRepository{db: db}, nil
}

// Close closes the database connection
func (r *MySQLRepository) Close() error {
	return r.db.Close()
}

// CreateUser creates a new user
func (r *MySQLRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (company_id, full_name, email, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(query, user.CompanyID, user.FullName, user.Email, user.Password, now, now)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = uint(id)
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

// GetUserByEmail gets a user by email
func (r *MySQLRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT u.id, u.company_id, u.full_name, u.email, u.password, u.created_at, u.updated_at,
		       c.id, c.corporate_name, c.representative, c.phone_number, c.postal_code, c.address, c.created_at, c.updated_at
		FROM users u
		JOIN companies c ON u.company_id = c.id
		WHERE u.email = ?
	`
	row := r.db.QueryRow(query, email)

	user := &models.User{Company: &models.Company{}}
	err := row.Scan(
		&user.ID, &user.CompanyID, &user.FullName, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
		&user.Company.ID, &user.Company.CorporateName, &user.Company.Representative, &user.Company.PhoneNumber,
		&user.Company.PostalCode, &user.Company.Address, &user.Company.CreatedAt, &user.Company.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByID gets a user by ID
func (r *MySQLRepository) GetUserByID(id uint) (*models.User, error) {
	query := `
		SELECT u.id, u.company_id, u.full_name, u.email, u.password, u.created_at, u.updated_at,
		       c.id, c.corporate_name, c.representative, c.phone_number, c.postal_code, c.address, c.created_at, c.updated_at
		FROM users u
		JOIN companies c ON u.company_id = c.id
		WHERE u.id = ?
	`
	row := r.db.QueryRow(query, id)

	user := &models.User{Company: &models.Company{}}
	err := row.Scan(
		&user.ID, &user.CompanyID, &user.FullName, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
		&user.Company.ID, &user.Company.CorporateName, &user.Company.Representative, &user.Company.PhoneNumber,
		&user.Company.PostalCode, &user.Company.Address, &user.Company.CreatedAt, &user.Company.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateCompany creates a new company
func (r *MySQLRepository) CreateCompany(company *models.Company) error {
	query := `
		INSERT INTO companies (corporate_name, representative, phone_number, postal_code, address, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(query, company.CorporateName, company.Representative, company.PhoneNumber,
		company.PostalCode, company.Address, now, now)
	if err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	company.ID = uint(id)
	company.CreatedAt = now
	company.UpdatedAt = now
	return nil
}

// GetCompanyByID gets a company by ID
func (r *MySQLRepository) GetCompanyByID(id uint) (*models.Company, error) {
	query := `
		SELECT id, corporate_name, representative, phone_number, postal_code, address, created_at, updated_at
		FROM companies
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)

	company := &models.Company{}
	err := row.Scan(&company.ID, &company.CorporateName, &company.Representative, &company.PhoneNumber,
		&company.PostalCode, &company.Address, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("company not found")
		}
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return company, nil
}

// CreateBusinessPartner creates a new business partner
func (r *MySQLRepository) CreateBusinessPartner(partner *models.BusinessPartner) error {
	query := `
		INSERT INTO business_partners (company_id, corporate_name, representative, phone_number, postal_code, address, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(query, partner.CompanyID, partner.CorporateName, partner.Representative,
		partner.PhoneNumber, partner.PostalCode, partner.Address, now, now)
	if err != nil {
		return fmt.Errorf("failed to create business partner: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	partner.ID = uint(id)
	partner.CreatedAt = now
	partner.UpdatedAt = now
	return nil
}

// GetBusinessPartnerByID gets a business partner by ID
func (r *MySQLRepository) GetBusinessPartnerByID(id uint) (*models.BusinessPartner, error) {
	query := `
		SELECT id, company_id, corporate_name, representative, phone_number, postal_code, address, created_at, updated_at
		FROM business_partners
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)

	partner := &models.BusinessPartner{}
	err := row.Scan(&partner.ID, &partner.CompanyID, &partner.CorporateName, &partner.Representative,
		&partner.PhoneNumber, &partner.PostalCode, &partner.Address, &partner.CreatedAt, &partner.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("business partner not found")
		}
		return nil, fmt.Errorf("failed to get business partner: %w", err)
	}

	return partner, nil
}

// GetBusinessPartnersByCompanyID gets business partners by company ID
func (r *MySQLRepository) GetBusinessPartnersByCompanyID(companyID uint) ([]*models.BusinessPartner, error) {
	query := `
		SELECT id, company_id, corporate_name, representative, phone_number, postal_code, address, created_at, updated_at
		FROM business_partners
		WHERE company_id = ?
	`
	rows, err := r.db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business partners: %w", err)
	}
	defer rows.Close()

	var partners []*models.BusinessPartner
	for rows.Next() {
		partner := &models.BusinessPartner{}
		err := rows.Scan(&partner.ID, &partner.CompanyID, &partner.CorporateName, &partner.Representative,
			&partner.PhoneNumber, &partner.PostalCode, &partner.Address, &partner.CreatedAt, &partner.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan business partner: %w", err)
		}
		partners = append(partners, partner)
	}

	return partners, nil
}

// CreateInvoice creates a new invoice
func (r *MySQLRepository) CreateInvoice(invoice *models.Invoice) error {
	query := `
		INSERT INTO invoices (company_id, business_partner_id, issue_date, payment_amount, fee, fee_rate, 
		                     consumption_tax, consumption_tax_rate, invoice_amount, payment_due_date, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(query, invoice.CompanyID, invoice.BusinessPartnerID, invoice.IssueDate,
		invoice.PaymentAmount, invoice.Fee, invoice.FeeRate, invoice.ConsumptionTax, invoice.ConsumptionTaxRate,
		invoice.InvoiceAmount, invoice.PaymentDueDate, invoice.Status, now, now)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	invoice.ID = uint(id)
	invoice.CreatedAt = now
	invoice.UpdatedAt = now
	return nil
}

// GetInvoiceByID gets an invoice by ID
func (r *MySQLRepository) GetInvoiceByID(id uint) (*models.Invoice, error) {
	query := `
		SELECT i.id, i.company_id, i.business_partner_id, i.issue_date, i.payment_amount, i.fee, i.fee_rate,
		       i.consumption_tax, i.consumption_tax_rate, i.invoice_amount, i.payment_due_date, i.status, i.created_at, i.updated_at,
		       c.id, c.corporate_name, c.representative, c.phone_number, c.postal_code, c.address, c.created_at, c.updated_at,
		       bp.id, bp.company_id, bp.corporate_name, bp.representative, bp.phone_number, bp.postal_code, bp.address, bp.created_at, bp.updated_at
		FROM invoices i
		JOIN companies c ON i.company_id = c.id
		JOIN business_partners bp ON i.business_partner_id = bp.id
		WHERE i.id = ?
	`
	row := r.db.QueryRow(query, id)

	invoice := &models.Invoice{Company: &models.Company{}, BusinessPartner: &models.BusinessPartner{}}
	err := row.Scan(
		&invoice.ID, &invoice.CompanyID, &invoice.BusinessPartnerID, &invoice.IssueDate, &invoice.PaymentAmount,
		&invoice.Fee, &invoice.FeeRate, &invoice.ConsumptionTax, &invoice.ConsumptionTaxRate, &invoice.InvoiceAmount,
		&invoice.PaymentDueDate, &invoice.Status, &invoice.CreatedAt, &invoice.UpdatedAt,
		&invoice.Company.ID, &invoice.Company.CorporateName, &invoice.Company.Representative, &invoice.Company.PhoneNumber,
		&invoice.Company.PostalCode, &invoice.Company.Address, &invoice.Company.CreatedAt, &invoice.Company.UpdatedAt,
		&invoice.BusinessPartner.ID, &invoice.BusinessPartner.CompanyID, &invoice.BusinessPartner.CorporateName,
		&invoice.BusinessPartner.Representative, &invoice.BusinessPartner.PhoneNumber, &invoice.BusinessPartner.PostalCode,
		&invoice.BusinessPartner.Address, &invoice.BusinessPartner.CreatedAt, &invoice.BusinessPartner.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}

// GetInvoicesByCompanyID gets invoices by company ID with optional filters
func (r *MySQLRepository) GetInvoicesByCompanyID(companyID uint, req *models.GetInvoicesRequest) ([]*models.Invoice, error) {
	query := `
		SELECT i.id, i.company_id, i.business_partner_id, i.issue_date, i.payment_amount, i.fee, i.fee_rate,
		       i.consumption_tax, i.consumption_tax_rate, i.invoice_amount, i.payment_due_date, i.status, i.created_at, i.updated_at,
		       c.id, c.corporate_name, c.representative, c.phone_number, c.postal_code, c.address, c.created_at, c.updated_at,
		       bp.id, bp.company_id, bp.corporate_name, bp.representative, bp.phone_number, bp.postal_code, bp.address, bp.created_at, bp.updated_at
		FROM invoices i
		JOIN companies c ON i.company_id = c.id
		JOIN business_partners bp ON i.business_partner_id = bp.id
		WHERE i.company_id = ?
	`

	args := []interface{}{companyID}

	if req.StartDate != nil {
		query += " AND i.payment_due_date >= ?"
		args = append(args, *req.StartDate)
	}

	if req.EndDate != nil {
		query += " AND i.payment_due_date <= ?"
		args = append(args, *req.EndDate)
	}

	if req.Status != nil {
		query += " AND i.status = ?"
		args = append(args, *req.Status)
	}

	query += " ORDER BY i.payment_due_date DESC"

	if req.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, req.Limit)

		if req.Page > 1 {
			query += " OFFSET ?"
			args = append(args, (req.Page-1)*req.Limit)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{Company: &models.Company{}, BusinessPartner: &models.BusinessPartner{}}
		err := rows.Scan(
			&invoice.ID, &invoice.CompanyID, &invoice.BusinessPartnerID, &invoice.IssueDate, &invoice.PaymentAmount,
			&invoice.Fee, &invoice.FeeRate, &invoice.ConsumptionTax, &invoice.ConsumptionTaxRate, &invoice.InvoiceAmount,
			&invoice.PaymentDueDate, &invoice.Status, &invoice.CreatedAt, &invoice.UpdatedAt,
			&invoice.Company.ID, &invoice.Company.CorporateName, &invoice.Company.Representative, &invoice.Company.PhoneNumber,
			&invoice.Company.PostalCode, &invoice.Company.Address, &invoice.Company.CreatedAt, &invoice.Company.UpdatedAt,
			&invoice.BusinessPartner.ID, &invoice.BusinessPartner.CompanyID, &invoice.BusinessPartner.CorporateName,
			&invoice.BusinessPartner.Representative, &invoice.BusinessPartner.PhoneNumber, &invoice.BusinessPartner.PostalCode,
			&invoice.BusinessPartner.Address, &invoice.BusinessPartner.CreatedAt, &invoice.BusinessPartner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// UpdateInvoiceStatus updates the status of an invoice
func (r *MySQLRepository) UpdateInvoiceStatus(id uint, status models.InvoiceStatus) error {
	query := `UPDATE invoices SET status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}
	return nil
}
