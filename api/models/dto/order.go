package dto

type OrderReq struct {
	CouponCode string `json:"couponCode" binding:"omitempty,min=8,max=10"`
	Items      []struct {
		ProductID string `json:"productId" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
	} `json:"items" binding:"required"`
}
