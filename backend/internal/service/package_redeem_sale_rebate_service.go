package service

import (
	"context"
	"math"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

const (
	PackageRedeemSaleRebateStatusPending           = "pending"
	PackageRedeemSaleRebateStatusApplied           = "applied"
	PackageRedeemSaleRebateStatusSkippedSelfRedeem = "skipped_self_redeem"
	PackageRedeemSaleRebateStatusSkippedZeroAmount = "skipped_zero_amount"
	PackageRedeemSaleRebateStatusSkippedDisabled   = "skipped_disabled"
	PackageRedeemSaleRebateStatusFailed            = "failed"
)

type PackageRedeemSaleRebate struct {
	ID              int64
	RedeemCodeID    int64
	PurchaseOrderID *int64
	PurchaserID     int64
	RedeemerID      int64
	BaseAmount      float64
	RebateRate      float64
	RebateAmount    float64
	Currency        string
	Status          string
	Reason          string
	Error           string
	AppliedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreditPackageRedeemSaleRebateInput struct {
	SaleRebateID    int64
	PurchaserID     int64
	RedeemerID      int64
	RedeemCodeID    int64
	PurchaseOrderID *int64
	BaseAmount      float64
	RebateRate      float64
	RebateAmount    float64
	Currency        string
	FreezeHours     int
}

type PackageRedeemSaleRebateRepository interface {
	CreatePendingIfNotExists(ctx context.Context, input PackageRedeemSaleRebate) (*PackageRedeemSaleRebate, error)
	CreateSkippedIfNotExists(ctx context.Context, input PackageRedeemSaleRebate, status, reason string) (*PackageRedeemSaleRebate, error)
	MarkApplied(ctx context.Context, id int64, rebateRate, rebateAmount float64, currency string) error
	MarkSkipped(ctx context.Context, id int64, status, reason string, rebateRate, rebateAmount float64, currency string) error
	MarkFailed(ctx context.Context, id int64, rebateRate, rebateAmount float64, currency string, err error) error
}

type PackageRedeemSaleRebateRecorder interface {
	RecordPending(ctx context.Context, redeemCode *RedeemCode, redeemerID int64) error
}

type PackageRedeemSaleLedgerRepository interface {
	CreditSaleRebate(ctx context.Context, input CreditPackageRedeemSaleRebateInput) (bool, error)
	GetCreditedAmountByRedeemer(ctx context.Context, purchaserID, redeemerID int64) (float64, error)
}

type PackageRedeemSaleRebateService struct {
	rebateRepo           PackageRedeemSaleRebateRepository
	ledgerRepo           PackageRedeemSaleLedgerRepository
	settingService       *SettingService
	affiliateService     *AffiliateService
	entClient            *dbent.Client
	authCacheInvalidator APIKeyAuthCacheInvalidator
	billingCacheService  *BillingCacheService
}

func NewPackageRedeemSaleRebateService(
	rebateRepo PackageRedeemSaleRebateRepository,
	ledgerRepo PackageRedeemSaleLedgerRepository,
	settingService *SettingService,
	affiliateService *AffiliateService,
	entClient *dbent.Client,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	billingCacheService *BillingCacheService,
) *PackageRedeemSaleRebateService {
	return &PackageRedeemSaleRebateService{
		rebateRepo:           rebateRepo,
		ledgerRepo:           ledgerRepo,
		settingService:       settingService,
		affiliateService:     affiliateService,
		entClient:            entClient,
		authCacheInvalidator: authCacheInvalidator,
		billingCacheService:  billingCacheService,
	}
}

func (s *PackageRedeemSaleRebateService) Process(ctx context.Context, redeemCode *RedeemCode, redeemerID int64) error {
	if redeemCode == nil || redeemCode.ID <= 0 || redeemCode.PurchasedBy == nil || *redeemCode.PurchasedBy <= 0 {
		return nil
	}
	if s == nil || s.rebateRepo == nil || s.ledgerRepo == nil {
		return infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "package redeem sale rebate service unavailable")
	}

	purchaserID := *redeemCode.PurchasedBy
	baseAmount := sanitizeMoney(redeemCode.PurchasePayAmount)
	currency := strings.TrimSpace(derefStringPtr(redeemCode.PurchaseCurrency))

	if purchaserID == redeemerID {
		_, err := s.rebateRepo.CreateSkippedIfNotExists(ctx, packageRedeemSaleRebateInput(redeemCode, purchaserID, redeemerID, baseAmount, 0, 0, currency), PackageRedeemSaleRebateStatusSkippedSelfRedeem, PackageRedeemSaleRebateStatusSkippedSelfRedeem)
		return err
	}

	rebate, err := s.rebateRepo.CreatePendingIfNotExists(ctx, packageRedeemSaleRebateInput(redeemCode, purchaserID, redeemerID, baseAmount, 0, 0, currency))
	if err != nil {
		return err
	}
	if rebate.Status == PackageRedeemSaleRebateStatusApplied || strings.HasPrefix(rebate.Status, "skipped_") {
		return nil
	}

	if s.settingService != nil && !s.settingService.IsPackageRedeemSaleRebateEnabled(ctx) {
		return s.rebateRepo.MarkSkipped(ctx, rebate.ID, PackageRedeemSaleRebateStatusSkippedDisabled, PackageRedeemSaleRebateStatusSkippedDisabled, 0, 0, currency)
	}

	if s.affiliateService != nil {
		if err := s.affiliateService.BindInviterIfAbsent(ctx, redeemerID, purchaserID); err != nil {
			logger.LegacyPrintf("service.package_redeem_sale_rebate", "[PackageRedeemSaleRebate] bind inviter failed redeemer=%d purchaser=%d: %v", redeemerID, purchaserID, err)
		}
	}

	rate := s.resolveSaleRebateRate(ctx)
	amount := roundTo(baseAmount*(rate/100), 8)
	amount, err = s.applyCaps(ctx, purchaserID, redeemerID, amount)
	if err != nil {
		_ = s.rebateRepo.MarkFailed(ctx, rebate.ID, rate, amount, currency, err)
		return err
	}
	if amount <= 0 {
		return s.rebateRepo.MarkSkipped(ctx, rebate.ID, PackageRedeemSaleRebateStatusSkippedZeroAmount, PackageRedeemSaleRebateStatusSkippedZeroAmount, rate, 0, currency)
	}

	freezeHours := 0
	if s.settingService != nil {
		freezeHours = s.settingService.GetPackageRedeemSaleRebateFreezeHours(ctx)
	}
	applied, err := s.ledgerRepo.CreditSaleRebate(ctx, CreditPackageRedeemSaleRebateInput{
		SaleRebateID:    rebate.ID,
		PurchaserID:     purchaserID,
		RedeemerID:      redeemerID,
		RedeemCodeID:    redeemCode.ID,
		PurchaseOrderID: redeemCode.PurchaseOrderID,
		BaseAmount:      baseAmount,
		RebateRate:      rate,
		RebateAmount:    amount,
		Currency:        currency,
		FreezeHours:     freezeHours,
	})
	if err != nil {
		_ = s.rebateRepo.MarkFailed(ctx, rebate.ID, rate, amount, currency, err)
		return err
	}
	if !applied {
		return nil
	}
	if err := s.rebateRepo.MarkApplied(ctx, rebate.ID, rate, amount, currency); err != nil {
		return err
	}
	s.invalidatePurchaserCaches(ctx, purchaserID)
	return nil
}

func (s *PackageRedeemSaleRebateService) RecordPending(ctx context.Context, redeemCode *RedeemCode, redeemerID int64) error {
	if redeemCode == nil || redeemCode.ID <= 0 || redeemCode.PurchasedBy == nil || *redeemCode.PurchasedBy <= 0 || s == nil || s.rebateRepo == nil {
		return nil
	}
	purchaserID := *redeemCode.PurchasedBy
	baseAmount := sanitizeMoney(redeemCode.PurchasePayAmount)
	currency := strings.TrimSpace(derefStringPtr(redeemCode.PurchaseCurrency))
	if purchaserID == redeemerID {
		_, err := s.rebateRepo.CreateSkippedIfNotExists(ctx, packageRedeemSaleRebateInput(redeemCode, purchaserID, redeemerID, baseAmount, 0, 0, currency), PackageRedeemSaleRebateStatusSkippedSelfRedeem, PackageRedeemSaleRebateStatusSkippedSelfRedeem)
		return err
	}
	_, err := s.rebateRepo.CreatePendingIfNotExists(ctx, packageRedeemSaleRebateInput(redeemCode, purchaserID, redeemerID, baseAmount, 0, 0, currency))
	return err
}

func (s *PackageRedeemSaleRebateService) resolveSaleRebateRate(ctx context.Context) float64 {
	if s == nil || s.settingService == nil {
		return PackageRedeemSaleRebateDefaultRatePercentDefault
	}
	return s.settingService.GetPackageRedeemSaleRebateDefaultRatePercent(ctx)
}

func (s *PackageRedeemSaleRebateService) applyCaps(ctx context.Context, purchaserID, redeemerID int64, amount float64) (float64, error) {
	if amount <= 0 || s == nil || s.settingService == nil {
		return amount, nil
	}
	if perOrderCap := s.settingService.GetPackageRedeemSaleRebatePerOrderCap(ctx); perOrderCap > 0 && amount > perOrderCap {
		amount = roundTo(perOrderCap, 8)
	}
	if perRedeemerCap := s.settingService.GetPackageRedeemSaleRebatePerRedeemerCap(ctx); perRedeemerCap > 0 {
		existing, err := s.ledgerRepo.GetCreditedAmountByRedeemer(ctx, purchaserID, redeemerID)
		if err != nil {
			return 0, err
		}
		if existing >= perRedeemerCap {
			return 0, nil
		}
		if remaining := perRedeemerCap - existing; amount > remaining {
			amount = roundTo(remaining, 8)
		}
	}
	return amount, nil
}

func (s *PackageRedeemSaleRebateService) invalidatePurchaserCaches(ctx context.Context, purchaserID int64) {
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, purchaserID)
	}
	if s.billingCacheService != nil {
		if err := s.billingCacheService.InvalidateUserBalance(ctx, purchaserID); err != nil {
			logger.LegacyPrintf("service.package_redeem_sale_rebate", "[PackageRedeemSaleRebate] invalidate billing cache failed purchaser=%d: %v", purchaserID, err)
		}
	}
}

func packageRedeemSaleRebateInput(redeemCode *RedeemCode, purchaserID, redeemerID int64, baseAmount, rate, amount float64, currency string) PackageRedeemSaleRebate {
	return PackageRedeemSaleRebate{
		RedeemCodeID:    redeemCode.ID,
		PurchaseOrderID: redeemCode.PurchaseOrderID,
		PurchaserID:     purchaserID,
		RedeemerID:      redeemerID,
		BaseAmount:      baseAmount,
		RebateRate:      rate,
		RebateAmount:    amount,
		Currency:        currency,
	}
}

func sanitizeMoney(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
		return 0
	}
	return v
}

func derefStringPtr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
