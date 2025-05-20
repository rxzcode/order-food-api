package main

import (
	"path/filepath"

	"github.com/gin-gonic/gin"

	"order-food-api/core/config"
	"order-food-api/core/database"
	"order-food-api/handlers"
	"order-food-api/middleware"
	"order-food-api/models"
)

func main() {
	absPath, err := filepath.Abs("./config.ini")
	if err != nil {
		panic("Failed to get absolute path to config.ini: " + err.Error())
	}

	cfg := config.LoadConfig(absPath)
	db := database.Connect(cfg.Database)
	db.AutoMigrate(&models.Product{}, &models.Order{}, &models.OrderItem{})

	r := gin.Default()
	api := r.Group("/api")
	{
		handle := handlers.NewHandler(handlers.WithDB(db))
		api.GET("/product", handle.ListProducts())
		api.GET("/product/:productId", handle.GetProduct())
		api.POST("/product", middleware.APIKeyAuth(), handle.CreateProduct())
		api.POST("/order", middleware.APIKeyAuth(), handle.PlaceOrder())
	}

	r.Run(":8080")
}
