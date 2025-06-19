package weather

type Report struct {
	Temperature float64 `json:"temperature"` // Temperature in degrees Celsius
	Humidity    int     `json:"humidity"`    // Relative humidity in percent (0–100)
	Description string  `json:"description"` // Short text description (e.g. "Clear", "Rainy")
}
