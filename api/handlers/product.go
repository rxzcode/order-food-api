package handlers

import (
	"net/http"
	"order-food-api/core"
	"order-food-api/models"

	"github.com/gin-gonic/gin"
)

const (
	ErrProductInvalidInput = "Invalid input"
	ErrProductNotFound     = "Product not found"
	ErrProductCreate       = "Failed to create product"
)

func (h *Handler) ListProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []models.Product
		h.DB.Find(&products)
		c.JSON(http.StatusOK, products)
	}
}

func (h *Handler) GetProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("productId")
		var product models.Product
		if err := h.DB.First(&product, "id = ?", id).Error; err != nil {
			core.RespondError(c, http.StatusNotFound, ErrProductNotFound, err)
			return
		}
		c.JSON(http.StatusOK, product)
	}
}

func (h *Handler) CreateProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var product models.Product
		if err := c.ShouldBindJSON(&product); err != nil {
			core.RespondError(c, http.StatusBadRequest, ErrProductInvalidInput, err)
			return
		}

		if err := h.DB.Create(&product).Error; err != nil {
			core.RespondError(c, http.StatusInternalServerError, ErrProductCreate, err)
			return
		}

		c.JSON(http.StatusCreated, product)
	}
}
