package di

import (
	"log"
	"net/http"

	"monolith/internal/config"
	"monolith/internal/email"
	"monolith/internal/email/sendgrid"
	"monolith/internal/token"
	"monolith/internal/token/jwt"
	"monolith/internal/weather"
	cache2 "monolith/internal/weather/cache"
	"monolith/internal/weather/chain"
	"monolith/internal/weather/logger"
	"monolith/internal/weather/provider/tomorrowio"
	"monolith/internal/weather/provider/weatherapi"

	"github.com/redis/go-redis/v9"
)

type Providers struct {
	EmailProvider        email.Provider
	TokenProvider        token.Provider
	WeatherChainProvider weather.Provider
}

type ProviderDeps struct {
	Cfg         *config.Config
	Logger      *log.Logger
	RedisClient *redis.Client
	HttpClient  *http.Client
	Metrics     *cache2.Metrics
}

func BuildProviders(deps ProviderDeps) Providers {
	emailProvider := sendgrid.NewSendgrid(deps.Cfg.SendGridKey, deps.Cfg.EmailFrom)
	tokenProvider := jwt.NewJWT(deps.Cfg.JWTSecret)

	weatherChainProvider := buildWeatherProvider(deps)

	return Providers{
		EmailProvider:        emailProvider,
		TokenProvider:        tokenProvider,
		WeatherChainProvider: weatherChainProvider,
	}
}

func buildWeatherProvider(deps ProviderDeps) weather.Provider {
	baseWeatherAPI := weatherapi.New(deps.Cfg.WeatherAPIKey, deps.HttpClient)
	baseTomorrowIO := tomorrowio.New(deps.Cfg.TomorrowioAPIKey, deps.HttpClient)

	var wrappedWeatherAPI, wrappedTomorrowIO weather.Provider = baseWeatherAPI, baseTomorrowIO
	var redisCache cache2.RedisCache

	if deps.RedisClient != nil && deps.Cfg.Cache.Enabled {
		redisCache = cache2.NewRedisCache(deps.RedisClient)

		wrappedWeatherAPI = cache2.NewWriter(
			baseWeatherAPI,
			redisCache,
			"WeatherAPI",
			deps.Cfg.Cache.WeatherApiTTL,
			deps.Cfg.Cache.NotFoundTTL,
		)

		wrappedTomorrowIO = cache2.NewWriter(
			baseTomorrowIO,
			redisCache,
			"TomorrowIO",
			deps.Cfg.Cache.TomorrowIoTTL,
			deps.Cfg.Cache.NotFoundTTL,
		)
	}

	loggedWeatherAPI := logger.NewWrapper(wrappedWeatherAPI, "WeatherAPI", deps.Logger)
	loggedTomorrowIO := logger.NewWrapper(wrappedTomorrowIO, "TomorrowIO", deps.Logger)

	nodeWeatherAPI := chain.NewNode(loggedWeatherAPI)
	nodeTomorrowIO := chain.NewNode(loggedTomorrowIO)
	nodeWeatherAPI.SetNext(nodeTomorrowIO)

	if deps.RedisClient != nil && deps.Cfg.Cache.Enabled {
		return cache2.NewReader(
			nodeWeatherAPI,
			redisCache,
			deps.Metrics,
			[]string{"WeatherAPI", "TomorrowIO"},
		)
	}

	return nodeWeatherAPI
}
