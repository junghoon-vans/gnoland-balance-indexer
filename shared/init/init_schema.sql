-- Create blocks table
CREATE TABLE IF NOT EXISTS blocks (
    id SERIAL PRIMARY KEY,
    hash VARCHAR(64) UNIQUE NOT NULL,
    height BIGINT UNIQUE NOT NULL,
    time TIMESTAMP NOT NULL,
    num_txs INTEGER NOT NULL DEFAULT 0,
    total_txs INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    hash VARCHAR(64) UNIQUE NOT NULL,
    index INTEGER NOT NULL,
    block_height BIGINT NOT NULL,
    success BOOLEAN NOT NULL DEFAULT false,
    gas_wanted BIGINT DEFAULT 0,
    gas_used BIGINT DEFAULT 0,
    memo TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create transaction messages table
CREATE TABLE IF NOT EXISTS transaction_msgs (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL,
    route VARCHAR(255),
    type_url VARCHAR(255),
    value JSONB,
    FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
);

-- Create transaction events table
CREATE TABLE IF NOT EXISTS transaction_events (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL,
    type VARCHAR(255) NOT NULL,
    func VARCHAR(255),
    pkg_path VARCHAR(255),
    FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
);

-- Create transaction event attributes table
CREATE TABLE IF NOT EXISTS transaction_event_attrs (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT,
    FOREIGN KEY (event_id) REFERENCES transaction_events(id) ON DELETE CASCADE
);

-- Create token balances table
CREATE TABLE IF NOT EXISTS token_balances (
    id SERIAL PRIMARY KEY,
    address VARCHAR(40) NOT NULL,
    token_path VARCHAR(255) NOT NULL,
    amount NUMERIC(78,0) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create token transfers table
CREATE TABLE IF NOT EXISTS token_transfers (
    id SERIAL PRIMARY KEY,
    block_height BIGINT NOT NULL,
    tx_hash VARCHAR(64) NOT NULL,
    event_id INTEGER NOT NULL,
    from_address VARCHAR(40),
    to_address VARCHAR(40),
    token_path VARCHAR(255) NOT NULL,
    amount NUMERIC(78,0) NOT NULL,
    transfer_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    -- Add unique constraint to prevent duplicate processing
    CONSTRAINT unique_transfer_event UNIQUE (tx_hash, event_id)
);

-- Create processed events table for idempotency tracking
CREATE TABLE IF NOT EXISTS processed_events (
    id SERIAL PRIMARY KEY,
    event_identifier VARCHAR(255) UNIQUE NOT NULL, -- tx_hash-event_id format
    tx_hash VARCHAR(64) NOT NULL,
    event_id INTEGER NOT NULL,
    block_height BIGINT NOT NULL,
    processor_instance VARCHAR(100),
    processed_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_blocks_height ON blocks(height);
CREATE INDEX IF NOT EXISTS idx_transactions_block_height ON transactions(block_height);
CREATE INDEX IF NOT EXISTS idx_transactions_hash ON transactions(hash);
CREATE INDEX IF NOT EXISTS idx_transaction_msgs_transaction_id ON transaction_msgs(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transaction_events_transaction_id ON transaction_events(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transaction_event_attrs_event_id ON transaction_event_attrs(event_id);

-- Create unique constraint for token balances
CREATE UNIQUE INDEX IF NOT EXISTS idx_token_balance_address_token
ON token_balances (address, token_path);

-- Create indexes for token transfers
CREATE INDEX IF NOT EXISTS idx_token_transfers_block_height ON token_transfers(block_height);
CREATE INDEX IF NOT EXISTS idx_token_transfers_tx_hash ON token_transfers(tx_hash);
CREATE INDEX IF NOT EXISTS idx_token_transfers_from_address ON token_transfers(from_address);
CREATE INDEX IF NOT EXISTS idx_token_transfers_to_address ON token_transfers(to_address);
CREATE INDEX IF NOT EXISTS idx_token_transfers_token_path ON token_transfers(token_path);
CREATE INDEX IF NOT EXISTS idx_token_transfers_transfer_type ON token_transfers(transfer_type);

-- Create indexes for token balances
CREATE INDEX IF NOT EXISTS idx_token_balances_address ON token_balances(address);
CREATE INDEX IF NOT EXISTS idx_token_balances_token_path ON token_balances(token_path);
CREATE INDEX IF NOT EXISTS idx_token_balances_amount ON token_balances(amount);

-- Create indexes for processed events
CREATE INDEX IF NOT EXISTS idx_processed_events_tx_hash ON processed_events(tx_hash);
CREATE INDEX IF NOT EXISTS idx_processed_events_block_height ON processed_events(block_height);
CREATE INDEX IF NOT EXISTS idx_processed_events_processor_instance ON processed_events(processor_instance);
CREATE INDEX IF NOT EXISTS idx_processed_events_processed_at ON processed_events(processed_at);

-- Add updated_at trigger for blocks
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_blocks_updated_at BEFORE UPDATE ON blocks
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_token_balances_updated_at BEFORE UPDATE ON token_balances
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
