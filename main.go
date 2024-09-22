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

var db *sql.DB

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	// Connect to PostgreSQL
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "testdb"),
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// schema check
	http.HandleFunc("/api/accounts", createAccountHandler)
	http.HandleFunc("/api/accounts/add-money", addMoneyHandler)
	http.HandleFunc("/api/accounts/transfer-money", transferMoneyHandler)

	log.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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

func transferMoneyHandler(w http.ResponseWriter, r *http.Request) {
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

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	// Log a transaction
	query := `INSERT INTO transactions (user_id, source_account_id, target_account_id, amount, currency) VALUES ($1, $2, $3, $4, $5)`
	if _, err := tx.Exec(query, requestData.UserID, requestData.SourceAccountID, requestData.TargetAccountID, requestData.Amount, requestData.Currency); err != nil {
		log.Println("Failed to log transaction:", err)
		tx.Rollback()
		http.Error(w, "Failed to log transaction", http.StatusInternalServerError)
		return
	}

	// Lower the account balance of source account
	query = `UPDATE accounts SET amount = amount - $1 WHERE account_id = $2 AND currency = $3 RETURNING amount, currency`
	var sourceAccountTotalAmount float64
	var sourceCurrency string
	if err := tx.QueryRow(query, requestData.Amount, requestData.SourceAccountID, requestData.Currency).Scan(&sourceAccountTotalAmount, &sourceCurrency); err != nil {
		log.Println("Failed to decrease account balance:", err)
		tx.Rollback()
		http.Error(w, "Failed to update account balance", http.StatusInternalServerError)
		return
	}

	// Increase the account balance of target account
	query = `UPDATE accounts SET amount = amount + $1 WHERE account_id = $2 AND currency = $3 RETURNING amount, currency`
	var targetAccountTotalAmount float64
	var targetCurrency string
	if err := tx.QueryRow(query, requestData.Amount, requestData.TargetAccountID, requestData.Currency).Scan(&targetAccountTotalAmount, &targetCurrency); err != nil {
		log.Println("Failed to increase account balance:", err)
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

func addMoneyHandler(w http.ResponseWriter, r *http.Request) {
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

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	// Log a transaction
	query := `INSERT INTO transactions (user_id, source_account_id, amount, currency) VALUES ($1, $2, $3, $4)`
	if _, err := tx.Exec(query, requestData.UserID, requestData.AccountID, requestData.Amount, requestData.Currency); err != nil {
		log.Println("Failed to log transaction:", err)
		tx.Rollback()
		http.Error(w, "Failed to log transaction", http.StatusInternalServerError)
		return
	}

	// Update the account balance
	query = `UPDATE accounts SET amount = amount + $1 WHERE account_id = $2 AND currency = $3 RETURNING amount`
	var totalAmount float64
	if err := tx.QueryRow(query, requestData.Amount, requestData.AccountID, requestData.Currency).Scan(&totalAmount); err != nil {
		log.Println("Failed to update account balance:", err)
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

func createAccountHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := db.QueryRow(query, requestData.UserID, requestData.Currency).Scan(&accountID, &createdAt); err != nil {
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
