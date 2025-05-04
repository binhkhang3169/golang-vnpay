package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the status of an invoice payment
type PaymentStatus string

// Payment status constants
const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

// PaymentMethod represents the method used for payment
type PaymentMethod string

// Payment method constants
const (
	PaymentMethodVNPay PaymentMethod = "VNPAY"
)

// Invoice represents an invoice in the system
type Invoice struct {
	InvoiceID      uuid.UUID     `json:"invoice_id"`
	InvoiceNumber  string        `json:"invoice_number"`
	InvoiceType    string        `json:"invoice_type"`
	CustomerID     string        `json:"customer_id"`
	TicketID       string        `json:"ticket_id"`
	TotalAmount    float64       `json:"total_amount"`
	DiscountAmount float64       `json:"discount_amount"`
	TaxAmount      float64       `json:"tax_amount"`
	FinalAmount    float64       `json:"final_amount"`
	PaymentStatus  PaymentStatus `json:"payment_status"`
	PaymentMethod  PaymentMethod `json:"payment_method"`
	IssueDate      time.Time     `json:"issue_date"`
	Notes          string        `json:"notes"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`

	// VNPay specific fields
	VNPayTxnRef   string `json:"vnpay_txn_ref,omitempty"`
	VNPayBankCode string `json:"vnpay_bank_code,omitempty"`
	VNPayTxnNo    string `json:"vnpay_txn_no,omitempty"`
	VNPayPayDate  string `json:"vnpay_pay_date,omitempty"`
}

// VNPayPaymentRequest holds the request data for creating a new VNPay payment
type VNPayPaymentRequest struct {
	CustomerID     string  `json:"customer_id" binding:"required"`
	TicketID       string  `json:"ticket_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	Language       string  `json:"language" binding:"required"`
	BankCode       string  `json:"bank_code"`
	InvoiceType    string  `json:"invoice_type"`
	DiscountAmount float64 `json:"discount_amount"`
	TaxAmount      float64 `json:"tax_amount"`
}

// VNPayPaymentResponse represents the response from the payment creation request
type VNPayPaymentResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		PaymentURL string `json:"payment_url"`
		TxnRef     int    `json:"txn_ref"`
	} `json:"data"`
	InvoiceID string `json:"invoice_id"`
}

// VNPayReturnResponse represents the response data after VNPay payment completion
type VNPayReturnResponse struct {
	IsValid        bool    `json:"isValid"`
	TransactionNo  string  `json:"transactionNo"`
	Amount         float64 `json:"amount"`
	OrderInfo      string  `json:"orderInfo"`
	ResponseCode   string  `json:"responseCode"`
	BankCode       string  `json:"bankCode"`
	PaymentTime    string  `json:"paymentTime"`
	TransactionRef string  `json:"transactionRef"`
	Result         string  `json:"result"`
	InvoiceID      string  `json:"invoice_id"`
}

// VNPayIPNRequest represents the IPN request data from VNPay
type VNPayIPNRequest struct {
	Amount            string `form:"vnp_Amount"`
	BankCode          string `form:"vnp_BankCode"`
	OrderInfo         string `form:"vnp_OrderInfo"`
	PayDate           string `form:"vnp_PayDate"`
	ResponseCode      string `form:"vnp_ResponseCode"`
	SecureHash        string `form:"vnp_SecureHash"`
	TmnCode           string `form:"vnp_TmnCode"`
	TransactionNo     string `form:"vnp_TransactionNo"`
	TransactionStatus string `form:"vnp_TransactionStatus"`
	TxnRef            string `form:"vnp_TxnRef"`
}

// VNPayIPNResponse represents the response to an IPN request
type VNPayIPNResponse struct {
	RspCode string `json:"RspCode"`
	Message string `json:"Message"`
}

// VNPayQueryRequest represents a request to query a transaction
type VNPayQueryRequest struct {
	TxnRef          string `json:"txnRef" binding:"required"`
	TransactionDate string `json:"transactionDate" binding:"required"`
}

// VNPayRefundRequest represents a request to refund a transaction
type VNPayRefundRequest struct {
	TxnRef          string  `json:"txnRef" binding:"required"`
	TransactionType string  `json:"transactionType" binding:"required"`
	Amount          float64 `json:"amount" binding:"required"`
	TransactionDate string  `json:"transactionDate" binding:"required"`
	CreateBy        string  `json:"createBy" binding:"required"`
}
