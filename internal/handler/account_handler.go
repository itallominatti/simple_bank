package handler

import (
	"net/http"
	"simple_bank/internal/application"
	"simple_bank/internal/domain"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	service *application.AccountService
}

func NewAccountHandler(service *application.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

// POST /accounts
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida: " + err.Error()})
		return
	}

	account, err := h.service.CreateAccount(c.Request.Context(), req.OwnerName)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toAccountResponse(account))
}

// GET /accounts/:id
func (h *AccountHandler) GetAccount(c *gin.Context) {
	id := c.Param("id")

	account, err := h.service.GetAccount(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toAccountResponse(account))
}

// POST /accounts/:id/deposit
func (h *AccountHandler) Deposit(c *gin.Context) {
	var req AmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida: " + err.Error()})
		return
	}

	account, err := h.service.Deposit(c.Request.Context(), c.Param("id"), domain.Money(req.Amount))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toAccountResponse(account))
}

// POST /accounts/:id/withdraw
func (h *AccountHandler) Withdraw(c *gin.Context) {
	var req AmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida: " + err.Error()})
		return
	}

	account, err := h.service.Withdraw(c.Request.Context(), c.Param("id"), domain.Money(req.Amount))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toAccountResponse(account))
}
