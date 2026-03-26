-- Migration: Create FIBO deals table for Fortune 5 compliance
-- This migration creates the core deals table for financial instrument management
-- with compliance controls (KYC, AML, SOX verification)

-- Create deals table
CREATE TABLE IF NOT EXISTS deals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(500) NOT NULL,
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD' CHECK (currency IN ('USD', 'EUR', 'GBP', 'JPY', 'CAD', 'AUD', 'CHF', 'CNY', 'INR', 'SGD')),
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'proposed', 'negotiating', 'approved', 'executed', 'settled', 'closed', 'rejected')),
    domain VARCHAR(100) NOT NULL CHECK (domain IN ('equity', 'fixed_income', 'derivatives', 'commodities', 'fx', 'structured', 'other')),
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    kcy_verified BOOLEAN DEFAULT false,
    aml_verified BOOLEAN DEFAULT false,
    sox_verified BOOLEAN DEFAULT false,
    counterparty_id UUID,
    counterparty_name VARCHAR(500),
    deal_date DATE,
    settlement_date DATE,
    maturity_date DATE,
    description TEXT,
    risk_rating VARCHAR(10) CHECK (risk_rating IN ('AAA', 'AA', 'A', 'BBB', 'BB', 'B', 'CCC', 'CC', 'C', 'D')),
    internal_reference VARCHAR(100),
    external_reference VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT deal_dates CHECK (deal_date IS NULL OR settlement_date IS NULL OR deal_date <= settlement_date),
    CONSTRAINT settlement_maturity CHECK (settlement_date IS NULL OR maturity_date IS NULL OR settlement_date <= maturity_date)
);

-- Indexes for optimal query performance
CREATE INDEX IF NOT EXISTS idx_deals_status ON deals(status);
CREATE INDEX IF NOT EXISTS idx_deals_created_by ON deals(created_by);
CREATE INDEX IF NOT EXISTS idx_deals_assigned_to ON deals(assigned_to);
CREATE INDEX IF NOT EXISTS idx_deals_created_at ON deals(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_deals_domain ON deals(domain);
CREATE INDEX IF NOT EXISTS idx_deals_currency ON deals(currency);
CREATE INDEX IF NOT EXISTS idx_deals_deal_date ON deals(deal_date DESC);
CREATE INDEX IF NOT EXISTS idx_deals_settlement_date ON deals(settlement_date DESC);
CREATE INDEX IF NOT EXISTS idx_deals_kcy_aml_sox ON deals(kcy_verified, aml_verified, sox_verified);
CREATE INDEX IF NOT EXISTS idx_deals_counterparty ON deals(counterparty_id, counterparty_name);

-- Table comments
COMMENT ON TABLE deals IS 'Core deals table for financial instrument management with compliance tracking';
COMMENT ON COLUMN deals.amount_cents IS 'Deal amount in cents (to avoid floating-point precision issues)';
COMMENT ON COLUMN deals.currency IS 'ISO 4217 currency code';
COMMENT ON COLUMN deals.status IS 'Deal lifecycle status from draft to closed';
COMMENT ON COLUMN deals.domain IS 'Financial instrument type (equity, fixed income, etc.)';
COMMENT ON COLUMN deals.kcy_verified IS 'Know Your Customer verification status';
COMMENT ON COLUMN deals.aml_verified IS 'Anti-Money Laundering verification status';
COMMENT ON COLUMN deals.sox_verified IS 'Sarbanes-Oxley compliance verification';
COMMENT ON COLUMN deals.risk_rating IS 'Credit risk rating (S&P/Moody''s equivalent)';

-- Add trigger to update updated_at automatically
CREATE OR REPLACE FUNCTION update_deals_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_deals_updated_at ON deals;
CREATE TRIGGER trigger_deals_updated_at
BEFORE UPDATE ON deals
FOR EACH ROW
EXECUTE FUNCTION update_deals_updated_at();
