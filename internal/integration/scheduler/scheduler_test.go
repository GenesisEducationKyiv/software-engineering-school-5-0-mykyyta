package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"

	"weatherApi/internal/app"
	"weatherApi/internal/integration/testutils"
	"weatherApi/internal/scheduler"
	"weatherApi/internal/subscription"
)

func TestEmailDispatch_DailyConfirmed(t *testing.T) {
	ctx := context.Background()

	pg, err := testutils.StartPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	}()

	emailProvider := &testutils.FakeEmailProvider{}
	weatherProvider := &testutils.FakeWeatherProvider{Valid: true}
	tokenProvider := &testutils.FakeTokenProvider{}

	builder := &app.ServiceBuilder{
		DB:              pg.DB,
		BaseURL:         "http://localhost:8080",
		EmailProvider:   emailProvider,
		TokenProvider:   tokenProvider,
		WeatherProvider: weatherProvider,
	}
	services, err := builder.BuildServices()
	require.NoError(t, err)

	err = pg.DB.Gorm.Exec("DELETE FROM subscriptions").Error
	require.NoError(t, err)

	repo := subscription.NewSubscriptionRepository(pg.DB.Gorm)
	sub := &subscription.Subscription{
		Email:          "test@example.com",
		City:           "Kyiv",
		Token:          "some-token",
		Frequency:      "daily",
		IsConfirmed:    true,
		IsUnsubscribed: false,
	}
	err = repo.Create(ctx, sub)
	require.NoError(t, err)

	// Запускаємо Scheduler
	s := scheduler.NewScheduler(services.SubService, services.WeatherService, services.EmailService)
	defer s.Stop()
	s.Start()

	// Симулюємо cron запуск
	s.Dispatcher.DispatchScheduledEmails("daily")

	// Очікуємо, поки email буде надіслано
	time.Sleep(2 * time.Second)

	// Перевірка
	require.True(t, emailProvider.Sent, "Expected email to be sent")
	require.Equal(t, "test@example.com", emailProvider.To)
	require.Contains(t, emailProvider.Plain, "Kyiv")
	require.Contains(t, emailProvider.Plain, "Поточна погода в")
	require.Contains(t, emailProvider.Plain, "Відписатися")
}

func TestEmailDispatch_HourlyAndDailyConfirmed(t *testing.T) {
	ctx := context.Background()

	// Старт тестової PostgreSQL
	pg, err := testutils.StartPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	}()

	// Тестові залежності
	emailProvider := &testutils.FakeEmailProvider{}
	weatherProvider := &testutils.FakeWeatherProvider{Valid: true}
	tokenProvider := &testutils.FakeTokenProvider{}

	// Сервіси
	builder := &app.ServiceBuilder{
		DB:              pg.DB,
		BaseURL:         "http://localhost:8080",
		EmailProvider:   emailProvider,
		TokenProvider:   tokenProvider,
		WeatherProvider: weatherProvider,
	}
	services, err := builder.BuildServices()
	require.NoError(t, err)

	err = pg.DB.Gorm.Exec("DELETE FROM subscriptions").Error
	require.NoError(t, err)

	repo := subscription.NewSubscriptionRepository(pg.DB.Gorm)
	subs := []*subscription.Subscription{
		{
			ID:             uuid.NewString(),
			Email:          "daily@example.com",
			City:           "Kyiv",
			Token:          "daily-token",
			Frequency:      "daily",
			IsConfirmed:    true,
			IsUnsubscribed: false,
		},
		{
			ID:             uuid.NewString(),
			Email:          "hourly@example.com",
			City:           "Lviv",
			Token:          "hourly-token",
			Frequency:      "hourly",
			IsConfirmed:    true,
			IsUnsubscribed: false,
		},
	}
	for _, sub := range subs {
		err := repo.Create(ctx, sub)
		require.NoError(t, err)
	}

	s := scheduler.NewScheduler(services.SubService, services.WeatherService, services.EmailService)
	defer s.Stop()
	s.Start()

	s.Dispatcher.DispatchScheduledEmails("daily")
	time.Sleep(1 * time.Second)

	require.True(t, emailProvider.Sent, "Expected email to be sent (daily)")
	require.Equal(t, "daily@example.com", emailProvider.To)
	require.Contains(t, emailProvider.Plain, "Kyiv")
	require.Contains(t, emailProvider.Plain, "Поточна погода в")

	emailProvider.Sent = false
	emailProvider.To = ""
	emailProvider.Plain = ""

	s.Dispatcher.DispatchScheduledEmails("hourly")
	time.Sleep(1 * time.Second)

	require.True(t, emailProvider.Sent, "Expected email to be sent (hourly)")
	require.Equal(t, "hourly@example.com", emailProvider.To)
	require.Contains(t, emailProvider.Plain, "Lviv")
	require.Contains(t, emailProvider.Plain, "Поточна погода в")
}
