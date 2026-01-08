package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(h *Handler) *gin.Engine {
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // In production, limit this to actual frontend URL
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	r.Use(cors.New(config))

	api := r.Group("/api")
	{
		api.POST("/scan-product", h.ScanProduct)
		api.GET("/search-history", h.GetSearchHistory)
		api.DELETE("/search-history", h.ClearSearchHistory)
		api.GET("/products", h.GetProducts)
		api.GET("/product/:barcode", h.GetProduct)
		api.GET("/product/:barcode/macros", h.GetMacros)
		api.GET("/product/:barcode/recipes", h.GetRecipes)
		api.GET("/product/:barcode/recommendations", h.GetRecommendations)
		api.GET("/impact/stats", h.GetImpactStats)
		api.POST("/goals", h.CreateGoal)
		api.POST("/basket", h.AnalyzeBasket)
		api.GET("/badges", h.GetBadges)
	}

	return r
}
