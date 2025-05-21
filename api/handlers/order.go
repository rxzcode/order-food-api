// handler.go
package handlers

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

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

		// METHOD 01: Verify coupon code by read file
		// files := []string{
		// 	filepath.Join(h.Info.BasePath, "./files/couponbase1"),
		// 	filepath.Join(h.Info.BasePath, "./files/couponbase2"),
		// 	filepath.Join(h.Info.BasePath, "./files/couponbase3"),
		// }
		// if !checkCouponCode(req.CouponCode, files) {
		// 	core.RespondError(c, http.StatusBadRequest, ErrOrderInvalidInput, nil)
		// 	return
		// }

		// METHOD 02: Verify coupon by cache
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

// Function that check code by reading files - it take 6s per request (I want a faster way)
func checkCouponCode(code string, files []string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan bool, len(files))
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go containsCode(ctx, file, code, resultChan, &wg)
	}

	matches := 0
	for i := 0; i < len(files); i++ {
		if <-resultChan {
			matches++
			if matches >= 2 {
				cancel()
				break
			}
		}
	}

	wg.Wait()
	return matches >= 2
}

func containsCode(ctx context.Context, filePath string, checkCode string, resultChan chan<- bool, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filePath, err)
		resultChan <- false
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return // Early exit
		default:
			if scanner.Text() == checkCode {
				resultChan <- true
				return
			}
		}
	}
	resultChan <- false
}
