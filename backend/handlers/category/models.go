package category

type CategoriesResponse struct {
	PrimaryCategories  map[string]string `json:"primary_categories"`
	DetailedCategories map[string]string `json:"detailed_categories"`
}
