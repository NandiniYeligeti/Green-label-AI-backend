package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"greenlabelai/backend/internal/models"
)

type OFFResponse struct {
	Status  int `json:"status"`
	Product struct {
		ProductName string   `json:"product_name"`
		Brands      string   `json:"brands"`
		ImageURL    string   `json:"image_url"`
		Nutriscore  string   `json:"nutriscore_grade"`
		EcoscoreNum int      `json:"ecoscore_score"`
		Ecoscore    string   `json:"ecoscore_grade"`
		Packaging   string   `json:"packaging_text"`
		Ingredients string   `json:"ingredients_text"`
		Categories  string   `json:"categories"`
		CatTags     []string `json:"categories_tags"`
	} `json:"product"`
}

type OFFProductCategoryResponse struct {
	Product struct {
		CategoriesTags []string `json:"categories_tags"`
		Categories     string   `json:"categories"`
	} `json:"product"`
}

type OFFSearchResponse struct {
	Products []struct {
		Code          string   `json:"code"`
		ProductName   string   `json:"product_name"`
		Brands        string   `json:"brands"`
		Categories    string   `json:"categories"`
		CategoriesTag []string `json:"categories_tags"`
		ImageURL      string   `json:"image_url"`
		Nutriscore    string   `json:"nutriscore_grade"`
		EcoscoreNum   int      `json:"ecoscore_score"`
		Ecoscore      string   `json:"ecoscore_grade"`
	} `json:"products"`
}

type OFFMacrosResponse struct {
	Status  int `json:"status"`
	Product struct {
		Nutriments map[string]interface{} `json:"nutriments"`
	} `json:"product"`
}

type Macros struct {
	CaloriesKcal float64 `json:"calories_kcal"`
	ProteinG     float64 `json:"protein_g"`
	CarbsG       float64 `json:"carbs_g"`
	FatG         float64 `json:"fat_g"`
	Per          string  `json:"per"`
}

type OpenFoodFactsService struct{}

func (s *OpenFoodFactsService) GetProductByBarcode(barcode string) (*models.Product, error) {
	return FetchProductFromOFF(barcode)
}

func FetchProductFromOFF(barcode string) (*models.Product, error) {
	url := fmt.Sprintf("https://world.openfoodfacts.org/api/v2/product/%s.json", barcode)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OFF API returned status: %d", resp.StatusCode)
	}

	var offResp OFFResponse
	if err := json.NewDecoder(resp.Body).Decode(&offResp); err != nil {
		return nil, err
	}

	if offResp.Status == 0 {
		return nil, fmt.Errorf("product not found")
	}

	greenScore := offResp.Product.EcoscoreNum
	if greenScore <= 0 {
		greenScore = rand.Intn(50) + 50
	}

	product := &models.Product{
		Barcode:         barcode,
		Name:            offResp.Product.ProductName,
		ProductName:     offResp.Product.ProductName,
		Brand:           offResp.Product.Brands,
		ImageURL:        offResp.Product.ImageURL,
		Image:           offResp.Product.ImageURL,
		GreenScore:      greenScore,
		EcoScore:        greenScore,
		NutritionGrade:  offResp.Product.Nutriscore,
		EcoScoreGrade:   offResp.Product.Ecoscore,
		PackagingInfo:   offResp.Product.Packaging,
		Packaging:       offResp.Product.Packaging,
		IngredientsText: offResp.Product.Ingredients,
		Source:          "OpenFoodFacts",
	}

	// For RawData, we'll just store a simplified version for now
	rawData, _ := json.Marshal(offResp.Product)
	product.RawData = string(rawData)

	return product, nil
}

