package services

import (
	"time"

	"greenlabelai/backend/internal/models"

	"gorm.io/gorm"
)

type ImpactService struct {
	db *gorm.DB
}

func NewImpactService(db *gorm.DB) *ImpactService {
	return &ImpactService{db: db}
}

type ImpactStats struct {
	TotalCarbonSaved float64           `json:"total_carbon_saved"`
	WeeklyReport     string            `json:"weekly_report"`
	ActiveGoals      []models.UserGoal `json:"active_goals"`
}

func (s *ImpactService) GetImpactStats() (*ImpactStats, error) {
	// Mock calculation for now, in a real app this would aggregate scan history
	var totalSaved float64 = 125.5 // kg CO2
	report := "You've saved 15% more carbon this week properly choosing eco-friendly products!"

	var goals []models.UserGoal
	// Fetch active goals
	result := s.db.Where("is_completed = ?", false).Find(&goals)
	if result.Error != nil {
		return nil, result.Error
	}

	return &ImpactStats{
		TotalCarbonSaved: totalSaved,
		WeeklyReport:     report,
		ActiveGoals:      goals,
	}, nil
}

func (s *ImpactService) CreateGoal(goal *models.UserGoal) error {
	goal.StartDate = time.Now()
	goal.IsCompleted = false
	return s.db.Create(goal).Error
}

func (s *ImpactService) GetBadges() ([]models.UserBadge, error) {
	var userBadges []models.UserBadge
	// Preload Badge details
	result := s.db.Preload("Badge").Find(&userBadges)
	return userBadges, result.Error
}

// SeedBadges creates some default badges if they don't exist
func (s *ImpactService) SeedBadges() {
	badges := []models.Badge{
		{Name: "Eco Starter", Description: "Scanned your first eco-friendly product", Icon: "Leaf", Criteria: "First Scan"},
		{Name: "Carbon Cutter", Description: "Saved 10kg of CO2", Icon: "Wind", Criteria: "Save 10kg CO2"},
		{Name: "Healthy Choice", Description: "Chose 5 products with Grade A/B", Icon: "Heart", Criteria: "5 Healthy Scans"},
	}

	for _, b := range badges {
		var count int64
		s.db.Model(&models.Badge{}).Where("name = ?", b.Name).Count(&count)
		if count == 0 {
			s.db.Create(&b)
		}
	}
}
