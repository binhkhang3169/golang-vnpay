package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"payment_service/config"
	"payment_service/domain/model"
)

// VNPayService handles the VNPay payment integration
type VNPayService struct {
	config     *config.VNPayConfig
	invoiceSvc *InvoiceService
}

// NewVNPayService creates a new VNPay service
func NewVNPayService(cfg *config.VNPayConfig, invoiceSvc *InvoiceService) *VNPayService {
	return &VNPayService{
		config:     cfg,
		invoiceSvc: invoiceSvc,
	}
}

// CreatePayment creates a new payment URL for VNPay
func (s *VNPayService) CreatePayment(ctx context.Context, req model.VNPayPaymentRequest) (*model.VNPayPaymentResponse, error) {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Set transaction reference (order ID)
	txnRef := strconv.Itoa(rand.Intn(9999999) + 1000000)

	// Create invoice in database
	invoice, err := s.invoiceSvc.CreateInvoice(ctx, req, txnRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Get current time for transaction
	now := time.Now()
	createDate := now.Format("20060102150405")

	// Calculate expiry time (15 minutes from now)
	expireTime := now.Add(15 * time.Minute).Format("20060102150405")

	// Convert amount to VND cents (multiply by 100)
	amountInCents := int(req.Amount * 100)

	// Create input data map
	inputData := map[string]string{
		"vnp_Version":    "2.1.0",
		"vnp_TmnCode":    s.config.TmnCode,
		"vnp_Amount":     strconv.Itoa(amountInCents),
		"vnp_Command":    "pay",
		"vnp_CreateDate": createDate,
		"vnp_CurrCode":   "VND",
		"vnp_IpAddr":     "127.0.0.1", // This should be passed from the controller
		"vnp_Locale":     req.Language,
		"vnp_OrderInfo":  fmt.Sprintf("Thanh toan cho don hang %s", invoice.InvoiceNumber),
		"vnp_OrderType":  "other",
		"vnp_ReturnUrl":  s.config.ReturnURL,
		"vnp_TxnRef":     txnRef,
		"vnp_ExpireDate": expireTime,
	}

	// Add bank code if provided
	if req.BankCode != "" {
		inputData["vnp_BankCode"] = req.BankCode
	}

	// Sort the keys for secure hash generation
	keys := make([]string, 0, len(inputData))
	for k := range inputData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query and hash data
	var queryBuilder strings.Builder
	var hashDataBuilder strings.Builder

	for i, k := range keys {
		// URL encode both key and value
		encodedKey := url.QueryEscape(k)
		encodedValue := url.QueryEscape(inputData[k])

		// Add to query string
		if i > 0 {
			queryBuilder.WriteString("&")
		}
		queryBuilder.WriteString(encodedKey)
		queryBuilder.WriteString("=")
		queryBuilder.WriteString(encodedValue)

		// Add to hash data
		if i > 0 {
			hashDataBuilder.WriteString("&")
		}
		hashDataBuilder.WriteString(encodedKey)
		hashDataBuilder.WriteString("=")
		hashDataBuilder.WriteString(encodedValue)
	}

	// Create the payment URL
	vnpURL := s.config.VNPayURL + "?" + queryBuilder.String()

	// Calculate secure hash
	hashData := hashDataBuilder.String()
	hmacObj := hmac.New(sha512.New, []byte(s.config.HashSecret))
	hmacObj.Write([]byte(hashData))
	vnpSecureHash := hex.EncodeToString(hmacObj.Sum(nil))

	// Add secure hash to URL
	vnpURL = vnpURL + "&vnp_SecureHash=" + vnpSecureHash

	// Prepare response
	txnRefInt, _ := strconv.Atoi(txnRef)
	response := &model.VNPayPaymentResponse{
		Code:    "00",
		Message: "success",
		Data: struct {
			PaymentURL string `json:"payment_url"`
			TxnRef     int    `json:"txn_ref"`
		}{
			PaymentURL: vnpURL,
			TxnRef:     txnRefInt,
		},
		InvoiceID: invoice.InvoiceID.String(),
	}

	return response, nil
}

// ProcessReturn processes the return from VNPay payment gateway
func (s *VNPayService) ProcessReturn(ctx context.Context, queryParams url.Values) (*model.VNPayReturnResponse, error) {
	// Get the secure hash from the query
	vnpSecureHash := queryParams.Get("vnp_SecureHash")

	// Create a map to store all "vnp_" parameters
	inputData := make(map[string]string)
	for key, values := range queryParams {
		if strings.HasPrefix(key, "vnp_") && key != "vnp_SecureHash" {
			inputData[key] = values[0]
		}
	}

	// Sort the keys
	keys := make([]string, 0, len(inputData))
	for k := range inputData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build the hash data
	var hashDataBuilder strings.Builder
	for i, k := range keys {
		if i > 0 {
			hashDataBuilder.WriteString("&")
		}
		hashDataBuilder.WriteString(url.QueryEscape(k))
		hashDataBuilder.WriteString("=")
		hashDataBuilder.WriteString(url.QueryEscape(inputData[k]))
	}

	// Calculate secure hash
	hashData := hashDataBuilder.String()
	hmacObj := hmac.New(sha512.New, []byte(s.config.HashSecret))
	hmacObj.Write([]byte(hashData))
	secureHash := hex.EncodeToString(hmacObj.Sum(nil))

	// Verify the secure hash
	isValidSignature := secureHash == vnpSecureHash

	// Get amount, convert to number
	amountStr := queryParams.Get("vnp_Amount")
	amountInt, _ := strconv.Atoi(amountStr)
	amount := float64(amountInt) / 100 // Convert back from VND cents

	// Get transaction reference
	txnRef := queryParams.Get("vnp_TxnRef")

	// Check payment status
	responseCode := queryParams.Get("vnp_ResponseCode")

	// Extract details for updating the invoice
	vnpayData := map[string]string{
		"transactionNo": queryParams.Get("vnp_TransactionNo"),
		"bankCode":      queryParams.Get("vnp_BankCode"),
		"payDate":       queryParams.Get("vnp_PayDate"),
	}

	// Prepare result message
	var result string
	var paymentStatus model.PaymentStatus

	if isValidSignature {
		if responseCode == "00" {
			result = "Payment successful"
			paymentStatus = model.PaymentStatusCompleted
		} else {
			result = "Payment failed"
			paymentStatus = model.PaymentStatusFailed
		}

		// Update invoice payment status
		err := s.invoiceSvc.UpdateInvoicePaymentStatus(ctx, txnRef, paymentStatus, vnpayData)
		if err != nil {
			return nil, fmt.Errorf("failed to update invoice: %w", err)
		}
	} else {
		result = "Invalid signature"
	}

	// Get invoice ID for the transaction
	invoice, err := s.invoiceSvc.GetInvoiceByVNPayTxnRef(ctx, txnRef)
	var invoiceID string
	if err == nil {
		invoiceID = invoice.InvoiceID.String()
	}

	// Return payment result
	return &model.VNPayReturnResponse{
		IsValid:        isValidSignature,
		TransactionNo:  queryParams.Get("vnp_TransactionNo"),
		Amount:         amount,
		OrderInfo:      queryParams.Get("vnp_OrderInfo"),
		ResponseCode:   responseCode,
		BankCode:       queryParams.Get("vnp_BankCode"),
		PaymentTime:    queryParams.Get("vnp_PayDate"),
		TransactionRef: txnRef,
		Result:         result,
		InvoiceID:      invoiceID,
	}, nil
}

// ProcessIPN processes the Instant Payment Notification from VNPay
func (s *VNPayService) ProcessIPN(ctx context.Context, queryParams url.Values) (*model.VNPayIPNResponse, error) {
	// Get the secure hash from the query
	vnpSecureHash := queryParams.Get("vnp_SecureHash")

	// Create a map to store all "vnp_" parameters
	inputData := make(map[string]string)
	for key, values := range queryParams {
		if strings.HasPrefix(key, "vnp_") && key != "vnp_SecureHash" {
			inputData[key] = values[0]
		}
	}

	// Sort the keys
	keys := make([]string, 0, len(inputData))
	for k := range inputData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build the hash data
	var hashDataBuilder strings.Builder
	for i, k := range keys {
		if i > 0 {
			hashDataBuilder.WriteString("&")
		}
		hashDataBuilder.WriteString(url.QueryEscape(k))
		hashDataBuilder.WriteString("=")
		hashDataBuilder.WriteString(url.QueryEscape(inputData[k]))
	}

	// Calculate secure hash
	hashData := hashDataBuilder.String()
	hmacObj := hmac.New(sha512.New, []byte(s.config.HashSecret))
	hmacObj.Write([]byte(hashData))
	secureHash := hex.EncodeToString(hmacObj.Sum(nil))

	// Prepare return data
	returnData := &model.VNPayIPNResponse{
		RspCode: "99",
		Message: "Unknown error",
	}

	// Verify the secure hash
	if secureHash == vnpSecureHash {
		// Get transaction data
		txnRef := queryParams.Get("vnp_TxnRef")
		responseCode := queryParams.Get("vnp_ResponseCode")
		transactionStatus := queryParams.Get("vnp_TransactionStatus")

		// Extract details for updating the invoice
		vnpayData := map[string]string{
			"transactionNo": queryParams.Get("vnp_TransactionNo"),
			"bankCode":      queryParams.Get("vnp_BankCode"),
			"payDate":       queryParams.Get("vnp_PayDate"),
		}

		// Try to get the invoice
		invoice, err := s.invoiceSvc.GetInvoiceByVNPayTxnRef(ctx, txnRef)
		if err != nil {
			returnData.RspCode = "01"
			returnData.Message = "Order not found"
			return returnData, nil
		}

		// Verify amount
		vnpAmount := queryParams.Get("vnp_Amount")
		amount, _ := strconv.Atoi(vnpAmount)
		amount = amount / 100 // Convert back from VND cents

		expectedAmount := int(invoice.FinalAmount * 100)
		if amount != expectedAmount {
			returnData.RspCode = "04"
			returnData.Message = "Invalid amount"
			return returnData, nil
		}

		// Check if payment is already processed
		if invoice.PaymentStatus != model.PaymentStatusPending {
			returnData.RspCode = "02"
			returnData.Message = "Order already confirmed"
			return returnData, nil
		}

		// Update payment status based on VNPay response
		var paymentStatus model.PaymentStatus
		if responseCode == "00" && transactionStatus == "00" {
			paymentStatus = model.PaymentStatusCompleted
		} else {
			paymentStatus = model.PaymentStatusFailed
		}

		// Update invoice payment status
		err = s.invoiceSvc.UpdateInvoicePaymentStatus(ctx, txnRef, paymentStatus, vnpayData)
		if err != nil {
			returnData.RspCode = "99"
			returnData.Message = "Error updating payment status"
			return returnData, err
		}

		returnData.RspCode = "00"
		returnData.Message = "Confirm Success"
	} else {
		returnData.RspCode = "97"
		returnData.Message = "Invalid signature"
	}

	return returnData, nil
}

// QueryTransaction prepares data for querying a transaction
func (s *VNPayService) QueryTransaction(ctx context.Context, req model.VNPayQueryRequest, ipAddr string) (map[string]string, error) {
	// Create a request ID
	requestId := strconv.Itoa(rand.Intn(9999) + 1)

	// Get current time for request
	createDate := time.Now().Format("20060102150405")

	// Build data request
	dataRequest := map[string]string{
		"vnp_RequestId":       requestId,
		"vnp_Version":         "2.1.0",
		"vnp_Command":         "querydr",
		"vnp_TmnCode":         s.config.TmnCode,
		"vnp_TxnRef":          req.TxnRef,
		"vnp_OrderInfo":       "Query transaction",
		"vnp_TransactionDate": req.TransactionDate,
		"vnp_CreateDate":      createDate,
		"vnp_IpAddr":          ipAddr,
	}

	// Format string for hash
	hashFormat := "%s|%s|%s|%s|%s|%s|%s|%s|%s"

	// Create hash data string
	hashData := fmt.Sprintf(
		hashFormat,
		dataRequest["vnp_RequestId"],
		dataRequest["vnp_Version"],
		dataRequest["vnp_Command"],
		dataRequest["vnp_TmnCode"],
		dataRequest["vnp_TxnRef"],
		dataRequest["vnp_TransactionDate"],
		dataRequest["vnp_CreateDate"],
		dataRequest["vnp_IpAddr"],
		dataRequest["vnp_OrderInfo"],
	)

	// Calculate checksum
	hmacObj := hmac.New(sha512.New, []byte(s.config.HashSecret))
	hmacObj.Write([]byte(hashData))
	checksum := hex.EncodeToString(hmacObj.Sum(nil))

	// Add checksum to request
	dataRequest["vnp_SecureHash"] = checksum

	return dataRequest, nil
}

// RefundTransaction prepares data for refunding a transaction
func (s *VNPayService) RefundTransaction(ctx context.Context, req model.VNPayRefundRequest, ipAddr string) (map[string]string, error) {
	// Create a request ID
	requestId := strconv.Itoa(rand.Intn(9999) + 1)

	// Get current time for request
	createDate := time.Now().Format("20060102150405")

	// Convert amount to VND cents (multiply by 100)
	amountInCents := int(req.Amount * 100)

	// Build data request
	refundData := map[string]string{
		"vnp_RequestId":       requestId,
		"vnp_Version":         "2.1.0",
		"vnp_Command":         "refund",
		"vnp_TmnCode":         s.config.TmnCode,
		"vnp_TransactionType": req.TransactionType, // 02: full refund, 03: partial refund
		"vnp_TxnRef":          req.TxnRef,
		"vnp_Amount":          strconv.Itoa(amountInCents),
		"vnp_OrderInfo":       "Hoan Tien Giao Dich",
		"vnp_TransactionNo":   "0", // "0": merchant didn't receive transaction code from VNPAY
		"vnp_TransactionDate": req.TransactionDate,
		"vnp_CreateDate":      createDate,
		"vnp_CreateBy":        req.CreateBy,
		"vnp_IpAddr":          ipAddr,
	}

	// Format string for hash
	hashFormat := "%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s"

	// Create hash data string
	hashData := fmt.Sprintf(
		hashFormat,
		refundData["vnp_RequestId"],
		refundData["vnp_Version"],
		refundData["vnp_Command"],
		refundData["vnp_TmnCode"],
		refundData["vnp_TransactionType"],
		refundData["vnp_TxnRef"],
		refundData["vnp_Amount"],
		refundData["vnp_TransactionNo"],
		refundData["vnp_TransactionDate"],
		refundData["vnp_CreateBy"],
		refundData["vnp_CreateDate"],
		refundData["vnp_IpAddr"],
		refundData["vnp_OrderInfo"],
	)

	// Calculate checksum
	hmacObj := hmac.New(sha512.New, []byte(s.config.HashSecret))
	hmacObj.Write([]byte(hashData))
	checksum := hex.EncodeToString(hmacObj.Sum(nil))

	// Add checksum to request
	refundData["vnp_SecureHash"] = checksum

	// If successful, update invoice status
	_, err := s.invoiceSvc.GetInvoiceByVNPayTxnRef(ctx, req.TxnRef)
	if err == nil {
		vnpayData := map[string]string{}
		err = s.invoiceSvc.UpdateInvoicePaymentStatus(ctx, req.TxnRef, model.PaymentStatusRefunded, vnpayData)
		if err != nil {
			return refundData, fmt.Errorf("failed to update invoice status: %w", err)
		}
	}

	return refundData, nil
}
