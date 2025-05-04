package route

import (
	"github.com/gin-gonic/gin"

	"payment_service/api/controller"
)

// SetupRoutes configures all the API routes for the application
func SetupRoutes(r *gin.Engine, vnpayController *controller.VNPayController) {
	// API group
	api := r.Group("/api")

	// VNPay routes
	vnpay := api.Group("/vnpay")
	{
		vnpay.POST("/create-payment", vnpayController.CreatePayment)
		vnpay.GET("/return", vnpayController.HandleReturn)
		vnpay.POST("/ipn", vnpayController.HandleIPN)
		vnpay.POST("/query", vnpayController.QueryTransaction)
		vnpay.POST("/refund", vnpayController.RefundTransaction)
	}

	// Invoice routes
	invoices := api.Group("/invoices")
	{
		invoices.GET("/:id", vnpayController.GetInvoice)
		invoices.GET("/customer/:customerId", vnpayController.GetInvoicesByCustomer)
	}
}
