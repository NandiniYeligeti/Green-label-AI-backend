package services

import (
	"greenlabelai/backend/internal/models"

	"gorm.io/gorm"
)

type BasketService struct {
	db              *gorm.DB
	openFoodService *OpenFoodFactsService
}

func NewBasketService(db *gorm.DB, openFoodService *OpenFoodFactsService) *BasketService {
	return &BasketService{
		db:              db,
		openFoodService: openFoodService,
	}
}

func (s *BasketService) AnalyzeBasket(barcodes []string) (*models.ShoppingBasket, error) {
	basket := &models.ShoppingBasket{
		Items: []models.BasketItem{},
	}

	var totalScore float64
	var totalCarbon float64

	for _, code := range barcodes {
		// Try to find product locally first
		var product models.Product
		result := s.db.Where("barcode = ?", code).First(&product)

		if result.Error != nil {
			// Fetch from OpenFoodFacts if not found
			fetchedProduct, err := s.openFoodService.GetProductByBarcode(code)
			if err == nil {
				product = *fetchedProduct
			} else {
				// Skip if not found anywhere (or handle error)
				continue
			}
		}

		// Calculate simulated Carbon Footprint (inverse of EcoScore for demo)
		// 100 EcoScore = 0 Carbon, 0 EcoScore = 5 Carbon (arbitrary scale)
		carbon := 5.0 - (float64(product.EcoScore) / 20.0)
		if carbon < 0 {
			carbon = 0
		}

		item := models.BasketItem{
			Barcode:     code,
			ProductName: product.ProductName,
			Carbon:      carbon,
			HealthScore: product.GreenScore, // Using GreenScore as HealthScore for simplicity
		}

		basket.Items = append(basket.Items, item)
		totalCarbon += carbon
		totalScore += float64(product.GreenScore)
	}

	basket.TotalItems = len(basket.Items)
	basket.TotalCarbon = totalCarbon
	if basket.TotalItems > 0 {
		basket.AvgHealthScore = totalScore / float64(basket.TotalItems)
	}

	// Save basket record
	if err := s.db.Create(basket).Error; err != nil {
		return nil, err
	}

	return basket, nil
}
