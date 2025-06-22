package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"super-payment/internal/api"
	"super-payment/internal/config"
	"super-payment/internal/models"
	"super-payment/internal/repository"
	"super-payment/internal/service"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// APITestSuite defines the test suite
type APITestSuite struct {
	suite.Suite
	router      *gin.Engine
	authToken   string
	testUserID  uint
	testCompany models.Company
	testUser    models.User
}

// SetupSuite sets up the test suite
func (suite *APITestSuite) SetupSuite() {
	// Load test configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "3306",
			User:     "app_user",
			Password: "app_password",
			Name:     "super_payment",
		},
		JWT: config.JWTConfig{
			Secret:      "your-super-secret-jwt-key-change-in-production",
			ExpiryHours: 24,
		},
	}

	// Initialize repository (you might want to use a test database or mock)
	repo, err := repository.NewMySQLRepository(cfg.GetDSN())
	suite.NoError(err)

	// Initialize service
	svc := service.NewInvoiceService(repo)

	// Initialize handler
	handler := api.NewHandler(svc, cfg)

	// Setup router
	suite.router = handler.SetupRoutes()

	// Create a test user and company for authentication
	suite.createTestUser()
}

// createTestUser creates a test user for authentication in all tests
func (suite *APITestSuite) createTestUser() {
	// Generate unique email to avoid conflicts
	uniqueEmail := fmt.Sprintf("testuser%d@example.com", time.Now().UnixNano())

	registerData := map[string]interface{}{
		"company": map[string]interface{}{
			"corporate_name": "Test Company Inc.",
			"representative": "Test Representative",
			"phone_number":   "03-1234-5678",
			"postal_code":    "100-0001",
			"address":        "Tokyo, Test Address 1-1-1",
		},
		"user": map[string]interface{}{
			"full_name": "Test User",
			"email":     uniqueEmail,
			"password":  "password123",
		},
	}

	jsonData, _ := json.Marshal(registerData)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Ensure registration was successful
	suite.Equal(http.StatusCreated, w.Code)

	var response models.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.NotEmpty(response.Token)

	// Store for use in all tests
	suite.authToken = response.Token
	suite.testUserID = response.User.ID
	suite.testUser = response.User
	suite.testCompany = *response.User.Company
}

// TestHealthCheck tests the health check endpoint
func (suite *APITestSuite) TestHealthCheck() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "ok", response["status"])
}

// TestUserRegistration tests user registration with a different user
func (suite *APITestSuite) TestUserRegistration() {
	// Generate unique email to avoid conflicts with setup user
	uniqueEmail := fmt.Sprintf("newuser%d@example.com", time.Now().UnixNano())

	registerData := map[string]interface{}{
		"company": map[string]interface{}{
			"corporate_name": "New Test Company Inc.",
			"representative": "New Test Representative",
			"phone_number":   "03-9876-5432",
			"postal_code":    "101-0001",
			"address":        "Tokyo, New Test Address 2-2-2",
		},
		"user": map[string]interface{}{
			"full_name": "New Test User",
			"email":     uniqueEmail,
			"password":  "password123",
		},
	}

	jsonData, _ := json.Marshal(registerData)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), response.Token)
	assert.Equal(suite.T(), "New Test User", response.User.FullName)
	assert.Equal(suite.T(), uniqueEmail, response.User.Email)
}

// TestUserLogin tests user login with the test user created in setup
func (suite *APITestSuite) TestUserLogin() {
	loginData := models.LoginRequest{
		Email:    suite.testUser.Email,
		Password: "password123",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), response.Token)
}

// TestCreateBusinessPartner tests business partner creation
func (suite *APITestSuite) TestCreateBusinessPartner() {
	partnerData := models.BusinessPartnerCreateRequest{
		CorporateName:  "Test Partner Corp.",
		Representative: "Partner Representative",
		PhoneNumber:    "03-9876-5432",
		PostalCode:     "101-0001",
		Address:        "Tokyo, Partner Address 2-2-2",
	}

	jsonData, _ := json.Marshal(partnerData)
	req, _ := http.NewRequest("POST", "/api/business-partners", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Business partner created successfully", response.Message)
}

// TestCreateInvoice tests invoice creation
func (suite *APITestSuite) TestCreateInvoice() {
	// First create a business partner
	partnerData := models.BusinessPartnerCreateRequest{
		CorporateName:  "Invoice Test Partner",
		Representative: "Invoice Partner Rep",
		PhoneNumber:    "03-1111-1111",
		PostalCode:     "102-0001",
		Address:        "Tokyo, Invoice Partner Address 3-3-3",
	}

	jsonData, _ := json.Marshal(partnerData)
	req, _ := http.NewRequest("POST", "/api/business-partners", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var partnerResponse models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &partnerResponse)
	assert.NoError(suite.T(), err)

	// Extract business partner ID
	partnerMap := partnerResponse.Data.(map[string]interface{})
	businessPartnerID := uint(partnerMap["id"].(float64))

	// Now create an invoice
	invoiceData := models.CreateInvoiceRequest{
		BusinessPartnerID: businessPartnerID,
		PaymentAmount:     10000.00,
		PaymentDueDate:    time.Now().AddDate(0, 1, 0), // 1 month from now
	}

	jsonData, _ = json.Marshal(invoiceData)
	req, _ = http.NewRequest("POST", "/api/invoices", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.SuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invoice created successfully", response.Message)

	// Verify invoice calculations
	invoiceMap := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), 10000.00, invoiceMap["payment_amount"])
	assert.Equal(suite.T(), 400.00, invoiceMap["fee"])              // 10000 * 0.04
	assert.Equal(suite.T(), 40.00, invoiceMap["consumption_tax"])   // 400 * 0.10
	assert.Equal(suite.T(), 10440.00, invoiceMap["invoice_amount"]) // 10000 + 400 + 40
}

