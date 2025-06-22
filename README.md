# Super Payment-kun.com REST API

A Go-based REST API for a fictional corporate payment service that allows users to register invoice data for future payment processing with automatic bank transfers.

## Features

- **User Authentication**: JWT-based authentication system
- **Company Management**: Multi-tenant system supporting multiple companies
- **Business Partner Management**: Manage business partners and their bank accounts
- **Invoice Management**: Create and manage invoices with automatic fee calculation
- **Automatic Calculations**: 4% service fee + 10% consumption tax on fees
- **Date-based Filtering**: Retrieve invoices within specified periods
- **Security**: Password hashing, JWT tokens, and proper authorization

## Folder Structure

Clean architecture pattern.

```
super-payment/
├── cmd/server/           # Application entry point
├── internal/
│   ├── api/             # HTTP handlers and routing
│   ├── config/          # Configuration management
│   ├── middleware/      # HTTP middleware (auth, CORS, logging)
│   ├── models/          # Data models and DTOs
│   ├── repository/      # Data access layer
│   └── service/         # Business logic layer
├── migrations/          # Database migration scripts
├── tests/               # Test files
├── go.mod              # Go module dependencies
└── README.md           # This file
```

## API Endpoints

Live Version: [OpenAPI Spec](./api-docs.yaml)

## Prerequisites

- Go 1.21+
- MySQL 8.0+
- Git
- Docker with Docker Compose (optional)

## Setup Instructions

### 1. Clone the Repository

```bash
git clone <repository-url>
cd super-payment
```

Then, you can start with Docker Compose (optional):

```bash
docker-compose up -d
# restart: docker-compose restart
```

Visit `http://localhost:8080/health` to check the health of the server.

or alternatively, you can set up manually as follows:

### 2. Install Go Dependencies

```bash
go mod tidy
```

### 3. Database Setup

Create a MySQL database:

```sql
CREATE DATABASE super_payment CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Run the migration scripts:

```bash
mysql -u root -p super_payment < migrations/001_create_tables.sql
mysql -u root -p super_payment < migrations/002_insert_sample_data.sql
```

### 4. Environment Configuration

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` with your database credentials:

```env
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=super_payment

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY_HOURS=24
```

### 5. Run the Application

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## Usage Examples

### 1. Register a New User and Company

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "company": {
      "corporate_name": "My Company Inc.",
      "representative": "John Doe",
      "phone_number": "03-1234-5678",
      "postal_code": "100-0001",
      "address": "Tokyo, Chiyoda-ku, Example 1-1-1"
    },
    "user": {
      "full_name": "John Doe",
      "email": "john@mycompany.com",
      "password": "securepassword123"
    }
  }'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@mycompany.com",
    "password": "securepassword123"
  }'
```

### 3. Create a Business Partner

```bash
curl -X POST http://localhost:8080/api/business-partners \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "corporate_name": "Supplier Corp.",
    "representative": "Jane Smith",
    "phone_number": "03-9876-5432",
    "postal_code": "101-0001",
    "address": "Tokyo, Chiyoda-ku, Supplier 2-2-2"
  }'
```

### 4. Create an Invoice

```bash
curl -X POST http://localhost:8080/api/invoices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "business_partner_id": 1,
    "payment_amount": 100000,
    "payment_due_date": "2024-12-31T00:00:00Z"
  }'
```

### 5. Get Invoices with Date Filter

```bash
curl "http://localhost:8080/api/invoices?start_date=2024-01-01&end_date=2024-12-31&status=unprocessed" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Invoice Calculation Logic

The system automatically calculates invoice amounts based on the following fee and tax structure:

```
Payment Amount: 100,000 (input)
Fee (4%): 100,000 × 0.04 = 4,000
Consumption Tax (10% on fee): 4,000 × 0.10 = 400
Invoice Amount: 100,000 + 4,000 + 400 = 104,400
```

## Data Models

Live Version: [models.go](internal/models/models.go)

## Testing

Run the test suite:

```bash
go test ./tests/... -v
```

For test coverage:

```bash
go test ./tests/... -cover
```


## API Specification

All protected endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Error Responses

All error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human readable error message"
}
```

### Success Responses

Success responses follow this format:

```json
{
  "message": "Success message",
  "data": { ... }
}
```