func FetchBetterAlternativesFromOFF(currentBarcode string, minEcoScore int, limit int) ([]models.Product, error) {
	if limit <= 0 {
		return []models.Product{}, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}

	catURL := fmt.Sprintf("https://world.openfoodfacts.org/api/v2/product/%s.json?fields=categories_tags,categories", currentBarcode)
	catResp, err := client.Get(catURL)
	if err != nil {
		return nil, err
	}
	defer catResp.Body.Close()

	if catResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OFF API returned status: %d", catResp.StatusCode)
	}

	var catData OFFProductCategoryResponse
	if err := json.NewDecoder(catResp.Body).Decode(&catData); err != nil {
		return nil, err
	}

	categoryTags := catData.Product.CategoriesTags
	if len(categoryTags) == 0 {
		return []models.Product{}, nil
	}

	categoryTag := ""
	for i := len(categoryTags) - 1; i >= 0; i-- {
		if strings.HasPrefix(categoryTags[i], "en:") {
			categoryTag = categoryTags[i]
			break
		}
	}
	if categoryTag == "" {
		categoryTag = categoryTags[len(categoryTags)-1]
	}

	q := url.Values{}
	q.Set("categories_tags", categoryTag)
	q.Set("fields", "code,product_name,brands,categories,categories_tags,image_url,nutriscore_grade,ecoscore_grade,ecoscore_score")
	q.Set("page_size", "50")
	q.Set("sort_by", "ecoscore_score")

	searchURL := fmt.Sprintf("https://world.openfoodfacts.org/api/v2/search?%s", q.Encode())
	searchResp, err := client.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer searchResp.Body.Close()

	if searchResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OFF API returned status: %d", searchResp.StatusCode)
	}

	var searchData OFFSearchResponse
	if err := json.NewDecoder(searchResp.Body).Decode(&searchData); err != nil {
		return nil, err
	}

	filtered := make([]models.Product, 0, limit)
	for _, p := range searchData.Products {
		if p.Code == "" || p.Code == currentBarcode {
			continue
		}
		if p.EcoscoreNum <= minEcoScore {
			continue
		}
		if strings.TrimSpace(p.ProductName) == "" {
			continue
		}

		rawData, _ := json.Marshal(p)
		filtered = append(filtered, models.Product{
			Barcode:        p.Code,
			Name:           p.ProductName,
			ProductName:    p.ProductName,
			Brand:          p.Brands,
			ImageURL:       p.ImageURL,
			Image:          p.ImageURL,
			GreenScore:     p.EcoscoreNum,
			EcoScore:       p.EcoscoreNum,
			NutritionGrade: p.Nutriscore,
			EcoScoreGrade:  p.Ecoscore,
			Source:         "OpenFoodFacts",
			RawData:        string(rawData),
		})
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].EcoScore > filtered[j].EcoScore
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

func FetchMacrosFromOFF(barcode string) (*Macros, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	macrosURL := fmt.Sprintf("https://world.openfoodfacts.org/api/v2/product/%s.json?fields=nutriments", barcode)

	resp, err := client.Get(macrosURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OFF API returned status: %d", resp.StatusCode)
	}

	var data OFFMacrosResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Status == 0 {
		return nil, fmt.Errorf("product not found")
	}

	getFloat := func(key string) float64 {
		v, ok := data.Product.Nutriments[key]
		if !ok || v == nil {
			return 0
		}
		switch t := v.(type) {
		case float64:
			return t
		case int:
			return float64(t)
		case string:
			f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
			if err == nil {
				return f
			}
		}
		return 0
	}

	caloriesKcal := getFloat("energy-kcal_100g")
	if caloriesKcal <= 0 {
		energyKJ := getFloat("energy_100g")
		if energyKJ > 0 {
			caloriesKcal = energyKJ * 0.239005736
		}
	}

	macros := &Macros{
		CaloriesKcal: caloriesKcal,
		ProteinG:     getFloat("proteins_100g"),
		CarbsG:       getFloat("carbohydrates_100g"),
		FatG:         getFloat("fat_100g"),
		Per:          "100g",
	}

	if macros.CaloriesKcal == 0 && macros.ProteinG == 0 && macros.CarbsG == 0 && macros.FatG == 0 {
		return nil, fmt.Errorf("macros not available")
	}

	return macros, nil
}
