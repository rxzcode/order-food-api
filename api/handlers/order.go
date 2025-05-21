package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"order-food-api/core"
	"order-food-api/models"
	"order-food-api/models/dto"
)

const (
	ErrOrderInvalidInput       = "Invalid input"
	ErrOrderInvalidProductID   = "Invalid product ID"
	ErrOrderFailedCreateOrder  = "Failed to create order"
	ErrOrderFailedFetchProduct = "Failed to fetch products"
)

func (h *Handler) PlaceOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.OrderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			core.RespondError(c, http.StatusBadRequest, ErrOrderInvalidInput, err)
			return
		}

		// Verify coupon by cache
		if !h.Info.CouponCache.AppearsInAtLeastN(req.CouponCode, 2) {
			core.RespondError(c, http.StatusBadRequest, ErrOrderInvalidInput, nil)
			return
		}

		orderID := uuid.NewString()
		order := models.Order{
			ID:         orderID,
			CouponCode: req.CouponCode,
		}

		var productIDs []int

		for _, item := range req.Items {
			productIDInt, err := strconv.Atoi(item.ProductID)
			if err != nil {
				core.RespondError(c, http.StatusBadRequest, ErrOrderInvalidProductID, err)
				return
			}
			order.Items = append(order.Items, models.OrderItem{
				ProductID: models.ProductID(productIDInt),
				Quantity:  item.Quantity,
			})
			productIDs = append(productIDs, productIDInt)
		}

		if err := h.DB.Create(&order).Error; err != nil {
			core.RespondError(c, http.StatusInternalServerError, ErrOrderFailedCreateOrder, err)
			return
		}

		var products []models.Product
		if err := h.DB.Where("id IN ?", productIDs).Find(&products).Error; err != nil {
			core.RespondError(c, http.StatusInternalServerError, ErrOrderFailedFetchProduct, err)
			return
		}

		order.Products = products
		core.RespondSuccess(c, order)
	}
}
