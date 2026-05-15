package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type PackageRedeemAffiliateHandler struct {
	redeemService *service.RedeemService
	rebateRepo    service.PackageRedeemSaleRebateRepository
	setting       *service.SettingService
}

func NewPackageRedeemAffiliateHandler(redeemService *service.RedeemService, rebateRepo service.PackageRedeemSaleRebateRepository, setting *service.SettingService) *PackageRedeemAffiliateHandler {
	return &PackageRedeemAffiliateHandler{redeemService: redeemService, rebateRepo: rebateRepo, setting: setting}
}

func (h *PackageRedeemAffiliateHandler) Summary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	summary, err := h.rebateRepo.GetUserSummary(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if h.setting != nil {
		summary.DefaultRebateRatePercent = h.setting.GetPackageRedeemSaleRebateDefaultRatePercent(c.Request.Context())
	}
	response.Success(c, summary)
}

func (h *PackageRedeemAffiliateHandler) ListCodes(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	status := strings.TrimSpace(c.Query("status"))
	codes, result, err := h.redeemService.ListPurchasedPackageCodes(c.Request.Context(), subject.UserID, pagination.PaginationParams{Page: page, PageSize: pageSize}, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.RedeemCode, 0, len(codes))
	for i := range codes {
		out = append(out, *dto.RedeemCodeFromService(&codes[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}
