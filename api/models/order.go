package models

type Order struct {
	ID         string      `json:"id" gorm:"primaryKey"`
	CouponCode string      `json:"couponCode"`
	Items      []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	Products   []Product   `json:"products" gorm:"-"`
}

type OrderItem struct {
	ID        uint      `json:"-" gorm:"primaryKey"`
	OrderID   string    `json:"-" gorm:"index"`
	ProductID ProductID `json:"productId"`
	Quantity  int       `json:"quantity"`
}
