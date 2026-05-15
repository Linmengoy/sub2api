package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type packageRedeemSaleRebateRepository struct {
	client *dbent.Client
}

func NewPackageRedeemSaleRebateRepository(client *dbent.Client) service.PackageRedeemSaleRebateRepository {
	return &packageRedeemSaleRebateRepository{client: client}
}

func (r *packageRedeemSaleRebateRepository) CreatePendingIfNotExists(ctx context.Context, input service.PackageRedeemSaleRebate) (*service.PackageRedeemSaleRebate, error) {
	return r.createIfNotExists(ctx, input, service.PackageRedeemSaleRebateStatusPending, "")
}

func (r *packageRedeemSaleRebateRepository) CreateSkippedIfNotExists(ctx context.Context, input service.PackageRedeemSaleRebate, status, reason string) (*service.PackageRedeemSaleRebate, error) {
	return r.createIfNotExists(ctx, input, status, reason)
}

func (r *packageRedeemSaleRebateRepository) GetUserSummary(ctx context.Context, userID int64) (*service.UserPackageRedeemSummary, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
SELECT
    COUNT(rc.id)::bigint,
    COUNT(rc.id) FILTER (WHERE rc.status = 'unused')::bigint,
    COUNT(rc.id) FILTER (WHERE rc.status = 'used')::bigint,
    COALESCE(SUM(rc.purchase_pay_amount), 0)::double precision,
    COALESCE(SUM(pr.rebate_amount) FILTER (WHERE pr.status = 'applied'), 0)::double precision,
    COALESCE(ua.aff_quota, 0)::double precision,
    COALESCE(ua.aff_frozen_quota, 0)::double precision,
    COALESCE(ua.aff_history_quota, 0)::double precision
FROM users u
LEFT JOIN redeem_codes rc ON rc.purchased_by = u.id AND rc.type = 'subscription'
LEFT JOIN package_redeem_sale_rebates pr ON pr.redeem_code_id = rc.id
LEFT JOIN user_affiliates ua ON ua.user_id = u.id
WHERE u.id = $1
GROUP BY ua.aff_quota, ua.aff_frozen_quota, ua.aff_history_quota`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user package redeem summary: %w", err)
	}
	defer func() { _ = rows.Close() }()
	summary := &service.UserPackageRedeemSummary{}
	if rows.Next() {
		if err := rows.Scan(&summary.TotalCodes, &summary.UnusedCodes, &summary.UsedCodes, &summary.TotalPurchasePayAmount, &summary.TotalRebateAmount, &summary.AvailableQuota, &summary.FrozenQuota, &summary.HistoryQuota); err != nil {
			return nil, err
		}
	}
	return summary, rows.Close()
}

func (r *packageRedeemSaleRebateRepository) GetAdminSummary(ctx context.Context) (*service.AdminPackageRedeemSummary, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
SELECT
    COUNT(rc.id)::bigint,
    COUNT(rc.id) FILTER (WHERE rc.status = 'unused')::bigint,
    COUNT(rc.id) FILTER (WHERE rc.status = 'used')::bigint,
    COALESCE(SUM(rc.purchase_pay_amount), 0)::double precision,
    COALESCE(SUM(pr.rebate_amount) FILTER (WHERE pr.status = 'applied'), 0)::double precision,
    COUNT(pr.id) FILTER (WHERE pr.status = 'applied')::bigint,
    COUNT(pr.id) FILTER (WHERE pr.status = 'pending')::bigint,
    COUNT(pr.id) FILTER (WHERE pr.status = 'failed')::bigint
FROM redeem_codes rc
LEFT JOIN package_redeem_sale_rebates pr ON pr.redeem_code_id = rc.id
WHERE rc.type = 'subscription' AND rc.purchased_by IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("query admin package redeem summary: %w", err)
	}
	defer func() { _ = rows.Close() }()
	summary := &service.AdminPackageRedeemSummary{}
	if rows.Next() {
		if err := rows.Scan(&summary.TotalCodes, &summary.UnusedCodes, &summary.UsedCodes, &summary.TotalPurchasePayAmount, &summary.TotalRebateAmount, &summary.AppliedRebateCount, &summary.PendingRebateCount, &summary.FailedRebateCount); err != nil {
			return nil, err
		}
	}
	return summary, rows.Close()
}

func (r *packageRedeemSaleRebateRepository) createIfNotExists(ctx context.Context, input service.PackageRedeemSaleRebate, status, reason string) (*service.PackageRedeemSaleRebate, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
INSERT INTO package_redeem_sale_rebates (
    redeem_code_id, purchase_order_id, purchaser_id, redeemer_id,
    base_amount, rebate_rate, rebate_amount, currency, status, reason,
    applied_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9::varchar, NULLIF($10, ''),
        CASE WHEN $9::text = 'applied' THEN NOW() ELSE NULL END, NOW(), NOW())
ON CONFLICT (redeem_code_id) DO UPDATE SET updated_at = package_redeem_sale_rebates.updated_at
RETURNING id, redeem_code_id, purchase_order_id, purchaser_id, redeemer_id,
          base_amount::double precision, rebate_rate::double precision, rebate_amount::double precision,
          COALESCE(currency, ''), status, COALESCE(reason, ''), COALESCE(error, ''), applied_at, created_at, updated_at`,
		input.RedeemCodeID,
		nullableInt64Value(input.PurchaseOrderID),
		input.PurchaserID,
		input.RedeemerID,
		input.BaseAmount,
		input.RebateRate,
		input.RebateAmount,
		strings.TrimSpace(input.Currency),
		status,
		reason,
	)
	if err != nil {
		return nil, fmt.Errorf("create package redeem sale rebate: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanPackageRedeemSaleRebateRows(rows)
}

func (r *packageRedeemSaleRebateRepository) MarkApplied(ctx context.Context, id int64, rebateRate, rebateAmount float64, currency string) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.ExecContext(ctx, `
UPDATE package_redeem_sale_rebates
SET status = $2, reason = NULL, error = NULL, rebate_rate = $3, rebate_amount = $4, currency = NULLIF($5, ''), applied_at = NOW(), updated_at = NOW()
WHERE id = $1 AND status IN ($6, $7)`, id, service.PackageRedeemSaleRebateStatusApplied, rebateRate, rebateAmount, strings.TrimSpace(currency), service.PackageRedeemSaleRebateStatusPending, service.PackageRedeemSaleRebateStatusFailed)
	if err != nil {
		return fmt.Errorf("mark package redeem sale rebate applied: %w", err)
	}
	return nil
}

