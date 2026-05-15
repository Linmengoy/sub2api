package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type PackageRedeemAffiliateHandler struct {
	adminService service.AdminService
	rebateRepo   service.PackageRedeemSaleRebateRepository
}

func NewPackageRedeemAffiliateHandler(adminService service.AdminService, rebateRepo service.PackageRedeemSaleRebateRepository) *PackageRedeemAffiliateHandler {
	return &PackageRedeemAffiliateHandler{adminService: adminService, rebateRepo: rebateRepo}
}

func (h *PackageRedeemAffiliateHandler) Summary(c *gin.Context) {
	summary, err := h.rebateRepo.GetAdminSummary(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, summary)
}

func (h *PackageRedeemAffiliateHandler) ListCodes(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := strings.TrimSpace(c.Query("status"))
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}
	codes, total, err := h.adminService.ListRedeemCodes(c.Request.Context(), page, pageSize, service.RedeemTypeSubscription, status, search, c.DefaultQuery("sort_by", "id"), c.DefaultQuery("sort_order", "desc"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.AdminRedeemCode, 0, len(codes))
	for i := range codes {
		if codes[i].PurchasedBy == nil {
			continue
		}
		out = append(out, *dto.RedeemCodeFromServiceAdmin(&codes[i]))
	}
	response.Paginated(c, out, total, page, pageSize)
}

func (h *PackageRedeemAffiliateHandler) GetCode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid redeem code ID")
		return
	}
	code, err := h.adminService.GetRedeemCode(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if code.PurchasedBy == nil {
		response.BadRequest(c, "not a package redeem affiliate code")
		return
	}
	response.Success(c, dto.RedeemCodeFromServiceAdmin(code))
}

func (h *PackageRedeemAffiliateHandler) DeleteCode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid redeem code ID")
		return
	}
	code, err := h.adminService.GetRedeemCode(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if code.PurchasedBy == nil {
		response.BadRequest(c, "not a package redeem affiliate code")
		return
	}
	if err := h.adminService.DeleteRedeemCode(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "deleted"})
}
