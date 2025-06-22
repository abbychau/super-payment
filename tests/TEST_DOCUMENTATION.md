# Super Payment API - Test Documentation


### Running Specific Test Categories
```bash
# Integration tests
go test ./tests/api_test.go -v

# Validation tests  
go test ./tests/validation_test.go -v

# Performance tests
go test ./tests/performance_test.go -v

# Business logic tests
go test ./tests/business_logic_test.go -v
```

### Running with Coverage
```bash
go test ./tests/... -cover -v
```
