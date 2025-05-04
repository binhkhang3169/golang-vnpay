package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"payment_service/config"
	"payment_service/domain/model"
	"payment_service/internal/service"
)

// VNPayController handles VNPay payment API endpoints
type VNPayController struct {
	vnpaySvc   *service.VNPayService
	invoiceSvc *service.InvoiceService
	config     *config.VNPayConfig
}

// NewVNPayController creates a new VNPay controller
func NewVNPayController(vnpaySvc *service.VNPayService, invoiceSvc *service.InvoiceService, cfg *config.VNPayConfig) *VNPayController {
	return &VNPayController{
		vnpaySvc:   vnpaySvc,
		invoiceSvc: invoiceSvc,
		config:     cfg,
	}
}

// CreatePayment handles the creation of a new payment
func (c *VNPayController) CreatePayment(ctx *gin.Context) {
	var paymentRequest model.VNPayPaymentRequest

	if err := ctx.ShouldBindJSON(&paymentRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.vnpaySvc.CreatePayment(ctx, paymentRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// HandleReturn handles the return from VNPay after payment
func (c *VNPayController) HandleReturn(ctx *gin.Context) {
	// Get all query parameters
	queryParams := ctx.Request.URL.Query()

	response, err := c.vnpaySvc.ProcessReturn(ctx, queryParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// HandleIPN handles Instant Payment Notification from VNPay
func (c *VNPayController) HandleIPN(ctx *gin.Context) {
	// Get all query parameters
	queryParams := ctx.Request.URL.Query()

	response, err := c.vnpaySvc.ProcessIPN(ctx, queryParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// QueryTransaction handles transaction query requests
func (c *VNPayController) QueryTransaction(ctx *gin.Context) {
	var queryRequest model.VNPayQueryRequest

	if err := ctx.ShouldBindJSON(&queryRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataRequest, err := c.vnpaySvc.QueryTransaction(ctx, queryRequest, ctx.ClientIP())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":     "00",
		"message":  "Request prepared successfully",
		"data":     dataRequest,
		"endpoint": c.config.TransactionAPI,
		"note":     "In a real implementation, you would make an HTTP POST request to the VNPAY API with this data",
	})
}

// RefundTransaction handles refund requests
func (c *VNPayController) RefundTransaction(ctx *gin.Context) {
	var refundRequest model.VNPayRefundRequest

	if err := ctx.ShouldBindJSON(&refundRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refundData, err := c.vnpaySvc.RefundTransaction(ctx, refundRequest, ctx.ClientIP())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":     "00",
		"message":  "Refund request prepared successfully",
		"data":     refundData,
		"endpoint": c.config.TransactionAPI,
		"note":     "In a real implementation, you would make an HTTP POST request to the VNPAY API with this data",
	})
}

// GetInvoice retrieves an invoice by ID
func (c *VNPayController) GetInvoice(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	invoice, err := c.invoiceSvc.GetInvoiceByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	ctx.JSON(http.StatusOK, invoice)
}

// GetInvoicesByCustomer retrieves all invoices for a customer
func (c *VNPayController) GetInvoicesByCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customerId")
	if customerID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Customer ID is required"})
		return
	}

	invoices, err := c.invoiceSvc.GetInvoicesByCustomerID(ctx, customerID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, invoices)
}
