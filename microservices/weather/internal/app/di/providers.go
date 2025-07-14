package di

import (
	"log"
	"net/http"

	cache2 "weather/internal/adapter/cache"
	"weather/internal/adapter/chain"
	"weather/internal/adapter/logger"
	"weather/internal/adapter/provider/tomorrowio"
	"weather/internal/adapter/provider/weatherapi"

	"github.com/redis/go-redis/v9"

	"weather/internal/config"
	"weather/internal/service"
)

type ProviderDeps struct {
	Cfg         *config.Config
	Logger      *log.Logger
	RedisClient *redis.Client
	HttpClient  *http.Client
	Metrics     *cache2.Metrics
}

func BuildProviders(deps ProviderDeps) service.Provider {
	baseWeatherAPI := weatherapi.New(deps.Cfg.WeatherAPIKey, deps.HttpClient)
	baseTomorrowIO := tomorrowio.New(deps.Cfg.TomorrowioAPIKey, deps.HttpClient)

	var wrappedWeatherAPI, wrappedTomorrowIO service.Provider = baseWeatherAPI, baseTomorrowIO
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
