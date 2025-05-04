package controller

import "fmt"

func print() {
	fmt.Print("hello")
}

// import (
// 	"net/http"
// 	"strconv"

// 	"github.com/gin-gonic/gin"

// 	"payment_service/domain/model"
// 	"payment_service/internal/service"
// )

// // InvoiceController handles HTTP requests related to invoices
// type InvoiceController struct {
// 	invoiceService *service.InvoiceService
// }

// // NewInvoiceController creates a new invoice controller
// func NewInvoiceController(invoiceService *service.InvoiceService) *InvoiceController {
// 	return &InvoiceController{
// 		invoiceService: invoiceService,
// 	}
// }

// // Create handles the creation of a new invoice
// func (c *InvoiceController) Create(ctx *gin.Context) {
// 	var req model.CreateInvoiceRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	invoice, err := c.invoiceService.CreateInvoice(ctx.Request.Context(), req)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx.JSON(http.StatusCreated, gin.H{
// 		"code":    "00",
// 		"message": "Invoice created successfully",
// 		"data":    invoice,
// 	})
// }

// // GetByID retrieves an invoice by its ID
// func (c *InvoiceController) GetByID(ctx *gin.Context) {
// 	id := ctx.Param("id")

// 	invoice, err := c.invoiceService.GetInvoice(ctx.Request.Context(), id)
// 	if err != nil {
// 		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"code":    "00",
// 		"message": "Invoice retrieved successfully",
// 		"data":    invoice,
// 	})
// }

// // GetByNumber retrieves an invoice by its invoice number
// func (c *InvoiceController) GetByNumber(ctx *gin.Context) {
// 	number := ctx.Param("number")

// 	invoice, err := c.invoiceService.GetInvoiceByNumber(ctx.Request.Context(), number)
// 	if err != nil {
// 		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"code":    "00",
// 		"message": "Invoice retrieved successfully",
// 		"data":    invoice,
// 	})
// }

// // ListInvoices retrieves a list of invoices with pagination
// func (c *InvoiceController) ListInvoices(ctx *gin.Context) {
// 	pageStr := ctx.DefaultQuery("page", "1")
// 	limitStr := ctx.DefaultQuery("limit", "10")

// 	page, err := strconv.Atoi(pageStr)
// 	if err != nil || page < 1 {
// 		page = 1
// 	}

// 	limit, err := strconv.Atoi(limitStr)
// 	if err != nil || limit < 1 || limit > 100 {
// 		limit = 10
// 	}

// 	invoices, err := c.invoiceService.ListInvoices(ctx.Request.Context(), page, limit)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"code":    "00",
// 		"message": "Invoices retrieved successfully",
// 		"data":    invoices,
// 	})
// }

// // GetCustomerInvoices retrieves invoices for a specific customer
// func (c *InvoiceController) GetCustomerInvoices(ctx *gin.Context) {
// 	customerID := ctx.Param("customerId")
// 	pageStr := ctx.DefaultQuery("page", "1")
// 	limitStr := ctx.DefaultQuery("limit", "10")

// 	page, err := strconv.Atoi(pageStr)
// 	if err != nil || page < 1 {
// 		page = 1
// 	}

// 	limit, err := strconv.Atoi(limitStr)
// 	if err != nil || limit < 1 || limit > 100 {
// 		limit = 10
// 	}

// 	invoices, err := c.invoiceService.GetCustomerInvoices(ctx.Request.Context(), customerID, page, limit)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"code":    "00",
// 		"message": "Customer invoices retrieved successfully",
// 		"data":    invoices,
// 	})
// }

// // UpdatePaymentStatus updates the payment status of an invoice
// func (c *InvoiceController) UpdatePaymentStatus(ctx *gin.Context) {
// 	id := ctx.Param("id")

// 	var req struct {
// 		Status string `json:"status" binding:"required"`
// 	}

// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	status := model.PaymentStatus(req.Status)
// 	err := c.invoiceService.UpdatePaymentStatus(ctx.Request.Context(), id, status)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"code":    "00",
// 		"message": "Payment status updated successfully",
// 	})
// }
