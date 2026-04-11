package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/emadhejazian/subscription_service/internal/delivery/http/middleware"
	domainusecase "github.com/emadhejazian/subscription_service/internal/domain/usecase"
)

type PlanHandler struct {
	usecase domainusecase.PlanUsecase
}

func NewPlanHandler(uc domainusecase.PlanUsecase) *PlanHandler {
	return &PlanHandler{usecase: uc}
}

// GetByProductID godoc
//
//	@Summary      List plans for a product
//	@Description  Returns all subscription plans available for a given sport course.
//	@Tags         plans
//	@Produce      json
//	@Param        id   path      int  true  "Product ID"
//	@Success      200  {object}  map[string]interface{}  "data: []entity.Plan"
//	@Failure      400  {object}  map[string]interface{}  "invalid product id"
//	@Failure      500  {object}  map[string]interface{}  "internal error"
//	@Router       /products/{id}/plans [get]
func (h *PlanHandler) GetByProductID(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
		return
	}

	plans, err := h.usecase.GetByProductID(uint(productID))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	middleware.SuccessResponse(c, http.StatusOK, plans)
}
