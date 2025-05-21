package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	cache "order-food-api/core/cacheBloomFilter"
	// cache "order-food-api/core/cacheBitmap"
	// cache "order-food-api/core/cacheMPH"
	// cache "order-food-api/core/cacheMap"
	// cache "order-food-api/core/search"

	"order-food-api/core/config"
	"order-food-api/core/database"
	"order-food-api/handlers"
	"order-food-api/middleware"
	"order-food-api/models"
)

func main() {
	go showServerStats()

	absPath, err := filepath.Abs(".")
	if err != nil {
		panic("Failed to get absolute path of program: " + err.Error())
	}

	files := []string{"./files/couponbase1.gz", "./files/couponbase2.gz", "./files/couponbase3.gz"}
	couponCache := cache.New()
	go func() {
		couponCache.LoadFiles(files)
	}()

	cfg := config.LoadConfig(filepath.Join(absPath, "config.ini"))
	db := database.Connect(cfg.Database)
	db.AutoMigrate(&models.Product{}, &models.Order{}, &models.OrderItem{})

	r := gin.Default()
	api := r.Group("/api")
	{
		handle := handlers.NewHandler(handlers.WithDB(db), handlers.WithInfo(handlers.InfoOption{BasePath: absPath, CouponCache: couponCache}))
		api.GET("/product", handle.ListProducts())
		api.GET("/product/:productId", handle.GetProduct())
		api.POST("/product", middleware.APIKeyAuth(), handle.CreateProduct())
		api.POST("/order", middleware.APIKeyAuth(), handle.PlaceOrder())
	}

	r.Run(":" + cfg.App.Port)
}

func showServerStats() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		allocMB := float64(m.Alloc) / 1024 / 1024
		totalMB := float64(m.TotalAlloc) / 1024 / 1024
		sysMB := float64(m.Sys) / 1024 / 1024
		numGC := m.NumGC
		numGoroutine := runtime.NumGoroutine()
		numCPU := runtime.NumCPU()
		fmt.Printf("[STATS] Goroutines: %d | CPUs: %d | Alloc: %.2f MB | TotalAlloc: %.2f MB | Sys: %.2f MB | GCs: %d\n",
			numGoroutine, numCPU, allocMB, totalMB, sysMB, numGC)
		debug.FreeOSMemory()
	}
}
