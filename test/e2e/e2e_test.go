//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestSubscribeViaUI(t *testing.T) {
	// Підключення до бази даних
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/weatherdb?sslmode=disable")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.Close()
		require.NoError(t, err, "failed to close database connection")
	})

	// Запускаємо headless браузер
	url := launcher.New().Headless(true).MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	t.Cleanup(func() {
		_ = browser.Close()
	})

	page := browser.MustPage("http://localhost:8080/subscribe")

	// Взаємодія зі сторінкою
	page.MustElement("input[name=email]").MustInput("ui-test@example.com")
	page.MustElement("input[name=city]").MustInput("Lviv")
	page.MustElement("select[name=frequency]").MustSelect("daily")
	page.MustElement("button[type=submit]").MustClick()

	// Дати час бекенду записати в базу
	time.Sleep(2 * time.Second)

	// Перевірка: чи створено підписку
	var count int
	err = db.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM subscriptions
		WHERE email = 'ui-test@example.com' AND city = 'Lviv' AND frequency = 'daily' AND is_confirmed = false
	`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "Expected one unconfirmed subscription to be created")
}
