package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Server struct {
	db *sql.DB
}

func NewServer() (*Server, error) {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "5432"),
		getEnvOrDefault("DB_USER", "postgres"),
		getEnvOrDefault("DB_PASSWORD", "password"),
		getEnvOrDefault("DB_NAME", "testdb"),
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Return new Server instance
	return &Server{db: db}, nil
}

func (s *Server) StartServer() {
	defer s.db.Close()

	http.HandleFunc("/api/accounts", s.CreateAccountHandler)
	http.HandleFunc("/api/accounts/add-money", s.AddMoneyHandler)
	http.HandleFunc("/api/accounts/transfer-money", s.TransferMoneyHandler)

	log.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

type TransferMoneyRequest struct {
	UserID          string  `json:"user_id"`
	SourceAccountID string  `json:"source_account_id"`
	TargetAccountID string  `json:"target_account_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
}

type TransferMoneyResponse struct {
	UserID            string  `json:"user_id"`
	SourceAccountID   string  `json:"source_account_id"`
	SourceTotalAmount float64 `json:"source_total_amount"`
	SourceCurrency    string  `json:"source_currency"`
	TargetAccountID   string  `json:"target_account_id"`
	TargetTotalAmount float64 `json:"target_total_amount"`
	TargetCurrency    string  `json:"target_currency"`
}

func (s *Server) TransferMoneyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData TransferMoneyRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	tx, err := s.db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	// Log a transaction
	query := `INSERT INTO transactions (user_id, source_account_id, target_account_id, amount, currency) VALUES ($1, $2, $3, $4, $5)`
	if _, err := tx.Exec(query, requestData.UserID, requestData.SourceAccountID, requestData.TargetAccountID, requestData.Amount, requestData.Currency); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to log transaction", http.StatusInternalServerError)
		return
	}

	// Lower the account balance of source account
	query = `UPDATE accounts SET amount = amount - $1 WHERE account_id = $2 AND currency = $3 RETURNING amount, currency`
	var sourceAccountTotalAmount float64
	var sourceCurrency string
	if err := tx.QueryRow(query, requestData.Amount, requestData.SourceAccountID, requestData.Currency).Scan(&sourceAccountTotalAmount, &sourceCurrency); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update account balance", http.StatusInternalServerError)
		return
	}

	// Increase the account balance of target account
	query = `UPDATE accounts SET amount = amount + $1 WHERE account_id = $2 AND currency = $3 RETURNING amount, currency`
	var targetAccountTotalAmount float64
	var targetCurrency string
	if err := tx.QueryRow(query, requestData.Amount, requestData.TargetAccountID, requestData.Currency).Scan(&targetAccountTotalAmount, &targetCurrency); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update account balance", http.StatusInternalServerError)
		return
	}

	// If all inserts were successful, commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	var responseData TransferMoneyResponse
	responseData.UserID = requestData.UserID
	responseData.SourceAccountID = requestData.SourceAccountID
	responseData.SourceCurrency = sourceCurrency
	responseData.SourceTotalAmount = sourceAccountTotalAmount
	responseData.TargetAccountID = requestData.TargetAccountID
	responseData.TargetCurrency = targetCurrency
	responseData.TargetTotalAmount = targetAccountTotalAmount

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseData)
}

type AddMoneyRequest struct {
	UserID    string  `json:"user_id"`
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

type AddMoneyResponse struct {
	UserID      string  `json:"user_id"`
	AccountID   string  `json:"account_id"`
	TotalAmount float64 `json:"total_amount"`
	Currency    string  `json:"currency"`
}

func (s *Server) AddMoneyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData AddMoneyRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	tx, err := s.db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	// Log a transaction
	query := `INSERT INTO transactions (user_id, source_account_id, amount, currency) VALUES ($1, $2, $3, $4)`
	if _, err := tx.Exec(query, requestData.UserID, requestData.AccountID, requestData.Amount, requestData.Currency); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to log transaction", http.StatusInternalServerError)
		return
	}

	// Update the account balance
	query = `UPDATE accounts SET amount = amount + $1 WHERE account_id = $2 AND currency = $3 RETURNING amount`
	var totalAmount float64
	if err := tx.QueryRow(query, requestData.Amount, requestData.AccountID, requestData.Currency).Scan(&totalAmount); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update account balance", http.StatusInternalServerError)
		return
	}

	// If all inserts were successful, commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	var responseData AddMoneyResponse
	responseData.UserID = requestData.UserID
	responseData.AccountID = requestData.AccountID
	responseData.Currency = requestData.Currency
	responseData.TotalAmount = totalAmount

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseData)
}

type CreateAccountRequest struct {
	UserID   string `json:"user_id"`
	Currency string `json:"currency"`
}

type CreateAccountResponse struct {
	UserID    string `json:"user_id"`
	AccountID string `json:"account_id"`
	Currency  string `json:"currency"`
	CreatedAt string `json:"created_at"`
}

func (s *Server) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Insert the account into the accounts table
	var accountID string
	var createdAt string
	query := `INSERT INTO accounts (user_id, currency) VALUES ($1, $2) RETURNING account_id, created_at`
	if err := s.db.QueryRow(query, requestData.UserID, requestData.Currency).Scan(&accountID, &createdAt); err != nil {
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	var responseData CreateAccountResponse
	responseData.UserID = requestData.UserID
	responseData.AccountID = accountID
	responseData.Currency = requestData.Currency
	responseData.CreatedAt = createdAt

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseData)
}
