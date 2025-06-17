//go:build integration

package subscription_test

import (
	"context"
	"testing"
	"time"

	"weatherApi/internal/subscription"

	"weatherApi/internal/integration/testutils"

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

	repo := subscription.NewSubscriptionRepository(pg.DB.Gorm)

	sub := &subscription.Subscription{
		ID:             uuid.NewString(),
		Email:          "test@example.com",
		City:           "Kyiv",
		Frequency:      "daily",
		IsConfirmed:    false,
		IsUnsubscribed: false,
		Token:          "mock-token",
	}

	// Create
	err = repo.Create(ctx, sub)
	require.NoError(t, err)

	// Read
	got, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	require.Equal(t, sub.Email, got.Email)
	require.Equal(t, sub.City, got.City)
	require.False(t, got.IsConfirmed)
	require.WithinDuration(t, time.Now(), got.CreatedAt, time.Minute)

	// Update
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

	repo := subscription.NewSubscriptionRepository(pg.DB.Gorm)

	sub, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	require.ErrorIs(t, err, subscription.ErrSubscriptionNotFound)
	require.Nil(t, sub)
}
