package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	"payment_service/domain/model"
)

// InvoiceRepository handles invoice database operations
type InvoiceRepository struct {
	db *pgxpool.Pool
}

// NewInvoiceRepository creates a new invoice repository
func NewInvoiceRepository(db *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{
		db: db,
	}
}

// CreateInvoice creates a new invoice in the database
func (r *InvoiceRepository) CreateInvoice(ctx context.Context, invoice model.Invoice) (model.Invoice, error) {
	query := `
		INSERT INTO invoices (
			invoice_id, invoice_number, invoice_type, customer_id, ticket_id,
			total_amount, discount_amount, tax_amount, final_amount,
			payment_status, payment_method, issue_date, notes, 
			vnpay_txn_ref
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING invoice_id, created_at, updated_at
	`

	// If invoice_id is nil (zero uuid), generate a new one
	if invoice.InvoiceID == uuid.Nil {
		invoice.InvoiceID = uuid.New()
	}

	// Generate invoice number with current timestamp if empty
	if invoice.InvoiceNumber == "" {
		invoice.InvoiceNumber = fmt.Sprintf("INV-%s-%d",
			time.Now().Format("20060102"),
			time.Now().UnixNano()%1000000)
	}

	err := r.db.QueryRow(ctx, query,
		invoice.InvoiceID, invoice.InvoiceNumber, invoice.InvoiceType, invoice.CustomerID,
		invoice.TicketID, invoice.TotalAmount, invoice.DiscountAmount, invoice.TaxAmount,
		invoice.FinalAmount, invoice.PaymentStatus, invoice.PaymentMethod, invoice.IssueDate,
		invoice.Notes, invoice.VNPayTxnRef,
	).Scan(&invoice.InvoiceID, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err != nil {
		return model.Invoice{}, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

// GetInvoiceByID retrieves an invoice by ID
func (r *InvoiceRepository) GetInvoiceByID(ctx context.Context, id uuid.UUID) (model.Invoice, error) {
	query := `
		SELECT 
			invoice_id, invoice_number, invoice_type, customer_id, ticket_id,
			total_amount, discount_amount, tax_amount, final_amount,
			payment_status, payment_method, issue_date, notes, 
			created_at, updated_at,
			vnpay_txn_ref, vnpay_bank_code, vnpay_txn_no, vnpay_pay_date
		FROM invoices 
		WHERE invoice_id = $1
	`

	var invoice model.Invoice
	err := r.db.QueryRow(ctx, query, id).Scan(
		&invoice.InvoiceID, &invoice.InvoiceNumber, &invoice.InvoiceType, &invoice.CustomerID,
		&invoice.TicketID, &invoice.TotalAmount, &invoice.DiscountAmount, &invoice.TaxAmount,
		&invoice.FinalAmount, &invoice.PaymentStatus, &invoice.PaymentMethod, &invoice.IssueDate,
		&invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt,
		&invoice.VNPayTxnRef, &invoice.VNPayBankCode, &invoice.VNPayTxnNo, &invoice.VNPayPayDate,
	)

	if err != nil {
		return model.Invoice{}, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}

// GetInvoiceByVNPayTxnRef retrieves an invoice by VNPay transaction reference
func (r *InvoiceRepository) GetInvoiceByVNPayTxnRef(ctx context.Context, txnRef string) (model.Invoice, error) {
	query := `
		SELECT 
			invoice_id, invoice_number, invoice_type, customer_id, ticket_id,
			total_amount, discount_amount, tax_amount, final_amount,
			payment_status, payment_method, issue_date, notes, 
			created_at, updated_at,
			vnpay_txn_ref, vnpay_bank_code, vnpay_txn_no, vnpay_pay_date
		FROM invoices 
		WHERE vnpay_txn_ref = $1
	`

	var invoice model.Invoice
	err := r.db.QueryRow(ctx, query, txnRef).Scan(
		&invoice.InvoiceID, &invoice.InvoiceNumber, &invoice.InvoiceType, &invoice.CustomerID,
		&invoice.TicketID, &invoice.TotalAmount, &invoice.DiscountAmount, &invoice.TaxAmount,
		&invoice.FinalAmount, &invoice.PaymentStatus, &invoice.PaymentMethod, &invoice.IssueDate,
		&invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt,
		&invoice.VNPayTxnRef, &invoice.VNPayBankCode, &invoice.VNPayTxnNo, &invoice.VNPayPayDate,
	)

	if err != nil {
		return model.Invoice{}, fmt.Errorf("failed to get invoice by VNPay reference: %w", err)
	}

	return invoice, nil
}

// UpdateInvoicePaymentStatus updates an invoice's payment status and VNPay information
func (r *InvoiceRepository) UpdateInvoicePaymentStatus(ctx context.Context, txnRef string, status model.PaymentStatus, vnpayData map[string]string) error {
	query := `
		UPDATE invoices
		SET 
			payment_status = $1,
			vnpay_bank_code = $2,
			vnpay_txn_no = $3,
			vnpay_pay_date = $4,
			updated_at = NOW()
		WHERE vnpay_txn_ref = $5
	`

	_, err := r.db.Exec(ctx, query,
		status,
		vnpayData["bankCode"],
		vnpayData["transactionNo"],
		vnpayData["payDate"],
		txnRef,
	)

	if err != nil {
		return fmt.Errorf("failed to update invoice payment status: %w", err)
	}

	return nil
}

// GetInvoicesByCustomerID retrieves all invoices for a customer
func (r *InvoiceRepository) GetInvoicesByCustomerID(ctx context.Context, customerID string) ([]model.Invoice, error) {
	query := `
		SELECT 
			invoice_id, invoice_number, invoice_type, customer_id, ticket_id,
			total_amount, discount_amount, tax_amount, final_amount,
			payment_status, payment_method, issue_date, notes, 
			created_at, updated_at,
			vnpay_txn_ref, vnpay_bank_code, vnpay_txn_no, vnpay_pay_date
		FROM invoices 
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []model.Invoice
	for rows.Next() {
		var invoice model.Invoice
		err := rows.Scan(
			&invoice.InvoiceID, &invoice.InvoiceNumber, &invoice.InvoiceType, &invoice.CustomerID,
			&invoice.TicketID, &invoice.TotalAmount, &invoice.DiscountAmount, &invoice.TaxAmount,
			&invoice.FinalAmount, &invoice.PaymentStatus, &invoice.PaymentMethod, &invoice.IssueDate,
			&invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt,
			&invoice.VNPayTxnRef, &invoice.VNPayBankCode, &invoice.VNPayTxnNo, &invoice.VNPayPayDate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over invoices: %w", err)
	}

	return invoices, nil
}
