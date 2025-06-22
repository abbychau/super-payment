package api

import (
	"fmt"
	"net/http"
	"strconv"
	"super-payment/internal/config"
	"super-payment/internal/middleware"
	"super-payment/internal/models"
	"super-payment/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler holds the HTTP handlers
type Handler struct {
	service service.Service
	config  *config.Config
}

// NewHandler creates a new HTTP handler
func NewHandler(service service.Service, config *config.Config) *Handler {
	return &Handler{
		service: service,
		config:  config,
	}
}

// SetupRoutes sets up the HTTP routes
func (h *Handler) SetupRoutes() *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.ErrorHandlingMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", h.healthCheck)

	// Public routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", h.register)
		auth.POST("/login", h.login)
	}

	// Protected routes
	api := router.Group("/api")
	api.Use(middleware.JWTMiddleware(h.config))
	{
		// Invoice routes
		api.POST("/invoices", h.createInvoice)
		api.GET("/invoices", h.getInvoices)
		api.GET("/invoices/:id", h.getInvoiceByID)

		// Business partner routes
		api.POST("/business-partners", h.createBusinessPartner)
		api.GET("/business-partners", h.getBusinessPartners)

		// Company routes
		api.POST("/companies", h.createCompany)
	}

	return router
}

// healthCheck handles health check requests
func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"service":   "super-payment-api",
	})
}

// register handles user registration
func (h *Handler) register(c *gin.Context) {
	var req struct {
		Company models.Company                 `json:"company" binding:"required"`
		User    models.UserRegistrationRequest `json:"user" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Create company first
	if err := h.service.CreateCompany(&req.Company); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "company_creation_failed",
			Message: err.Error(),
		})
		return
	}

	// Create user from registration request
	user := models.User{
		CompanyID: req.Company.ID,
		FullName:  req.User.FullName,
		Email:     req.User.Email,
		Password:  req.User.Password,
	}

	// Create user
	if err := h.service.RegisterUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "user_registration_failed",
			Message: err.Error(),
		})
		return
	}

	// Generate JWT token
	user.Company = &req.Company
	token, err := middleware.GenerateJWT(&user, h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "token_generation_failed",
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// login handles user login
func (h *Handler) login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	user, err := h.service.LoginUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "authentication_failed",
			Message: "Invalid email or password",
		})
		return
	}

	token, err := middleware.GenerateJWT(user, h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "token_generation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

// createInvoice handles invoice creation
func (h *Handler) createInvoice(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	var req models.CreateInvoiceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Additional validation
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	invoice, err := h.service.CreateInvoice(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "invoice_creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Invoice created successfully",
		Data:    invoice,
	})
}

// getInvoices handles invoice retrieval with filters
func (h *Handler) getInvoices(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	var req models.GetInvoicesRequest

	// Parse query parameters manually for better control
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "validation_error",
				Message: fmt.Sprintf("Invalid start_date format: %v", err),
			})
			return
		} else {
			req.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "validation_error",
				Message: fmt.Sprintf("Invalid end_date format: %v", err),
			})
			return
		} else {
			req.EndDate = &endDate
		}
	}

	if status := c.Query("status"); status != "" {
		req.Status = &status
	}

	// Parse pagination parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	} else {
		req.Page = 1
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	} else {
		req.Limit = 20
	}

	invoices, err := h.service.GetInvoices(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "invoice_retrieval_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Invoices retrieved successfully",
		Data:    invoices,
	})
}

// getInvoiceByID handles single invoice retrieval
func (h *Handler) getInvoiceByID(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	idStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid invoice ID",
		})
		return
	}

	invoice, err := h.service.GetInvoiceByID(userID, uint(invoiceID))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "invoice_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Invoice retrieved successfully",
		Data:    invoice,
	})
}

// createBusinessPartner handles business partner creation
func (h *Handler) createBusinessPartner(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	var req models.BusinessPartnerCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Additional validation
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	partner := req.ToBusinessPartner()

	if err := h.service.CreateBusinessPartner(userID, partner); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "business_partner_creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Business partner created successfully",
		Data:    partner,
	})
}

// getBusinessPartners handles business partner retrieval
func (h *Handler) getBusinessPartners(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	partners, err := h.service.GetBusinessPartners(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "business_partner_retrieval_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Business partners retrieved successfully",
		Data:    partners,
	})
}

// createCompany handles company creation (for admin use)
func (h *Handler) createCompany(c *gin.Context) {
	var company models.Company

	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	if err := h.service.CreateCompany(&company); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "company_creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Company created successfully",
		Data:    company,
	})
}
