package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/emadhejazian/subscription_service/internal/delivery/http/middleware"
	domainusecase "github.com/emadhejazian/subscription_service/internal/domain/usecase"
)

type ProductHandler struct {
	usecase domainusecase.ProductUsecase
}

func NewProductHandler(uc domainusecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{usecase: uc}
}

// GetAll godoc
//
//	@Summary      List all products
//	@Description  Returns all available subscription products
//	@Tags         products
//	@Produce      json
//	@Success      200  {object}  map[string]interface{}  "data: []entity.Product"
//	@Failure      500  {object}  map[string]interface{}  "error message"
//	@Router       /products [get]
func (h *ProductHandler) GetAll(c *gin.Context) {
	products, err := h.usecase.GetAll()
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	middleware.SuccessResponse(c, http.StatusOK, products)
}

// GetByID godoc
//
//	@Summary      Get a product by ID
//	@Description  Returns a single product
//	@Tags         products
//	@Produce      json
//	@Param        id   path      int  true  "Product ID"
//	@Success      200  {object}  map[string]interface{}  "data: entity.Product"
//	@Failure      400  {object}  map[string]interface{}  "invalid product id"
//	@Failure      404  {object}  map[string]interface{}  "product not found"
//	@Router       /products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.usecase.GetByID(uint(id))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusNotFound, "product not found")
		return
	}
	middleware.SuccessResponse(c, http.StatusOK, product)
}
