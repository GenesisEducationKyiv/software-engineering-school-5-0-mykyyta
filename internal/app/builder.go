package app

import (
	"log"

	"weatherApi/internal/weather/cache"

	"github.com/redis/go-redis/v9"

	"weatherApi/internal/config"
	"weatherApi/internal/email"
	"weatherApi/internal/subscription"
	"weatherApi/internal/token"
	"weatherApi/internal/weather"
	"weatherApi/internal/weather/providers/tomorrowio"
	"weatherApi/internal/weather/providers/weatherapi"
)

type ProviderSet struct {
	EmailProvider        email.Provider
	TokenProvider        token.Provider
	WeatherChainProvider weather.Provider
}
type ServiceSet struct {
	SubService     subscription.Service
	WeatherService weather.Service
	EmailService   email.Service
}

func BuildProviders(cfg *config.Config, logger *log.Logger, redisClient *redis.Client) ProviderSet {
	emailProvider := email.NewSendgrid(cfg.SendGridKey, cfg.EmailFrom)
	tokenProvider := token.NewJWT(cfg.JWTSecret)

	weatherChainProvider := buildWeatherProvider(cfg, logger, redisClient)

	return ProviderSet{
		EmailProvider:        emailProvider,
		TokenProvider:        tokenProvider,
		WeatherChainProvider: weatherChainProvider,
	}
}

func buildWeatherProvider(cfg *config.Config, logger *log.Logger, redisClient *redis.Client) weather.Provider {
	baseWeatherAPI := weatherapi.New(cfg.WeatherAPIKey)
	baseTomorrowIO := tomorrowio.New(cfg.TomorrowioAPIKey)

	var wrappedWeatherAPI, wrappedTomorrowIO weather.Provider = baseWeatherAPI, baseTomorrowIO
	var redisCache *cache.RedisCache

	if redisClient != nil && cfg.Cache.Enabled {
		redisCache = cache.NewRedisCache(redisClient)

		wrappedWeatherAPI = cache.NewWriter(
			baseWeatherAPI,
			redisCache,
			"WeatherAPI",
			cfg.Cache.WeatherApiTTL,
		)

		wrappedTomorrowIO = cache.NewWriter(
			baseTomorrowIO,
			redisCache,
			"TomorrowIO",
			cfg.Cache.TomorrowIoTTL,
		)
	}

	loggedWeatherAPI := weather.NewLogWrapper(wrappedWeatherAPI, "WeatherAPI", logger)
	loggedTomorrowIO := weather.NewLogWrapper(wrappedTomorrowIO, "TomorrowIO", logger)

	nodeWeatherAPI := weather.NewChainNode(loggedWeatherAPI)
	nodeTomorrowIO := weather.NewChainNode(loggedTomorrowIO)
	nodeWeatherAPI.SetNext(nodeTomorrowIO)

	if redisClient != nil && cfg.Cache.Enabled {
		return cache.NewReader(
			nodeWeatherAPI,
			redisCache,
			[]string{"WeatherAPI", "TomorrowIO"},
		)
	}

	return nodeWeatherAPI
}

func BuildServices(db *DB, cfg *config.Config, p ProviderSet) ServiceSet {
	emailService := email.NewService(p.EmailProvider, cfg.BaseURL)
	tokenService := token.NewService(p.TokenProvider)
	weatherService := weather.NewService(p.WeatherChainProvider)

	repo := subscription.NewRepo(db.Gorm)
	subService := subscription.NewService(repo, emailService, weatherService, tokenService)

	return ServiceSet{
		SubService:     subService,
		WeatherService: weatherService,
		EmailService:   emailService,
	}
}
