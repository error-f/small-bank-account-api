CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE currency AS ENUM ('USD');

CREATE TABLE accounts (
    account_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    currency currency NOT NULL,
    amount NUMERIC(10, 2) CHECK (amount >= 0) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_accounts (
    user_id UUID NOT NULL,
    account_id UUID NOT NULL,
    PRIMARY KEY (user_id, account_id),
    FOREIGN KEY (account_id) REFERENCES accounts(account_id) ON DELETE CASCADE
);

CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    source_account_id UUID NOT NULL,
    target_account_id UUID,
    amount NUMERIC(10, 2) CHECK (amount > 0) NOT NULL,
    currency currency NOT NULL,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK (target_account_id != source_account_id)
);

CREATE INDEX idx_account_id ON transactions(source_account_id);
