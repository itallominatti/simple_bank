package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *AccountHandler) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	accounts := router.Group("/accounts")
	{
		accounts.POST("", h.CreateAccount)                    // POST /accounts
		accounts.GET("/:id", h.GetAccount)                    // GET  /accounts/{id}
		accounts.POST("/:id/deposit", h.Deposit)              // POST /accounts/{id}/deposit
		accounts.POST("/:id/withdraw", h.Withdraw)            // POST /accounts/{id}/withdraw
		accounts.GET("/:id/transactions", h.ListTransactions) // GET  /accounts/{id}/transactions
	}

	router.POST("/transfers", h.Transfer) // POST /transfers

	return router
}