func (r *packageRedeemSaleRebateRepository) MarkSkipped(ctx context.Context, id int64, status, reason string, rebateRate, rebateAmount float64, currency string) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.ExecContext(ctx, `
UPDATE package_redeem_sale_rebates
SET status = $2, reason = NULLIF($3, ''), error = NULL, rebate_rate = $4, rebate_amount = $5, currency = NULLIF($6, ''), updated_at = NOW()
WHERE id = $1 AND status IN ($7, $8)`, id, status, reason, rebateRate, rebateAmount, strings.TrimSpace(currency), service.PackageRedeemSaleRebateStatusPending, service.PackageRedeemSaleRebateStatusFailed)
	if err != nil {
		return fmt.Errorf("mark package redeem sale rebate skipped: %w", err)
	}
	return nil
}

func (r *packageRedeemSaleRebateRepository) MarkFailed(ctx context.Context, id int64, rebateRate, rebateAmount float64, currency string, cause error) error {
	client := clientFromContext(ctx, r.client)
	errText := ""
	if cause != nil {
		errText = cause.Error()
	}
	_, err := client.ExecContext(ctx, `
UPDATE package_redeem_sale_rebates
SET status = $2, error = NULLIF($3, ''), rebate_rate = $4, rebate_amount = $5, currency = NULLIF($6, ''), updated_at = NOW()
WHERE id = $1 AND status = $7`, id, service.PackageRedeemSaleRebateStatusFailed, errText, rebateRate, rebateAmount, strings.TrimSpace(currency), service.PackageRedeemSaleRebateStatusPending)
	if err != nil {
		return fmt.Errorf("mark package redeem sale rebate failed: %w", err)
	}
	return nil
}

