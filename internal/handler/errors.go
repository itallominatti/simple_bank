package handler

import (
	"errors"
	"log"
	"net/http"
	"simple_bank/internal/domain"

	"github.com/gin-gonic/gin"
)

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

	case errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrInvalidOwnerName),
		errors.Is(err, domain.ErrSameAccountTransfer):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, domain.ErrInsufficientFunds):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})

	default:
		log.Printf("erro inesperado: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno do servidor"})
	}
}
