package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/emadhejazian/subscription_service/internal/delivery/http/middleware"
	domainusecase "github.com/emadhejazian/subscription_service/internal/domain/usecase"
)

type VoucherHandler struct {
	voucherUsecase domainusecase.VoucherUsecase
	planUsecase    domainusecase.PlanUsecase
}

func NewVoucherHandler(voucherUC domainusecase.VoucherUsecase, planUC domainusecase.PlanUsecase) *VoucherHandler {
	return &VoucherHandler{
		voucherUsecase: voucherUC,
		planUsecase:    planUC,
	}
}

type validateVoucherRequest struct {
	Code      string `json:"code"       binding:"required" example:"SAVE10"`
	ProductID uint   `json:"product_id" binding:"required" example:"1"`
	PlanID    uint   `json:"plan_id"    binding:"required" example:"2"`
}

type validateVoucherResponse struct {
	VoucherID       uint    `json:"voucher_id"`
	Code            string  `json:"code"`
	DiscountType    string  `json:"discount_type"`
	DiscountValue   float64 `json:"discount_value"`
	OriginalPrice   float64 `json:"original_price"`
	DiscountedPrice float64 `json:"discounted_price"`
}

// Validate godoc
//
//	@Summary      Validate a voucher
//	@Description  Checks whether a voucher code is valid for a given product and returns the discounted price for the selected plan.
//	@Tags         vouchers
//	@Accept       json
//	@Produce      json
//	@Param        body  body      validateVoucherRequest  true  "Validate request"
//	@Success      200   {object}  map[string]interface{}  "data: validateVoucherResponse"
//	@Failure      400   {object}  map[string]interface{}  "bad request"
//	@Failure      404   {object}  map[string]interface{}  "plan not found"
//	@Failure      422   {object}  map[string]interface{}  "invalid or expired voucher"
//	@Router       /vouchers/validate [post]
func (h *VoucherHandler) Validate(c *gin.Context) {
	var req validateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	plan, err := h.planUsecase.GetByID(req.PlanID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusNotFound, "plan not found")
		return
	}

	voucher, err := h.voucherUsecase.Validate(req.Code, req.ProductID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	originalPrice := plan.Price
	discountedPrice := h.voucherUsecase.Apply(voucher, originalPrice)

	resp := validateVoucherResponse{
		VoucherID:       voucher.ID,
		Code:            voucher.Code,
		DiscountType:    string(voucher.Type),
		DiscountValue:   voucher.Value,
		OriginalPrice:   originalPrice,
		DiscountedPrice: discountedPrice,
	}
	middleware.SuccessResponse(c, http.StatusOK, resp)
}
