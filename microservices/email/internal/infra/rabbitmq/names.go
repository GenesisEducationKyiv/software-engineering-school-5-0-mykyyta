package rabbitmq

const (
	ExchangeEmail = "email.exchange"

	QueueEmail = "email.queue"

	RoutingKeyConfirmation = "cmd.email.send_confirmation"
	RoutingKeyWeather      = "cmd.email.send_weather_report"
)
