# Go VNPay Integration Service

A robust Golang-based payment integration service for VNPay with PostgreSQL invoice storage, containerized with Docker.

## Overview

This project provides a comprehensive API to integrate with the VNPay payment gateway, allowing your applications to process payment transactions securely and efficiently. All transaction records are stored in a PostgreSQL database for reliable record keeping and reporting.

### Features

- Create payment transactions with VNPay
- Authenticate and process callbacks from VNPay
- Check transaction status
- Process refunds for transactions
- Store all transaction details in PostgreSQL database (invoices table)
- Comprehensive logging and monitoring

## Technology Stack

- **Backend**: Golang
- **Database**: PostgreSQL
- **Containerization**: Docker & Docker Compose

## Prerequisites

- Docker and Docker Compose installed
- Git
- Basic understanding of RESTful APIs and Docker

## Environment Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/binhkhang3169/golang-vnpay.git
   cd golang-vnpay
   ```

2. Create an environment file:
   ```bash
   cp .env.example .env
   ```

3. Configure your `.env` file with the necessary values:
   ```
   # Server Configuration
   SERVER_PORT=8080
   GIN_MODE=release
   READ_TIMEOUT=30s
   WRITE_TIMEOUT=30s

   # Database Configuration
   DB_HOST=postgres
   DB_PORT=5432
   DB_NAME=payment_service
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_SSL_MODE=disable
   DB_MAX_CONNECTIONS=10
   DB_MIN_CONNECTIONS=2

   # Kafka Configuration
   KAFKA_BROKERS=kafka:9092
   KAFKA_CONSUMER_GROUP=payment_service-group
   KAFKA_PAYMENT_TOPIC=payment-events
   KAFKA_NOTIFICATION_TOPIC=notification-events

   # VNPay Configuration
   VNPAY_MERCHANT_ID=your-merchant-id
   VNPAY_MERCHANT_NAME="Payment Service"
   VNPAY_HASH_SECRET=your-secret-key
   VNPAY_API_URL=https://sandbox.vnpayment.vn/paymentv2/vpcpay.html
   VNPAY_API_VERSION=2.1.0
   VNPAY_COMMAND=pay
   VNPAY_CURRENCY_CODE=VND
   VNPAY_LOCALE=vn
   VNPAY_RETURN_URL=http://localhost:8080/api/v1/payment/vnpay-return

   # Auth Configuration
   API_SECRET_KEY=your-very-strong-secret-key-here
   API_TOKEN_EXPIRY=24h

   # App Configurations
   LOG_LEVEL=info
   ```

## Database Structure

The main database table for storing invoice information:

```sql
CREATE TABLE IF NOT EXISTS invoices (
    invoice_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    invoice_type VARCHAR(50),
    customer_id VARCHAR(100) NOT NULL,
    ticket_id VARCHAR(100) NOT NULL,
    total_amount DECIMAL(15, 2) NOT NULL,
    discount_amount DECIMAL(15, 2) DEFAULT 0.00,
    tax_amount DECIMAL(15, 2) DEFAULT 0.00,
    final_amount DECIMAL(15, 2) NOT NULL,
    payment_status VARCHAR(20) DEFAULT 'PENDING',
    payment_method VARCHAR(50),
    issue_date TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- VNPay specific fields
    vnpay_txn_ref VARCHAR(100),
    vnpay_bank_code VARCHAR(50),
    vnpay_txn_no VARCHAR(100),
    vnpay_pay_date VARCHAR(50)
);
```

## Running the Service

Build and start all services using Docker Compose:

```bash
docker-compose up -d --build
```

This command will:
1. Build or rebuild all images in the docker-compose file
2. Create and start containers in detached mode
3. Set up the PostgreSQL database with the required schema
4. Start the Go API service connected to both VNPay and the database

## API Endpoints

### Payment Endpoints

- **Create Payment**: `POST /api/vnpay/create-payment`
- **Process Refund**: `POST /api/vnpay/refund`

### Invoice Endpoints

- **Get Invoice by Transaction ID**: `GET /api/invoices/:transactionId`
- **Get Invoices by Customer ID**: `GET /api/invoices/customer/:customerId`

## Docker Architecture

The service consists of two main containers:

1. **go-vnpay-service**: The Golang application that handles VNPay integration and API endpoints
2. **postgres**: PostgreSQL database for storing invoice information

The Docker Compose network enables secure communication between these services.

## Development

### Rebuilding the Service

To rebuild and restart the service after code changes:

```bash
docker-compose down
docker-compose up -d --build
```

### Accessing Logs

```bash
# View all logs
docker-compose logs

# Follow logs from a specific service
docker-compose logs -f go-vnpay-service
```

### Database Management

You can connect to the PostgreSQL database using any PostgreSQL client with the following credentials:

- **Host**: localhost
- **Port**: 5432 (or the port specified in your .env)
- **Username**: As specified in POSTGRES_USER
- **Password**: As specified in POSTGRES_PASSWORD
- **Database**: As specified in POSTGRES_DB

## Testing

Run the automated tests:

```bash
# Run inside the container
docker exec -it go-vnpay-service go test ./...

# Or run locally with correctly configured environment
go test ./...
```

## Security Considerations

- All VNPay API keys and sensitive information should be stored in the `.env` file and not committed to version control
- HTTPS is recommended for production environments
- Implement proper validation for all incoming payment data
- Set up appropriate database access restrictions

## Troubleshooting

If you encounter any issues during setup or operation:

1. Check the logs for error messages
2. Verify that your `.env` file contains correct configuration
3. Ensure your VNPay credentials are valid
4. Confirm database connectivity

## Contributing

Please refer to CONTRIBUTING.md for guidelines on contributing to this project.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
