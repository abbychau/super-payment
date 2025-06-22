package service

import (
	"fmt"
	"math"
	"super-payment/internal/models"
	"super-payment/internal/repository"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Service interface defines the business logic contract
type Service interface {
	// Authentication
	RegisterUser(user *models.User) error
	LoginUser(email, password string) (*models.User, error)

	// Invoice operations
	CreateInvoice(userID uint, req *models.CreateInvoiceRequest) (*models.Invoice, error)
	GetInvoices(userID uint, req *models.GetInvoicesRequest) ([]*models.Invoice, error)
	GetInvoiceByID(userID uint, invoiceID uint) (*models.Invoice, error)

	// Company operations
	CreateCompany(company *models.Company) error

	// Business Partner operations
	CreateBusinessPartner(userID uint, partner *models.BusinessPartner) error
	GetBusinessPartners(userID uint) ([]*models.BusinessPartner, error)
}

// InvoiceService implements Service interface
type InvoiceService struct {
	repo repository.Repository
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(repo repository.Repository) *InvoiceService {
	return &InvoiceService{repo: repo}
}

// RegisterUser registers a new user
func (s *InvoiceService) RegisterUser(user *models.User) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Create user
	if err := s.repo.CreateUser(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// LoginUser authenticates a user
func (s *InvoiceService) LoginUser(email, password string) (*models.User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Clear password from response
	user.Password = ""
	return user, nil
}

// CreateInvoice creates a new invoice with automatic calculations
func (s *InvoiceService) CreateInvoice(userID uint, req *models.CreateInvoiceRequest) (*models.Invoice, error) {
	// Get user to get company ID
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify business partner belongs to the same company
	partner, err := s.repo.GetBusinessPartnerByID(req.BusinessPartnerID)
	if err != nil {
		return nil, fmt.Errorf("business partner not found: %w", err)
	}

	if partner.CompanyID != user.CompanyID {
		return nil, fmt.Errorf("business partner does not belong to your company")
	}

	// Calculate invoice amounts
	invoice := &models.Invoice{
		CompanyID:          user.CompanyID,
		BusinessPartnerID:  req.BusinessPartnerID,
		IssueDate:          time.Now(),
		PaymentAmount:      req.PaymentAmount,
		FeeRate:            0.04, // 4% fee rate
		ConsumptionTaxRate: 0.10, // 10% consumption tax rate
		PaymentDueDate:     req.PaymentDueDate,
		Status:             models.InvoiceStatusUnprocessed,
	}

	// Calculate fee: payment amount * 4%
	invoice.Fee = invoice.PaymentAmount * invoice.FeeRate

	// Calculate consumption tax: fee * 10%
	invoice.ConsumptionTax = invoice.Fee * invoice.ConsumptionTaxRate

	// Calculate invoice amount: payment amount + fee + consumption tax
	// Round to 2 decimal places
	invoice.InvoiceAmount = math.Round((invoice.PaymentAmount+invoice.Fee+invoice.ConsumptionTax)*100) / 100

	// Create invoice
	if err := s.repo.CreateInvoice(invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Get the created invoice with related data
	createdInvoice, err := s.repo.GetInvoiceByID(invoice.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created invoice: %w", err)
	}

	return createdInvoice, nil
}

// GetInvoices retrieves invoices for a user's company with optional filters
func (s *InvoiceService) GetInvoices(userID uint, req *models.GetInvoicesRequest) ([]*models.Invoice, error) {
	// Get user to get company ID
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Set default pagination if not provided
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100 // Maximum limit
	}

	// Get invoices
	invoices, err := s.repo.GetInvoicesByCompanyID(user.CompanyID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	return invoices, nil
}

// GetInvoiceByID retrieves a specific invoice by ID
func (s *InvoiceService) GetInvoiceByID(userID uint, invoiceID uint) (*models.Invoice, error) {
	// Get user to get company ID
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get invoice
	invoice, err := s.repo.GetInvoiceByID(invoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Verify invoice belongs to user's company
	if invoice.CompanyID != user.CompanyID {
		return nil, fmt.Errorf("invoice not found")
	}

	return invoice, nil
}

// CreateCompany creates a new company
func (s *InvoiceService) CreateCompany(company *models.Company) error {
	if err := s.repo.CreateCompany(company); err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}
	return nil
}

// CreateBusinessPartner creates a new business partner
func (s *InvoiceService) CreateBusinessPartner(userID uint, partner *models.BusinessPartner) error {
	// Get user to get company ID
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	partner.CompanyID = user.CompanyID

	if err := s.repo.CreateBusinessPartner(partner); err != nil {
		return fmt.Errorf("failed to create business partner: %w", err)
	}

	return nil
}

// GetBusinessPartners retrieves business partners for a user's company
func (s *InvoiceService) GetBusinessPartners(userID uint) ([]*models.BusinessPartner, error) {
	// Get user to get company ID
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	partners, err := s.repo.GetBusinessPartnersByCompanyID(user.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business partners: %w", err)
	}

	return partners, nil
}
