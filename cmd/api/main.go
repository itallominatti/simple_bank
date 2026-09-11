package main

import (
	"context"
	"log"
	"os"

	"simple_bank/internal/application"
	"simple_bank/internal/handler"
	"simple_bank/internal/infrastructure/postgres"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("a variável de ambiente DATABASE_URL é obrigatória")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := postgres.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	accountRepo := postgres.NewAccountRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	uow := postgres.NewUnitOfWork(db)

	accountService := application.NewAccountService(accountRepo, transactionRepo, uow)

	accountHandler := handler.NewAccountHandler(accountService)
	router := handler.NewRouter(accountHandler)

	log.Printf("API rodando na porta %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
