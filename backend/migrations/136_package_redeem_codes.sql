ALTER TABLE redeem_codes
    ADD COLUMN IF NOT EXISTS purchased_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS purchase_order_id BIGINT NULL REFERENCES payment_orders(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS purchase_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS purchase_pay_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
	ADD COLUMN IF NOT EXISTS purchase_currency VARCHAR(10) NULL;

CREATE INDEX IF NOT EXISTS idx_redeem_codes_purchased_by ON redeem_codes(purchased_by);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_redeem_codes_purchase_order_id
    ON redeem_codes(purchase_order_id)
    WHERE purchase_order_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS package_redeem_sale_rebates (
    id BIGSERIAL PRIMARY KEY,
    redeem_code_id BIGINT NOT NULL REFERENCES redeem_codes(id) ON DELETE CASCADE,
    purchase_order_id BIGINT NULL REFERENCES payment_orders(id) ON DELETE SET NULL,
    purchaser_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redeemer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    base_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    rebate_rate DECIMAL(10,4) NOT NULL DEFAULT 0,
    rebate_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    reason VARCHAR(100) NULL,
    error TEXT NULL,
    applied_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_package_redeem_sale_rebate_code
    ON package_redeem_sale_rebates(redeem_code_id);
CREATE INDEX IF NOT EXISTS idx_package_redeem_sale_rebates_purchaser
    ON package_redeem_sale_rebates(purchaser_id);
CREATE INDEX IF NOT EXISTS idx_package_redeem_sale_rebates_redeemer
    ON package_redeem_sale_rebates(redeemer_id);
CREATE INDEX IF NOT EXISTS idx_package_redeem_sale_rebates_status
    ON package_redeem_sale_rebates(status);

CREATE TABLE IF NOT EXISTS package_redeem_sale_rebate_ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    sale_rebate_id BIGINT NOT NULL REFERENCES package_redeem_sale_rebates(id) ON DELETE CASCADE,
    purchaser_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redeemer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redeem_code_id BIGINT NOT NULL REFERENCES redeem_codes(id) ON DELETE CASCADE,
    purchase_order_id BIGINT NULL REFERENCES payment_orders(id) ON DELETE SET NULL,
    amount DECIMAL(20,8) NOT NULL,
    currency VARCHAR(10) NULL,
    frozen_until TIMESTAMPTZ NULL,
    source_type VARCHAR(50) NOT NULL DEFAULT 'package_redeem_sale_rebate',
    source_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_package_redeem_sale_rebate_ledger_source
    ON package_redeem_sale_rebate_ledger_entries(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_package_redeem_sale_rebate_ledger_purchaser
    ON package_redeem_sale_rebate_ledger_entries(purchaser_id);
CREATE INDEX IF NOT EXISTS idx_package_redeem_sale_rebate_ledger_redeemer
    ON package_redeem_sale_rebate_ledger_entries(redeemer_id);
