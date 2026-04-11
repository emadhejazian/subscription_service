package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/emadhejazian/subscription_service/internal/delivery/http/middleware"
	domainusecase "github.com/emadhejazian/subscription_service/internal/domain/usecase"
)

type SubscriptionHandler struct {
	usecase domainusecase.SubscriptionUsecase
}

func NewSubscriptionHandler(uc domainusecase.SubscriptionUsecase) *SubscriptionHandler {
	return &SubscriptionHandler{usecase: uc}
}

type buyRequest struct {
	ProductID   uint    `json:"product_id"   binding:"required" example:"1"`
	VoucherCode *string `json:"voucher_code" example:"SAVE10"`
	WithTrial   bool    `json:"with_trial"   example:"false"`
}

// Buy godoc
//
//	@Summary      Purchase a subscription
//	@Description  Creates a new subscription for the authenticated user. Optionally applies a voucher and/or starts a trial period.
//	@Tags         subscriptions
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      buyRequest              true  "Buy request"
//	@Success      201   {object}  map[string]interface{}  "data: entity.Subscription"
//	@Failure      400   {object}  map[string]interface{}  "bad request"
//	@Failure      401   {object}  map[string]interface{}  "unauthorized"
//	@Failure      422   {object}  map[string]interface{}  "unprocessable — duplicate subscription or invalid voucher"
//	@Router       /subscriptions [post]
func (h *SubscriptionHandler) Buy(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req buyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := h.usecase.Buy(domainusecase.BuyRequest{
		UserID:      userID.(string),
		ProductID:   req.ProductID,
		VoucherCode: req.VoucherCode,
		WithTrial:   req.WithTrial,
	})
	if err != nil {
		middleware.ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	middleware.SuccessResponse(c, http.StatusCreated, sub)
}

// GetByID godoc
//
//	@Summary      Get a subscription by ID
//	@Description  Returns a subscription with its associated product
//	@Tags         subscriptions
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      int  true  "Subscription ID"
//	@Success      200  {object}  map[string]interface{}  "data: entity.Subscription"
//	@Failure      400  {object}  map[string]interface{}  "invalid id"
//	@Failure      401  {object}  map[string]interface{}  "unauthorized"
//	@Failure      404  {object}  map[string]interface{}  "not found"
//	@Router       /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid subscription id")
		return
	}

	sub, err := h.usecase.GetByID(id)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusNotFound, "subscription not found")
		return
	}
	middleware.SuccessResponse(c, http.StatusOK, sub)
}

// Pause godoc
//
//	@Summary      Pause a subscription
//	@Description  Pauses an active subscription. Records the pause timestamp.
//	@Tags         subscriptions
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      int  true  "Subscription ID"
//	@Success      200  {object}  map[string]interface{}  "data: entity.Subscription"
//	@Failure      400  {object}  map[string]interface{}  "invalid id"
//	@Failure      401  {object}  map[string]interface{}  "unauthorized"
//	@Failure      422  {object}  map[string]interface{}  "cannot pause in current status"
//	@Router       /subscriptions/{id}/pause [post]
func (h *SubscriptionHandler) Pause(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid subscription id")
		return
	}

	sub, err := h.usecase.Pause(id)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	middleware.SuccessResponse(c, http.StatusOK, sub)
}

// Unpause godoc
//
//	@Summary      Unpause a subscription
//	@Description  Resumes a paused subscription and extends the end date by the number of days paused.
//	@Tags         subscriptions
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      int  true  "Subscription ID"
//	@Success      200  {object}  map[string]interface{}  "data: entity.Subscription"
//	@Failure      400  {object}  map[string]interface{}  "invalid id"
//	@Failure      401  {object}  map[string]interface{}  "unauthorized"
//	@Failure      422  {object}  map[string]interface{}  "cannot unpause in current status"
//	@Router       /subscriptions/{id}/unpause [post]
func (h *SubscriptionHandler) Unpause(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid subscription id")
		return
	}

	sub, err := h.usecase.Unpause(id)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	middleware.SuccessResponse(c, http.StatusOK, sub)
}

// Cancel godoc
//
//	@Summary      Cancel a subscription
//	@Description  Cancels an active, paused, or trialing subscription.
//	@Tags         subscriptions
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      int  true  "Subscription ID"
//	@Success      200  {object}  map[string]interface{}  "data: entity.Subscription"
//	@Failure      400  {object}  map[string]interface{}  "invalid id"
//	@Failure      401  {object}  map[string]interface{}  "unauthorized"
//	@Failure      422  {object}  map[string]interface{}  "cannot cancel in current status"
//	@Router       /subscriptions/{id}/cancel [post]
func (h *SubscriptionHandler) Cancel(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid subscription id")
		return
	}

	sub, err := h.usecase.Cancel(id)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	middleware.SuccessResponse(c, http.StatusOK, sub)
}

// GetMySubscription godoc
//
//	@Summary      Get current user's active subscriptions
//	@Description  Returns all active, paused, or trialing subscriptions for the authenticated user.
//	@Tags         subscriptions
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200  {object}  map[string]interface{}  "data: []entity.Subscription"
//	@Failure      401  {object}  map[string]interface{}  "unauthorized"
//	@Failure      404  {object}  map[string]interface{}  "no active subscriptions found"
//	@Router       /subscriptions/me [get]
func (h *SubscriptionHandler) GetMySubscription(c *gin.Context) {
	userID := c.GetString("userID")

	subs, err := h.usecase.GetActiveByUserID(userID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	if len(subs) == 0 {
		middleware.ErrorResponse(c, http.StatusNotFound, "no active subscriptions found")
		return
	}
	middleware.SuccessResponse(c, http.StatusOK, subs)
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id), err
}
