package property

type Property struct {
	Name        string
	Description string
	Link        string
	Images      []string
	Bedrooms    int
	Bathrooms   int
	Features    []string
	Amenities   []string
	Price       float32
	Landarea    float32
}
