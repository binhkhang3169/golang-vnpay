package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"payment_service/domain/model"
	"payment_service/internal/repository"
)

// InvoiceService handles business logic related to invoices
type InvoiceService struct {
	repo *repository.InvoiceRepository
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(repo *repository.InvoiceRepository) *InvoiceService {
	return &InvoiceService{
		repo: repo,
	}
}

// CreateInvoice creates a new invoice with the given details
func (s *InvoiceService) CreateInvoice(ctx context.Context, req model.VNPayPaymentRequest, txnRef string) (model.Invoice, error) {
	// Calculate final amount
	finalAmount := req.Amount - req.DiscountAmount + req.TaxAmount

	// Create invoice object
	invoice := model.Invoice{
		InvoiceID:      uuid.New(),
		InvoiceType:    req.InvoiceType,
		CustomerID:     req.CustomerID,
		TicketID:       req.TicketID,
		TotalAmount:    req.Amount,
		DiscountAmount: req.DiscountAmount,
		TaxAmount:      req.TaxAmount,
		FinalAmount:    finalAmount,
		PaymentStatus:  model.PaymentStatusPending,
		PaymentMethod:  model.PaymentMethodVNPay,
		IssueDate:      time.Now(),
		Notes:          fmt.Sprintf("Payment via VNPay, TxnRef: %s", txnRef),
		VNPayTxnRef:    txnRef,
	}

	// Save invoice to database
	createdInvoice, err := s.repo.CreateInvoice(ctx, invoice)
	if err != nil {
		return model.Invoice{}, fmt.Errorf("failed to create invoice: %w", err)
	}

	return createdInvoice, nil
}

// GetInvoiceByID retrieves an invoice by its ID
func (s *InvoiceService) GetInvoiceByID(ctx context.Context, id uuid.UUID) (model.Invoice, error) {
	invoice, err := s.repo.GetInvoiceByID(ctx, id)
	if err != nil {
		return model.Invoice{}, fmt.Errorf("failed to get invoice: %w", err)
	}
	return invoice, nil
}

// GetInvoiceByVNPayTxnRef retrieves an invoice by its VNPay transaction reference
func (s *InvoiceService) GetInvoiceByVNPayTxnRef(ctx context.Context, txnRef string) (model.Invoice, error) {
	invoice, err := s.repo.GetInvoiceByVNPayTxnRef(ctx, txnRef)
	if err != nil {
		return model.Invoice{}, fmt.Errorf("failed to get invoice by VNPay reference: %w", err)
	}
	return invoice, nil
}

// UpdateInvoicePaymentStatus updates the payment status of an invoice
func (s *InvoiceService) UpdateInvoicePaymentStatus(ctx context.Context, txnRef string, status model.PaymentStatus, vnpayData map[string]string) error {
	err := s.repo.UpdateInvoicePaymentStatus(ctx, txnRef, status, vnpayData)
	if err != nil {
		return fmt.Errorf("failed to update invoice payment status: %w", err)
	}
	return nil
}

// GetInvoicesByCustomerID retrieves all invoices for a customer
func (s *InvoiceService) GetInvoicesByCustomerID(ctx context.Context, customerID string) ([]model.Invoice, error) {
	invoices, err := s.repo.GetInvoicesByCustomerID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices for customer: %w", err)
	}
	return invoices, nil
}
