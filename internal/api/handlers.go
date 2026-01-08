package api

import (
	"net/http"
	"strconv"
	"time"

	"greenlabelai/backend/internal/models"
	"greenlabelai/backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB            *gorm.DB
	ImpactService *services.ImpactService
	BasketService *services.BasketService
}

func NewHandler(db *gorm.DB) *Handler {
	impactService := services.NewImpactService(db)
	impactService.SeedBadges() // Seed badges on startup

	openFoodService := &services.OpenFoodFactsService{} // Assuming no state for now or create properly if needed
	basketService := services.NewBasketService(db, openFoodService)

	return &Handler{
		DB:            db,
		ImpactService: impactService,
		BasketService: basketService,
	}
}

func (h *Handler) ScanProduct(c *gin.Context) {
	var req struct {
		Barcode string `json:"barcode" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	// 1. Try to find product in DB
	var product models.Product
	result := h.DB.Where("barcode = ?", req.Barcode).Limit(1).Find(&product)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Database error"})
		return
	}

	if result.RowsAffected == 0 {
		// 2. If not in DB, fetch from OFF
		fetchedProduct, err := services.FetchProductFromOFF(req.Barcode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
			return
		}
		product = *fetchedProduct
		h.DB.Create(&product)
	}

	// 3. Save to search history
	history := models.SearchHistory{
		Barcode:     product.Barcode,
		ProductName: product.Name,
		SearchedAt:  time.Now(),
	}
	h.DB.Create(&history)

	c.JSON(http.StatusOK, product) // Home.tsx expects the product object directly when scanning
}

func (h *Handler) GetSearchHistory(c *gin.Context) {
	var history []models.SearchHistory
	h.DB.Order("searched_at desc").Limit(20).Find(&history)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"history": history,
	})
}

func (h *Handler) GetProducts(c *gin.Context) {
	var products []models.Product
	h.DB.Order("created_at desc").Find(&products)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"products": products,
	})
}

func (h *Handler) ClearSearchHistory(c *gin.Context) {
	h.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.SearchHistory{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "History cleared"})
}

func (h *Handler) GetProduct(c *gin.Context) {
	barcode := c.Param("barcode")
	var product models.Product
	result := h.DB.Where("barcode = ?", barcode).Limit(1).Find(&product)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Database error"})
		return
	}

	if result.RowsAffected == 0 {
		fetchedProduct, err := services.FetchProductFromOFF(barcode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Product not found"})
			return
		}
		product = *fetchedProduct
		h.DB.Create(&product)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"product": product,
	})
}

func (h *Handler) GetMacros(c *gin.Context) {
	barcode := c.Param("barcode")
	macros, err := services.FetchMacrosFromOFF(barcode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"macros":  macros,
	})
}

func (h *Handler) GetRecipes(c *gin.Context) {
	barcode := c.Param("barcode")
	count := 2
	if v := c.Query("count"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			count = parsed
		}
	}

	var product models.Product
	result := h.DB.Where("barcode = ?", barcode).Limit(1).Find(&product)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Database error"})
		return
	}

	if result.RowsAffected == 0 {
		fetchedProduct, err := services.FetchProductFromOFF(barcode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Product not found"})
			return
		}
		product = *fetchedProduct
		h.DB.Create(&product)
	}

	productName := product.Name
	if productName == "" {
		productName = product.ProductName
	}
	if productName == "" {
		productName = barcode
	}

	recipes, err := services.GenerateRecipesForProduct(productName, count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"recipes": recipes,
	})
}

func (h *Handler) GetRecommendations(c *gin.Context) {
	barcode := c.Param("barcode")

	var currentProduct models.Product
	currentResult := h.DB.Where("barcode = ?", barcode).Limit(1).Find(&currentProduct)
	if currentResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Database error"})
		return
	}

	if currentResult.RowsAffected == 0 {
		fetchedProduct, err := services.FetchProductFromOFF(barcode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Product not found"})
			return
		}
		currentProduct = *fetchedProduct
		h.DB.Create(&currentProduct)
	}

	currentScore := currentProduct.GreenScore
	if currentScore == 0 {
		currentScore = currentProduct.EcoScore
	}

	var betterProducts []models.Product
	betterResult := h.DB.
		Where("green_score > ? AND barcode <> ?", currentScore, barcode).
		Order("green_score desc").
		Limit(3).
		Find(&betterProducts)
	if betterResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Database error"})
		return
	}

	offAlternatives := make([]gin.H, 0)
	if len(betterProducts) < 3 {
		need := 3 - len(betterProducts)
		alternatives, err := services.FetchBetterAlternativesFromOFF(barcode, currentScore, need)
		if err == nil {
			for _, alt := range alternatives {
				category := ""
				if alt.RawData != "" {
					category = "OpenFoodFacts"
				}
				offAlternatives = append(offAlternatives, gin.H{
					"name":                  alt.Name,
					"brand":                 alt.Brand,
					"category":              category,
					"why_better":            "Higher Eco-Score (from OpenFoodFacts)",
					"estimated_green_score": alt.EcoScore,
					"key_benefits":          []string{"Higher Eco-Score"},
					"where_to_find":         "Check local stores / online",
					"price_comparison":      "Unknown",
					"certifications":        "",
				})
			}
		}
	}

	recommendations := gin.H{
		"database_products": betterProducts,
		"ai_suggestions":    offAlternatives,
		"current_score":     currentScore,
		"improvement_tips": []string{
			"Choose products with minimal or recyclable packaging",
			"Prefer locally produced items to reduce transport emissions",
			"Look for credible eco-certifications",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"recommendations": recommendations,
	})
}

func (h *Handler) GetImpactStats(c *gin.Context) {
	stats, err := h.ImpactService.GetImpactStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to calculate stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "stats": stats})
}

func (h *Handler) CreateGoal(c *gin.Context) {
	var goal models.UserGoal
	if err := c.ShouldBindJSON(&goal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}
	if err := h.ImpactService.CreateGoal(&goal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create goal"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "goal": goal})
}

func (h *Handler) GetBadges(c *gin.Context) {
	badges, err := h.ImpactService.GetBadges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch badges"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "badges": badges})
}

func (h *Handler) AnalyzeBasket(c *gin.Context) {
	var req struct {
		Barcodes []string `json:"barcodes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}
	basket, err := h.BasketService.AnalyzeBasket(req.Barcodes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to analyze basket"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "basket": basket})
}
