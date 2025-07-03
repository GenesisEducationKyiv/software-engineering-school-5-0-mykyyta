//go:build integration

package scheduler_test

import (
	"context"
	"testing"
	"time"

	"weatherApi/internal/config"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"

	"weatherApi/internal/app"
	"weatherApi/internal/integration/testutils"
	"weatherApi/internal/jobs"
	"weatherApi/internal/subscription"
)

type fakeEventSource struct {
	ch chan string
}

func (f *fakeEventSource) Events() <-chan string {
	return f.ch
}

func TestEmailDispatcher_DailyFrequency_SendsWeatherEmailToConfirmedUser(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pg, err := testutils.StartPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	}()

	emailProvider := &testutils.FakeEmailProvider{}

	providers := app.ProviderSet{
		EmailProvider:        emailProvider,
		TokenProvider:        &testutils.FakeTokenProvider{},
		WeatherChainProvider: &testutils.FakeWeatherProvider{Valid: true},
	}

	services := app.BuildServices(pg.DB, &config.Config{BaseURL: "http://localhost:8080"}, providers)

	err = pg.DB.Gorm.Exec("DELETE FROM subscriptions").Error
	require.NoError(t, err)

	repo := subscription.NewRepo(pg.DB.Gorm)
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

	eventChan := make(chan string, 1)
	fakeEventSource := &fakeEventSource{ch: eventChan}

	queue := jobs.NewLocalQueue(10)
	dispatcher := jobs.NewEmailDispatcher(services.SubService, queue, fakeEventSource)
	worker := jobs.NewWorker(queue, services.SubService)

	go worker.Start(ctx)
	dispatcher.Start(ctx)

	eventChan <- "daily"

	time.Sleep(2 * time.Second)

	require.True(t, emailProvider.Sent, "Expected email to be sent")
	require.Equal(t, "test@example.com", emailProvider.To)
	require.Contains(t, emailProvider.Plain, "Kyiv")
	require.Contains(t, emailProvider.Plain, "Поточна погода в")
	require.Contains(t, emailProvider.Plain, "Відписатися")
}

func TestEmailDispatcher_MultipleFrequencies_SendsToCorrectSubscribersOnly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pg, err := testutils.StartPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	}()

	emailProvider := &testutils.FakeEmailProvider{}

	providers := app.ProviderSet{
		EmailProvider:        emailProvider,
		TokenProvider:        &testutils.FakeTokenProvider{},
		WeatherChainProvider: &testutils.FakeWeatherProvider{Valid: true},
	}
	services := app.BuildServices(pg.DB, &config.Config{BaseURL: "http://localhost:8080"}, providers)

	err = pg.DB.Gorm.Exec("DELETE FROM subscriptions").Error
	require.NoError(t, err)

	repo := subscription.NewRepo(pg.DB.Gorm)
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

	eventChan := make(chan string, 2)
	fakeSource := &fakeEventSource{ch: eventChan}

	queue := jobs.NewLocalQueue(10)
	dispatcher := jobs.NewEmailDispatcher(services.SubService, queue, fakeSource)
	worker := jobs.NewWorker(queue, services.SubService)

	go dispatcher.Start(ctx)
	go worker.Start(ctx)

	eventChan <- "daily"
	time.Sleep(1 * time.Second)

	require.True(t, emailProvider.Sent, "Expected email to be sent (daily)")
	require.Equal(t, "daily@example.com", emailProvider.To)
	require.Contains(t, emailProvider.Plain, "Kyiv")
	require.Contains(t, emailProvider.Plain, "Поточна погода в")

	emailProvider.Sent = false
	emailProvider.To = ""
	emailProvider.Plain = ""

	eventChan <- "hourly"
	time.Sleep(1 * time.Second)

	require.True(t, emailProvider.Sent, "Expected email to be sent (hourly)")
	require.Equal(t, "hourly@example.com", emailProvider.To)
	require.Contains(t, emailProvider.Plain, "Lviv")
	require.Contains(t, emailProvider.Plain, "Поточна погода в")
}
