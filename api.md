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
