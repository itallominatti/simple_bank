CREATE TABLE IF NOT EXISTS accounts (
    id          UUID        PRIMARY KEY,                         
    owner_name  TEXT        NOT NULL,                            
    balance     BIGINT      NOT NULL DEFAULT 0 CHECK (balance >= 0), 
    created_at  TIMESTAMPTZ NOT NULL,                           
    updated_at  TIMESTAMPTZ NOT NULL                            
);


CREATE TABLE IF NOT EXISTS transactions (
    id          UUID        PRIMARY KEY,
    account_id  UUID        NOT NULL REFERENCES accounts(id),    
    type        TEXT        NOT NULL,                            
    amount      BIGINT      NOT NULL CHECK (amount > 0),         
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions (account_id);
