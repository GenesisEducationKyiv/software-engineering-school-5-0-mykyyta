//go:build integration

package subscription_test

import (
	"context"
	"testing"
	"time"
	"weatherApi/internal/domain"
	"weatherApi/internal/subscription/repo"
	"weatherApi/test/integration/testutils"

	"weatherApi/internal/subscription"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionRepository_CRUD(t *testing.T) {
	ctx := context.Background()

	pg, err := testutils.StartPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	}()

	repo := repo.NewRepo(pg.DB.Gorm)

	sub := &domain.Subscription{
		ID:             uuid.NewString(),
		Email:          "test@example.com",
		City:           "Kyiv",
		Frequency:      domain.FreqDaily,
		IsConfirmed:    false,
		IsUnsubscribed: false,
		Token:          "mock-token",
		CreatedAt:      time.Now(),
	}

	err = repo.Create(ctx, sub)
	require.NoError(t, err)

	got, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)

	require.Equal(t, sub.Email, got.Email)
	require.Equal(t, sub.City, got.City)
	require.Equal(t, sub.Frequency, got.Frequency)
	require.False(t, got.IsConfirmed)
	require.WithinDuration(t, sub.CreatedAt, got.CreatedAt, time.Minute)

	got.IsConfirmed = true
	err = repo.Update(ctx, got)
	require.NoError(t, err)

	updated, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	require.True(t, updated.IsConfirmed)
}

func TestSubscriptionRepository_NotFound(t *testing.T) {
	ctx := context.Background()

	pg, err := testutils.StartPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	}()

	repo := repo.NewRepo(pg.DB.Gorm)

	sub, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	require.ErrorIs(t, err, subscription.ErrSubscriptionNotFound)
	require.Nil(t, sub)
}
