package app

import (
	"log"
	"net/http"

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

type ProviderDeps struct {
	cfg         *config.Config
	logger      *log.Logger
	redisClient *redis.Client
	httpClient  *http.Client
	metrics     *cache.Metrics
}

func BuildProviders(deps ProviderDeps) ProviderSet {
	emailProvider := email.NewSendgrid(deps.cfg.SendGridKey, deps.cfg.EmailFrom)
	tokenProvider := token.NewJWT(deps.cfg.JWTSecret)

	weatherChainProvider := buildWeatherProvider(deps)

	return ProviderSet{
		EmailProvider:        emailProvider,
		TokenProvider:        tokenProvider,
		WeatherChainProvider: weatherChainProvider,
	}
}

func buildWeatherProvider(deps ProviderDeps) weather.Provider {
	baseWeatherAPI := weatherapi.New(deps.cfg.WeatherAPIKey, deps.httpClient)
	baseTomorrowIO := tomorrowio.New(deps.cfg.TomorrowioAPIKey, deps.httpClient)

	var wrappedWeatherAPI, wrappedTomorrowIO weather.Provider = baseWeatherAPI, baseTomorrowIO
	var redisCache cache.RedisCache

	if deps.redisClient != nil && deps.cfg.Cache.Enabled {
		redisCache = cache.NewRedisCache(deps.redisClient)

		wrappedWeatherAPI = cache.NewWriter(
			baseWeatherAPI,
			redisCache,
			"WeatherAPI",
			deps.cfg.Cache.WeatherApiTTL,
		)

		wrappedTomorrowIO = cache.NewWriter(
			baseTomorrowIO,
			redisCache,
			"TomorrowIO",
			deps.cfg.Cache.TomorrowIoTTL,
		)
	}

	loggedWeatherAPI := weather.NewLogWrapper(wrappedWeatherAPI, "WeatherAPI", deps.logger)
	loggedTomorrowIO := weather.NewLogWrapper(wrappedTomorrowIO, "TomorrowIO", deps.logger)

	nodeWeatherAPI := weather.NewChainNode(loggedWeatherAPI)
	nodeTomorrowIO := weather.NewChainNode(loggedTomorrowIO)
	nodeWeatherAPI.SetNext(nodeTomorrowIO)

	if deps.redisClient != nil && deps.cfg.Cache.Enabled {
		return cache.NewReader(
			nodeWeatherAPI,
			redisCache,
			deps.metrics,
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
