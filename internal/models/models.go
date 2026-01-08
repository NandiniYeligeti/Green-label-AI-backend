package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Barcode         string         `gorm:"uniqueIndex" json:"barcode"`
	Name            string         `json:"name"`
	ProductName     string         `json:"productName"` // For frontend compatibility in Home.tsx
	Brand           string         `json:"brand"`
	ImageURL        string         `json:"image_url"`
	Image           string         `json:"image"` // For frontend compatibility in Home.tsx
	GreenScore      int            `json:"green_score"`
	EcoScore        int            `json:"ecoScore"` // For frontend compatibility in Home.tsx
	NutritionGrade  string         `json:"nutrition_grade"`
	EcoScoreGrade   string         `json:"ecoscore_grade"`
	PackagingInfo   string         `json:"packaging_info"`
	Packaging       string         `json:"packaging"` // For frontend compatibility in Home.tsx
	IngredientsText string         `json:"ingredients_text"`
	Source          string         `json:"source"`
	RawData         string         `json:"raw_data"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type SearchHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Barcode     string    `json:"barcode"`
	ProductName string    `json:"product_name"`
	SearchedAt  time.Time `json:"searched_at"`
}

type UserGoal struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Type        string    `json:"type"` // e.g., "carbon_reduction", "healthy_swaps"
	Description string    `json:"description"`
	TargetValue int       `json:"target_value"` // e.g., 20 (%)
	Progress    int       `json:"progress"`     // Current progress
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	IsCompleted bool      `json:"is_completed"`
}

type Badge struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"` // Name of the icon to display
	Criteria    string `json:"criteria"`
}

type UserBadge struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	BadgeID  uint      `json:"badge_id"`
	Badge    Badge     `json:"badge"`
	EarnedAt time.Time `json:"earned_at"`
}

type ShoppingBasket struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time    `json:"created_at"`
	TotalItems     int          `json:"total_items"`
	TotalCarbon    float64      `json:"total_carbon"`
	AvgHealthScore float64      `json:"avg_health_score"`
	Items          []BasketItem `json:"items" gorm:"foreignKey:BasketID"`
}

type BasketItem struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	BasketID    uint    `json:"basket_id"`
	Barcode     string  `json:"barcode"`
	ProductName string  `json:"product_name"`
	Carbon      float64 `json:"carbon"`
	HealthScore int     `json:"health_score"`
}
