package models

import (
	"fmt"
	"regexp"
	"time"
)

// Company represents a company entity
type Company struct {
	ID             uint      `json:"id" db:"id"`
	CorporateName  string    `json:"corporate_name" db:"corporate_name" binding:"required"`
	Representative string    `json:"representative" db:"representative" binding:"required"`
	PhoneNumber    string    `json:"phone_number" db:"phone_number" binding:"required"`
	PostalCode     string    `json:"postal_code" db:"postal_code" binding:"required"`
	Address        string    `json:"address" db:"address" binding:"required"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// User represents a user entity linked to a company
type User struct {
	ID        uint      `json:"id" db:"id"`
	CompanyID uint      `json:"company_id" db:"company_id" binding:"required"`
	FullName  string    `json:"full_name" db:"full_name" binding:"required"`
	Email     string    `json:"email" db:"email" binding:"required,email"`
	Password  string    `json:"-" db:"password" binding:"required,min=8"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Company   *Company  `json:"company,omitempty"`
}

// BusinessPartner represents a business partner entity linked to a company
type BusinessPartner struct {
	ID             uint      `json:"id" db:"id"`
	CompanyID      uint      `json:"company_id" db:"company_id" binding:"required"`
	CorporateName  string    `json:"corporate_name" db:"corporate_name" binding:"required"`
	Representative string    `json:"representative" db:"representative" binding:"required"`
	PhoneNumber    string    `json:"phone_number" db:"phone_number" binding:"required"`
	PostalCode     string    `json:"postal_code" db:"postal_code" binding:"required"`
	Address        string    `json:"address" db:"address" binding:"required"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// BusinessPartnerBankAccount represents bank account information for a business partner
type BusinessPartnerBankAccount struct {
	ID                uint      `json:"id" db:"id"`
	BusinessPartnerID uint      `json:"business_partner_id" db:"business_partner_id" binding:"required"`
	BankName          string    `json:"bank_name" db:"bank_name" binding:"required"`
	BranchName        string    `json:"branch_name" db:"branch_name" binding:"required"`
	AccountNumber     string    `json:"account_number" db:"account_number" binding:"required"`
	AccountName       string    `json:"account_name" db:"account_name" binding:"required"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusUnprocessed InvoiceStatus = "unprocessed"
	InvoiceStatusProcessing  InvoiceStatus = "processing"
	InvoiceStatusPaid        InvoiceStatus = "paid"
	InvoiceStatusError       InvoiceStatus = "error"
)

// Invoice represents invoice data linked to a company and business partner
type Invoice struct {
	ID                 uint             `json:"id" db:"id"`
	CompanyID          uint             `json:"company_id" db:"company_id" binding:"required"`
	BusinessPartnerID  uint             `json:"business_partner_id" db:"business_partner_id" binding:"required"`
	IssueDate          time.Time        `json:"issue_date" db:"issue_date" binding:"required"`
	PaymentAmount      float64          `json:"payment_amount" db:"payment_amount" binding:"required,gt=0"`
	Fee                float64          `json:"fee" db:"fee"`
	FeeRate            float64          `json:"fee_rate" db:"fee_rate"`
	ConsumptionTax     float64          `json:"consumption_tax" db:"consumption_tax"`
	ConsumptionTaxRate float64          `json:"consumption_tax_rate" db:"consumption_tax_rate"`
	InvoiceAmount      float64          `json:"invoice_amount" db:"invoice_amount"`
	PaymentDueDate     time.Time        `json:"payment_due_date" db:"payment_due_date" binding:"required"`
	Status             InvoiceStatus    `json:"status" db:"status"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
	Company            *Company         `json:"company,omitempty"`
	BusinessPartner    *BusinessPartner `json:"business_partner,omitempty"`
}

// CreateInvoiceRequest represents the request structure for creating an invoice
type CreateInvoiceRequest struct {
	BusinessPartnerID uint      `json:"business_partner_id" binding:"required"`
	PaymentAmount     float64   `json:"payment_amount" binding:"required,gt=0"`
	PaymentDueDate    time.Time `json:"payment_due_date" binding:"required"`
}

// GetInvoicesRequest represents the query parameters for retrieving invoices
type GetInvoicesRequest struct {
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
	Status    *string    `form:"status"`
	Page      int        `form:"page,default=1"`
	Limit     int        `form:"limit,default=20"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// UserRegistrationRequest represents the request structure for user registration
type UserRegistrationRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// BusinessPartnerCreateRequest represents the request structure for creating a business partner
type BusinessPartnerCreateRequest struct {
	CorporateName  string `json:"corporate_name" binding:"required"`
	Representative string `json:"representative" binding:"required"`
	PhoneNumber    string `json:"phone_number" binding:"required"`
	PostalCode     string `json:"postal_code" binding:"required"`
	Address        string `json:"address" binding:"required"`
}

// ToBusinessPartner converts the request to a BusinessPartner model
func (req *BusinessPartnerCreateRequest) ToBusinessPartner() *BusinessPartner {
	return &BusinessPartner{
		CorporateName:  req.CorporateName,
		Representative: req.Representative,
		PhoneNumber:    req.PhoneNumber,
		PostalCode:     req.PostalCode,
		Address:        req.Address,
	}
}

// Validation functions
var (
	// Japanese phone number pattern: XXX-XXXX-XXXX format
	phoneRegex = regexp.MustCompile(`^0\d{1,4}-\d{1,4}-\d{4}$`)
	// Japanese postal code pattern: XXX-XXXX format
	postalCodeRegex = regexp.MustCompile(`^\d{3}-\d{4}$`)
)

// ValidatePhoneNumber validates Japanese phone number format
func ValidatePhoneNumber(phone string) error {
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone number format. Expected format: XXX-XXXX-XXXX")
	}
	return nil
}

// ValidatePostalCode validates Japanese postal code format
func ValidatePostalCode(postalCode string) error {
	if !postalCodeRegex.MatchString(postalCode) {
		return fmt.Errorf("invalid postal code format. Expected format: XXX-XXXX")
	}
	return nil
}

// ValidatePaymentDueDate validates that the payment due date is in the future
func ValidatePaymentDueDate(dueDate time.Time) error {
	if dueDate.Before(time.Now()) {
		return fmt.Errorf("payment due date must be in the future")
	}
	return nil
}

// Validate validates the BusinessPartnerCreateRequest
func (req *BusinessPartnerCreateRequest) Validate() error {
	if err := ValidatePhoneNumber(req.PhoneNumber); err != nil {
		return err
	}
	if err := ValidatePostalCode(req.PostalCode); err != nil {
		return err
	}
	return nil
}

// Validate validates the CreateInvoiceRequest
func (req *CreateInvoiceRequest) Validate() error {
	if err := ValidatePaymentDueDate(req.PaymentDueDate); err != nil {
		return err
	}
	return nil
}
