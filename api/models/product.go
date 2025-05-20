package models

import "strconv"

type ProductID int

func (id ProductID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.Itoa(int(id)) + `"`), nil
}

type Product struct {
	ID       ProductID `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string    `json:"name"`
	Price    float64   `json:"price"`
	Category string    `json:"category"`
	Image    Image     `gorm:"embedded" json:"image"`
}

type Image struct {
	Thumbnail string `json:"thumbnail"`
	Mobile    string `json:"mobile"`
	Tablet    string `json:"tablet"`
	Desktop   string `json:"desktop"`
}
