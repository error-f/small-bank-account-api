# Small Bank Account API

A simple API for managing bank accounts, with the following functionality:
- Create an account
- Add money to an account
- Transfer money between accounts

## Requirements
- Docker
- Docker Compose
- Go 1.22.4 (for running tests)

## How To

To spawn a PostgreSQL database and connect the app to it, you can configure the database settings in the `docker-compose` file and through environment variables:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`

### Run Tests
This command will spawn a container with a clean database and run tests against it:
```bash
make test
```

### Run Locally

#### With a Clean Database
This will remove data from previous runs, then start the database and the app:
```bash
make clean-run
```

#### With Saved Data (from previous runs)
This will preserve data between sequential runs of the app:
```bash
make run
```

### Make a Request

1. Run the app using the command:
   ```bash
   make run
   ```
2. Make requests. You can find the request specifications in the `open-api-spec.yaml` file. Examples:

- **Create an Account**:
  ```bash
  curl -X POST "http://localhost:8080/api/accounts"   -d '{"user_id": "123e4567-e89b-12d3-a456-426614174000", "currency": "USD"}'
  ```

- **Add Money to an Account**:
  ```bash
  curl -X POST "http://localhost:8080/api/accounts/add-money"   -d '{"user_id": "123e4567-e89b-12d3-a456-426614174000", "account_id": "d1f75516-0526-4c07-bd5b-eebb0feec2a0", "currency": "USD", "amount": 1}'
  ```

- **Transfer Money Between Accounts**:
  ```bash
  curl -X POST "http://localhost:8080/api/accounts/transfer-money"   -d '{"user_id": "123e4567-e89b-12d3-a456-426614174000", "source_account_id": "d1f75516-0526-4c07-bd5b-eebb0feec2a0", "target_account_id": "d8b4dcff-8b69-4ce7-8db4-e3d9ae412dc1", "currency": "USD", "amount": 1}'
  ```

## Potential improvements

- Timezones for timestamps in database
- Tracing, logging
- Comprehensive error responses
- Log failed transactions
- Validate request/response schema
- More APIs:
    - Transfers in different currencies
    - Get balance
    - Get all accounts by user
    - ...