// TestGetInvoices tests invoice retrieval
func (suite *APITestSuite) TestGetInvoices() {
	req, _ := http.NewRequest("GET", "/api/invoices", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invoices retrieved successfully", response.Message)
}

// TestGetInvoicesWithDateFilter tests invoice retrieval with date filters
func (suite *APITestSuite) TestGetInvoicesWithDateFilter() {
	startDate := time.Now().Format(time.RFC3339)
	endDate := time.Now().AddDate(0, 2, 0).Format(time.RFC3339)

	url := fmt.Sprintf("/api/invoices?start_date=%s&end_date=%s",
		url.QueryEscape(startDate), url.QueryEscape(endDate))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invoices retrieved successfully", response.Message)
}

// TestUnauthorizedAccess tests accessing protected endpoints without token
func (suite *APITestSuite) TestUnauthorizedAccess() {
	req, _ := http.NewRequest("GET", "/api/invoices", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "unauthorized", response.Error)
}

// TestInvalidTokenAccess tests accessing protected endpoints with invalid token
func (suite *APITestSuite) TestInvalidTokenAccess() {
	req, _ := http.NewRequest("GET", "/api/invoices", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "unauthorized", response.Error)
}

// TestInvalidLoginCredentials tests login with invalid credentials
func (suite *APITestSuite) TestInvalidLoginCredentials() {
	loginData := models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "authentication_failed", response.Error)
}

// TestInvoiceCalculationAccuracy tests the accuracy of invoice calculations
func (suite *APITestSuite) TestInvoiceCalculationAccuracy() {
	testCases := []struct {
		paymentAmount   float64
		expectedFee     float64
		expectedTax     float64
		expectedInvoice float64
	}{
		{10000.00, 400.00, 40.00, 10440.00},
		{50000.00, 2000.00, 200.00, 52200.00},
		{25000.00, 1000.00, 100.00, 26100.00},
		{1000.00, 40.00, 4.00, 1044.00},
	}
	for _, tc := range testCases {
		// Create business partner for this test
		partnerData := models.BusinessPartnerCreateRequest{
			CorporateName:  fmt.Sprintf("Calc Test Partner %v", tc.paymentAmount),
			Representative: "Calc Test Rep",
			PhoneNumber:    "03-2222-2222",
			PostalCode:     "103-0001",
			Address:        "Tokyo, Calc Test Address 4-4-4",
		}

		jsonData, _ := json.Marshal(partnerData)
		req, _ := http.NewRequest("POST", "/api/business-partners", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		var partnerResponse models.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &partnerResponse)
		assert.NoError(suite.T(), err)

		partnerMap := partnerResponse.Data.(map[string]interface{})
		businessPartnerID := uint(partnerMap["id"].(float64))

		// Create invoice
		invoiceData := models.CreateInvoiceRequest{
			BusinessPartnerID: businessPartnerID,
			PaymentAmount:     tc.paymentAmount,
			PaymentDueDate:    time.Now().AddDate(0, 1, 0),
		}

		jsonData, _ = json.Marshal(invoiceData)
		req, _ = http.NewRequest("POST", "/api/invoices", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)

		w = httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response models.SuccessResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)

		invoiceMap := response.Data.(map[string]interface{})
		assert.Equal(suite.T(), tc.paymentAmount, invoiceMap["payment_amount"])
		assert.Equal(suite.T(), tc.expectedFee, invoiceMap["fee"])
		assert.Equal(suite.T(), tc.expectedTax, invoiceMap["consumption_tax"])
		assert.Equal(suite.T(), tc.expectedInvoice, invoiceMap["invoice_amount"])
	}
}

// TestSuite runs the test suite
func TestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
