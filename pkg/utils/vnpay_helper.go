package utils

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// VNPayHelper contains utility functions for VNPay integration
type VNPayHelper struct {
	MerchantID string
	SecretKey  string
	ReturnURL  string
	PaymentURL string
}

// NewVNPayHelper creates a new instance of VNPayHelper
func NewVNPayHelper(merchantID, secretKey, returnURL, paymentURL string) *VNPayHelper {
	return &VNPayHelper{
		MerchantID: merchantID,
		SecretKey:  secretKey,
		ReturnURL:  returnURL,
		PaymentURL: paymentURL,
	}
}

// GeneratePaymentURL generates a VNPay payment URL
func (h *VNPayHelper) GeneratePaymentURL(
	amount int,
	orderInfo string,
	orderType string,
	txnRef string,
	ipAddr string,
) (string, error) {
	// Create parameters
	params := make(map[string]string)
	params["vnp_Version"] = "2.1.0"
	params["vnp_Command"] = "pay"
	params["vnp_TmnCode"] = h.MerchantID
	params["vnp_Amount"] = fmt.Sprintf("%d", amount*100) // Convert to VND smallest unit
	params["vnp_CurrCode"] = "VND"
	params["vnp_TxnRef"] = txnRef
	params["vnp_OrderInfo"] = orderInfo
	params["vnp_OrderType"] = orderType
	params["vnp_Locale"] = "vn"
	params["vnp_ReturnUrl"] = h.ReturnURL
	params["vnp_IpAddr"] = ipAddr

	// Create timestamp
	now := time.Now().UTC()
	vietnamTime := now.Add(7 * time.Hour) // UTC+7 for Vietnam time
	params["vnp_CreateDate"] = vietnamTime.Format("20060102150405")

	// Expire time: 15 minutes
	expireTime := vietnamTime.Add(15 * time.Minute)
	params["vnp_ExpireDate"] = expireTime.Format("20060102150405")

	// Generate query string
	queryString := h.buildQueryString(params)

	// Generate secure hash
	secureHash := h.generateSecureHash(queryString)

	// Return payment URL
	return h.PaymentURL + "?" + queryString + "&vnp_SecureHash=" + secureHash, nil
}

// ValidateCallback validates the VNPay callback
func (h *VNPayHelper) ValidateCallback(callbackParams map[string]string) bool {
	// Extract the secure hash
	secureHash := callbackParams["vnp_SecureHash"]
	delete(callbackParams, "vnp_SecureHash")
	delete(callbackParams, "vnp_SecureHashType")

	// Generate query string
	queryString := h.buildQueryString(callbackParams)

	// Generate secure hash
	calculatedHash := h.generateSecureHash(queryString)

	// Compare the hashes
	return secureHash == calculatedHash
}

// ParseResponseCode parses the VNPay response code
func (h *VNPayHelper) ParseResponseCode(responseCode string) string {
	switch responseCode {
	case "00":
		return "Successful transaction"
	case "01":
		return "Transaction not completed"
	case "02":
		return "Transaction error"
	case "03":
		return "Invalid merchant"
	case "04":
		return "Invalid transaction"
	case "05":
		return "Transaction not found"
	case "06":
		return "System error"
	case "07":
		return "Transaction made but pending for approval"
	case "08":
		return "Transaction rejected by bank"
	case "09":
		return "Transaction has been cancelled"
	case "10":
		return "Transaction cancelled by customer"
	case "11":
		return "Transaction expired"
	case "12":
		return "Transaction with invalid amount"
	case "13":
		return "Transaction with invalid currency"
	case "24":
		return "Customer cancelled the transaction"
	case "51":
		return "Not enough balance"
	case "65":
		return "Maximum transaction limit exceeded"
	case "75":
		return "Maximum transaction attempts exceeded"
	case "79":
		return "Authentication failed"
	case "99":
		return "Other errors"
	default:
		return "Unknown error"
	}
}

// buildQueryString builds a query string for VNPay
func (h *VNPayHelper) buildQueryString(params map[string]string) string {
	// Sort the keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build the query string
	var queryBuilder strings.Builder
	for i, k := range keys {
		if i > 0 {
			queryBuilder.WriteString("&")
		}
		queryBuilder.WriteString(k)
		queryBuilder.WriteString("=")
		queryBuilder.WriteString(url.QueryEscape(params[k]))
	}

	return queryBuilder.String()
}

// generateSecureHash generates a HMAC-SHA512 hash for VNPay
func (h *VNPayHelper) generateSecureHash(data string) string {
	key := []byte(h.SecretKey)
	hmacObj := hmac.New(sha512.New, key)
	hmacObj.Write([]byte(data))
	return hex.EncodeToString(hmacObj.Sum(nil))
}
