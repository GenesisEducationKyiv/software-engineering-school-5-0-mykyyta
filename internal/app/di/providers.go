package di

import (
	"log"
	"net/http"

	"weatherApi/internal/weather/cache"
	"weatherApi/internal/weather/chain"
	"weatherApi/internal/weather/logger"

	"github.com/redis/go-redis/v9"

	"weatherApi/internal/config"
	"weatherApi/internal/email"
	"weatherApi/internal/token"
	"weatherApi/internal/weather"
	"weatherApi/internal/weather/provider/tomorrowio"
	"weatherApi/internal/weather/provider/weatherapi"
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
	Metrics     *cache.Metrics
}

func BuildProviders(deps ProviderDeps) Providers {
	emailProvider := email.NewSendgrid(deps.Cfg.SendGridKey, deps.Cfg.EmailFrom)
	tokenProvider := token.NewJWT(deps.Cfg.JWTSecret)

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
	var redisCache cache.RedisCache

	if deps.RedisClient != nil && deps.Cfg.Cache.Enabled {
		redisCache = cache.NewRedisCache(deps.RedisClient)

		wrappedWeatherAPI = cache.NewWriter(
			baseWeatherAPI,
			redisCache,
			"WeatherAPI",
			deps.Cfg.Cache.WeatherApiTTL,
			deps.Cfg.Cache.NotFoundTTL,
		)

		wrappedTomorrowIO = cache.NewWriter(
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
		return cache.NewReader(
			nodeWeatherAPI,
			redisCache,
			deps.Metrics,
			[]string{"WeatherAPI", "TomorrowIO"},
		)
	}

	return nodeWeatherAPI
}
