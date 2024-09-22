package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

// TestMain runs before all tests and sets up the test environment.
func TestMain(m *testing.M) {
	// Set up database connection
	connStr := "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable"
	var err error
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Set global DB variable
	db = testDB

	// Run the tests
	code := m.Run()

	// Close the connection
	testDB.Close()

	// Exit with the result of the tests
	os.Exit(code)
}
func TestCreateAccountHandler(t *testing.T) {
	tx, terr := testDB.Begin()
	if terr != nil {
		t.Fatalf("Failed to begin transaction: %v", terr)
	}
	defer tx.Rollback()

	// Setup request and recorder
	reqBody := CreateAccountRequest{UserID: "123e4567-e89b-12d3-a456-426614174000", Currency: "USD"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Call the handler
	createAccountHandler(w, req)

	// Assert the response
	assert.Equal(t, http.StatusCreated, w.Code, "expected status 201")
	var response CreateAccountResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", response.UserID)
	assert.Equal(t, "USD", response.Currency)
}

func createAccount(t *testing.T, user_id string, currency string) string {
	reqBody := CreateAccountRequest{UserID: user_id, Currency: currency}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()
	createAccountHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var response CreateAccountResponse
	json.NewDecoder(w.Body).Decode(&response)
	return response.AccountID
}

func TestAddMoneyHandler(t *testing.T) {
	tx, terr := testDB.Begin()
	if terr != nil {
		t.Fatalf("Failed to begin transaction: %v", terr)
	}
	defer tx.Rollback()

	// create an account
	accountID := createAccount(t, "123e4567-e89b-12d3-a456-426614174000", "USD")

	// Setup request and recorder
	reqBody := AddMoneyRequest{UserID: "123e4567-e89b-12d3-a456-426614174000", AccountID: accountID, Amount: 100.0, Currency: "USD"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/accounts/add-money", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Call the handler
	addMoneyHandler(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code, "expected status 200")
	var response AddMoneyResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", response.UserID)
	assert.Equal(t, accountID, response.AccountID)
	assert.Equal(t, 100.0, response.TotalAmount)
	assert.Equal(t, "USD", response.Currency)
}

func addMoney(t *testing.T, user_id string, account_id string, amount float64, currency string) {
	reqBody := AddMoneyRequest{UserID: user_id, AccountID: account_id, Amount: amount, Currency: currency}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/accounts/add-money", bytes.NewReader(body))
	w := httptest.NewRecorder()
	addMoneyHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "expected status 200")
	var response AddMoneyResponse
	json.NewDecoder(w.Body).Decode(&response)
}

func TestTransferMoneyHandler(t *testing.T) {
	tx, terr := testDB.Begin()
	if terr != nil {
		t.Fatalf("Failed to begin transaction: %v", terr)
	}
	defer tx.Rollback()

	sourceAccountID := createAccount(t, "123e4567-e89b-12d3-a456-426614174000", "USD")
	targetAccountID := createAccount(t, "123e4567-e89b-12d3-a456-426614174000", "USD")
	addMoney(t, "123e4567-e89b-12d3-a456-426614174000", sourceAccountID, 150.0, "USD")

	// Setup request and recorder
	reqBody := TransferMoneyRequest{UserID: "123e4567-e89b-12d3-a456-426614174000", SourceAccountID: sourceAccountID, TargetAccountID: targetAccountID, Amount: 50.0, Currency: "USD"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/accounts/transfer-money", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Call the handler
	transferMoneyHandler(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code, "expected status 200")
	var response TransferMoneyResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", response.UserID)
	assert.Equal(t, sourceAccountID, response.SourceAccountID)
	assert.Equal(t, targetAccountID, response.TargetAccountID)
	assert.Equal(t, 100.0, response.SourceTotalAmount)
	assert.Equal(t, 50.0, response.TargetTotalAmount)
}
