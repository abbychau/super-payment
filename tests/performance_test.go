package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"super-payment/internal/models"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentInvoiceCreation tests concurrent invoice creation
func (suite *APITestSuite) TestConcurrentInvoiceCreation() {
	// Create a business partner for concurrent tests
	partnerData := models.BusinessPartnerCreateRequest{
		CorporateName:  "Concurrent Test Partner",
		Representative: "Concurrent Test Rep",
		PhoneNumber:    "03-4444-4444",
		PostalCode:     "105-0001",
		Address:        "Tokyo, Concurrent Test Address 6-6-6",
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

	// Test concurrent invoice creation
	concurrentRequests := 10
	var wg sync.WaitGroup
	var mutex sync.Mutex
	results := make([]int, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			invoiceData := models.CreateInvoiceRequest{
				BusinessPartnerID: businessPartnerID,
				PaymentAmount:     float64(10000 + index*1000), // Different amounts
				PaymentDueDate:    time.Now().AddDate(0, 1, index),
			}

			jsonData, _ := json.Marshal(invoiceData)
			req, _ := http.NewRequest("POST", "/api/invoices", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			mutex.Lock()
			results[index] = w.Code
			mutex.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all requests succeeded
	successCount := 0
	for _, code := range results {
		if code == http.StatusOK {
			successCount++
		}
	}

	assert.Equal(suite.T(), concurrentRequests, successCount, "All concurrent invoice creations should succeed")
}

// TestConcurrentInvoiceRetrieval tests concurrent invoice retrieval
func (suite *APITestSuite) TestConcurrentInvoiceRetrieval() {
	concurrentRequests := 20
	var wg sync.WaitGroup
	var mutex sync.Mutex
	results := make([]int, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			req, _ := http.NewRequest("GET", "/api/invoices", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			mutex.Lock()
			results[index] = w.Code
			mutex.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all requests succeeded
	successCount := 0
	for _, code := range results {
		if code == http.StatusOK {
			successCount++
		}
	}

	assert.Equal(suite.T(), concurrentRequests, successCount, "All concurrent invoice retrievals should succeed")
}

// TestLargeDatasetHandling tests handling of large amounts of data
func (suite *APITestSuite) TestLargeDatasetHandling() {
	// Create multiple business partners
	partnerCount := 5
	businessPartnerIDs := make([]uint, partnerCount)
	for i := 0; i < partnerCount; i++ {
		partnerData := models.BusinessPartnerCreateRequest{
			CorporateName:  fmt.Sprintf("Large Data Partner %d", i),
			Representative: fmt.Sprintf("Rep %d", i),
			PhoneNumber:    fmt.Sprintf("03-5555-55%02d", i),
			PostalCode:     fmt.Sprintf("10%d-0001", i),
			Address:        fmt.Sprintf("Tokyo, Large Data Address %d-%d-%d", i, i, i),
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
		businessPartnerIDs[i] = uint(partnerMap["id"].(float64))
	}

	// Create many invoices for each business partner
	invoicesPerPartner := 10
	totalInvoices := partnerCount * invoicesPerPartner

	startTime := time.Now()

	for i, partnerID := range businessPartnerIDs {
		for j := 0; j < invoicesPerPartner; j++ {
			invoiceData := models.CreateInvoiceRequest{
				BusinessPartnerID: partnerID,
				PaymentAmount:     float64(5000 + (i*invoicesPerPartner+j)*500),
				PaymentDueDate:    time.Now().AddDate(0, 0, i*invoicesPerPartner+j+1),
			}

			jsonData, _ := json.Marshal(invoiceData)
			req, _ := http.NewRequest("POST", "/api/invoices", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			assert.Equal(suite.T(), http.StatusOK, w.Code, "Invoice creation should succeed for large dataset")
		}
	}

	creationTime := time.Since(startTime)
	suite.T().Logf("Created %d invoices in %v (avg: %v per invoice)",
		totalInvoices, creationTime, creationTime/time.Duration(totalInvoices))

	// Test retrieval performance with large dataset
	startTime = time.Now()

	req, _ := http.NewRequest("GET", "/api/invoices", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	retrievalTime := time.Since(startTime)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	suite.T().Logf("Retrieved invoices in %v", retrievalTime)

	// Verify response contains invoices
	if response.Data != nil {
		invoices := response.Data.([]interface{})
		assert.True(suite.T(), len(invoices) > 0, "Should return invoices from large dataset")
		suite.T().Logf("Retrieved %d invoices from large dataset", len(invoices))
	}
}

// TestPerformanceWithFiltering tests performance when filtering large datasets
func (suite *APITestSuite) TestPerformanceWithFiltering() {
	// Test filtering performance
	testCases := []struct {
		name        string
		queryParams string
	}{
		{
			name:        "Filter by status",
			queryParams: "?status=unprocessed",
		}, {
			name: "Filter by date range",
			queryParams: fmt.Sprintf("?start_date=%s&end_date=%s",
				url.QueryEscape(time.Now().Format(time.RFC3339)),
				url.QueryEscape(time.Now().AddDate(0, 0, 30).Format(time.RFC3339))),
		},
		{
			name: "Filter by status and date",
			queryParams: fmt.Sprintf("?status=unprocessed&start_date=%s&end_date=%s",
				url.QueryEscape(time.Now().Format(time.RFC3339)),
				url.QueryEscape(time.Now().AddDate(0, 0, 15).Format(time.RFC3339))),
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			startTime := time.Now()

			req, _ := http.NewRequest("GET", "/api/invoices"+tc.queryParams, nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			duration := time.Since(startTime)
			assert.Equal(t, http.StatusOK, w.Code)

			t.Logf("Filter '%s' completed in %v", tc.name, duration)

			// Verify reasonable response time (adjust threshold as needed)
			assert.True(t, duration < 5*time.Second, "Filter query should complete within reasonable time")
		})
	}
}

// TestMemoryUsage tests memory efficiency with large datasets
func (suite *APITestSuite) TestMemoryUsage() {
	// This is a basic test - in a real scenario, you'd use more sophisticated
	// memory profiling tools

	// Get initial memory stats (simplified)
	var initialMemStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMemStats)
	// Create business partner for memory test
	partnerData := models.BusinessPartnerCreateRequest{
		CorporateName:  "Memory Test Partner",
		Representative: "Memory Test Rep",
		PhoneNumber:    "03-6666-6666",
		PostalCode:     "106-0001",
		Address:        "Tokyo, Memory Test Address 7-7-7",
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

	// Perform multiple operations to test memory usage
	operationCount := 50
	for i := 0; i < operationCount; i++ {
		// Create invoice
		invoiceData := models.CreateInvoiceRequest{
			BusinessPartnerID: businessPartnerID,
			PaymentAmount:     float64(1000 + i*100),
			PaymentDueDate:    time.Now().AddDate(0, 0, i+1),
		}

		jsonData, _ := json.Marshal(invoiceData)
		req, _ := http.NewRequest("POST", "/api/invoices", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		// Retrieve invoices to test memory usage
		req, _ = http.NewRequest("GET", "/api/invoices", nil)
		req.Header.Set("Authorization", "Bearer "+suite.authToken)

		w = httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)
	}

	// Get final memory stats
	var finalMemStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&finalMemStats)

	// Log memory usage information
	suite.T().Logf("Memory usage - Initial: %d KB, Final: %d KB, Difference: %d KB",
		initialMemStats.Alloc/1024,
		finalMemStats.Alloc/1024,
		(finalMemStats.Alloc-initialMemStats.Alloc)/1024)
	// This is a basic check - in production, you'd want more sophisticated memory leak detection
	assert.True(suite.T(), finalMemStats.Alloc < initialMemStats.Alloc*10,
		"Memory usage should not grow excessively")
}
