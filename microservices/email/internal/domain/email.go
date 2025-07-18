package domain

type TemplateName string

const (
	TemplateConfirmation  TemplateName = "confirmation"
	TemplateWeatherReport TemplateName = "weather_report"
)

type SendEmailRequest struct {
	IdKey    string
	To       string
	Template TemplateName
	Data     map[string]string
}
