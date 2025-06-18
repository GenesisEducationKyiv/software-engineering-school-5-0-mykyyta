//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestSubscribeViaUI(t *testing.T) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/weatherdb?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.Close()
		require.NoError(t, err, "failed to close database connection")
	})

	var browser *rod.Browser
	remoteURL := os.Getenv("ROD_REMOTE_DEBUG_URL")
	if remoteURL != "" {
		t.Logf("Connecting to remote browser: %s", remoteURL)
		browser = rod.New().ControlURL(remoteURL).MustConnect()
	} else {
		launchURL := launcher.New().
			Headless(true).
			Set("no-sandbox").
			MustLaunch()
		browser = rod.New().ControlURL(launchURL).MustConnect()
	}
	t.Cleanup(func() {
		_ = browser.Close()
	})

	page := browser.MustPage("http://localhost:8080/subscribe")

	page.MustElement("input[name=email]").MustInput("ui-test@example.com")
	page.MustElement("input[name=city]").MustInput("Lviv")
	page.MustElement("select[name=frequency]").MustSelect("daily")
	page.MustElement("button[type=submit]").MustClick()

	time.Sleep(2 * time.Second)

	var count int
	err = db.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM subscriptions
		WHERE email = 'ui-test@example.com' AND city = 'Lviv' AND frequency = 'daily' AND is_confirmed = false
	`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "Expected one unconfirmed subscription to be created")
}