func scanPackageRedeemSaleRebateRows(rows *sql.Rows) (*service.PackageRedeemSaleRebate, error) {
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrRedeemCodeNotFound
	}
	var out service.PackageRedeemSaleRebate
	var purchaseOrderID sql.NullInt64
	var appliedAt sql.NullTime
	if err := rows.Scan(
		&out.ID,
		&out.RedeemCodeID,
		&purchaseOrderID,
		&out.PurchaserID,
		&out.RedeemerID,
		&out.BaseAmount,
		&out.RebateRate,
		&out.RebateAmount,
		&out.Currency,
		&out.Status,
		&out.Reason,
		&out.Error,
		&appliedAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if purchaseOrderID.Valid {
		out.PurchaseOrderID = &purchaseOrderID.Int64
	}
	if appliedAt.Valid {
		out.AppliedAt = &appliedAt.Time
	}
	return &out, rows.Close()
}

type packageRedeemSaleLedgerRepository struct {
	client *dbent.Client
}

func NewPackageRedeemSaleLedgerRepository(client *dbent.Client) service.PackageRedeemSaleLedgerRepository {
	return &packageRedeemSaleLedgerRepository{client: client}
}

func (r *packageRedeemSaleLedgerRepository) CreditSaleRebate(ctx context.Context, input service.CreditPackageRedeemSaleRebateInput) (bool, error) {
	if input.RebateAmount <= 0 {
		return false, nil
	}
	client := clientFromContext(ctx, r.client)
	tx, err := client.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin package redeem sale rebate ledger tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	txClient := tx.Client()
	sourceID := packageRedeemSaleRebateSourceID(input.RedeemCodeID)
	frozenUntil := any(nil)
	if input.FreezeHours > 0 {
		frozenUntil = time.Now().Add(time.Duration(input.FreezeHours) * time.Hour)
	}
	res, err := txClient.ExecContext(txCtx, `
INSERT INTO package_redeem_sale_rebate_ledger_entries (
    sale_rebate_id, purchaser_id, redeemer_id, redeem_code_id, purchase_order_id,
    amount, currency, frozen_until, source_type, source_id, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, 'package_redeem_sale_rebate', $9, NOW())
ON CONFLICT (source_type, source_id) DO NOTHING`,
		input.SaleRebateID,
		input.PurchaserID,
		input.RedeemerID,
		input.RedeemCodeID,
		nullableInt64Value(input.PurchaseOrderID),
		input.RebateAmount,
		strings.TrimSpace(input.Currency),
		frozenUntil,
		sourceID,
	)
	if err != nil {
		return false, fmt.Errorf("insert package redeem sale rebate ledger: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return false, tx.Commit()
	}
	if _, err = ensureUserAffiliateWithClient(txCtx, txClient, input.PurchaserID); err != nil {
		return false, err
	}
	var updateSQL string
	if input.FreezeHours > 0 {
		updateSQL = "UPDATE user_affiliates SET aff_frozen_quota = aff_frozen_quota + $1, aff_history_quota = aff_history_quota + $1, updated_at = NOW() WHERE user_id = $2"
	} else {
		updateSQL = "UPDATE user_affiliates SET aff_quota = aff_quota + $1, aff_history_quota = aff_history_quota + $1, updated_at = NOW() WHERE user_id = $2"
	}
	if _, err = txClient.ExecContext(txCtx, updateSQL, input.RebateAmount, input.PurchaserID); err != nil {
		return false, fmt.Errorf("credit package redeem sale rebate quota: %w", err)
	}
	return true, tx.Commit()
}

func (r *packageRedeemSaleLedgerRepository) GetCreditedAmountByRedeemer(ctx context.Context, purchaserID, redeemerID int64) (float64, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
SELECT COALESCE(SUM(amount), 0)::double precision
FROM package_redeem_sale_rebate_ledger_entries
WHERE purchaser_id = $1 AND redeemer_id = $2`, purchaserID, redeemerID)
	if err != nil {
		return 0, fmt.Errorf("query package redeem sale rebate per redeemer total: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var total float64
	if rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, err
		}
	}
	return total, rows.Close()
}

func nullableInt64Value(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func packageRedeemSaleRebateSourceID(redeemCodeID int64) string {
	return "redeem_code:" + strconv.FormatInt(redeemCodeID, 10)
}
