package subscription_test

import (
	"context"
	"testing"

	"weatherApi/internal/subscription"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Моки ---

type mockRepo struct{ mock.Mock }

func (m *mockRepo) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	args := m.Called(ctx, email)
	if s := args.Get(0); s != nil {
		return s.(*subscription.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, sub *subscription.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *mockRepo) Update(ctx context.Context, sub *subscription.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *mockRepo) GetConfirmedByFrequency(ctx context.Context, frequency string) ([]subscription.Subscription, error) {
	args := m.Called(ctx, frequency)
	return args.Get(0).([]subscription.Subscription), args.Error(1)
}

type mockTokenService struct{ mock.Mock }

func (m *mockTokenService) Generate(email string) (string, error) {
	args := m.Called(email)
	return args.String(0), args.Error(1)
}

func (m *mockTokenService) Parse(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

type mockEmailService struct{ mock.Mock }

func (m *mockEmailService) SendConfirmationEmail(email, token string) error {
	return m.Called(email, token).Error(0)
}

type mockCityValidator struct{ mock.Mock }

func (m *mockCityValidator) CityIsValid(ctx context.Context, city string) (bool, error) {
	args := m.Called(ctx, city)
	return args.Bool(0), args.Error(1)
}

// --- Фабрика мок-сервісу ---

type testDeps struct {
	repo      *mockRepo
	tokens    *mockTokenService
	emails    *mockEmailService
	validator *mockCityValidator
	service   *subscription.SubscriptionService
}

func createTestService() *testDeps {
	repo := new(mockRepo)
	tokens := new(mockTokenService)
	emails := new(mockEmailService)
	validator := new(mockCityValidator)
	service := subscription.NewSubscriptionService(repo, emails, validator, tokens)

	return &testDeps{repo, tokens, emails, validator, service}
}

// --- Тести ---

func TestConfirm_ValidToken_Success(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "valid-token"
	sub := &subscription.Subscription{Email: email, IsConfirmed: false}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)
	d.repo.On("Update", ctx, sub).Return(nil)

	err := d.service.Confirm(ctx, token)

	assert.NoError(t, err)
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestUnsubscribe_ValidToken_Success(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "valid-token"
	sub := &subscription.Subscription{Email: email, IsUnsubscribed: false}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)
	d.repo.On("Update", ctx, sub).Return(nil)

	err := d.service.Unsubscribe(ctx, token)

	assert.NoError(t, err)
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestListConfirmedByFrequency(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	frequency := "weekly"
	subs := []subscription.Subscription{
		{Email: "a@example.com", IsConfirmed: true},
		{Email: "b@example.com", IsConfirmed: true},
	}

	d.repo.On("GetConfirmedByFrequency", ctx, frequency).Return(subs, nil)

	result, err := d.service.ListConfirmedByFrequency(ctx, frequency)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	d.repo.AssertExpectations(t)
}

func TestConfirm_SubscriptionNotFound(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "notfound@example.com"
	token := "token-404"

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(nil, subscription.ErrSubscriptionNotFound)

	err := d.service.Confirm(ctx, token)

	assert.ErrorIs(t, err, subscription.ErrSubscriptionNotFound)
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestConfirm_InvalidToken(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	token := "invalid-token"

	d.tokens.On("Parse", token).Return("", subscription.ErrInvalidToken)

	err := d.service.Confirm(ctx, token)

	assert.ErrorIs(t, err, subscription.ErrInvalidToken)
	d.tokens.AssertExpectations(t)
}

func TestConfirm_AlreadyConfirmed(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "valid-token"
	sub := &subscription.Subscription{Email: email, IsConfirmed: true}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)

	err := d.service.Confirm(ctx, token)

	assert.NoError(t, err)
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestUnsubscribe_AlreadyUnsubscribed(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "token"
	sub := &subscription.Subscription{Email: email, IsUnsubscribed: true}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)

	err := d.service.Unsubscribe(ctx, token)

	assert.NoError(t, err)
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestConfirm_UpdateFails(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "token"
	sub := &subscription.Subscription{Email: email, IsConfirmed: false}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)
	d.repo.On("Update", ctx, sub).Return(assert.AnError)

	err := d.service.Confirm(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to confirm subscription")
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestSubscribe_RenewsUnsubscribedUser(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	city := "Kyiv"
	frequency := "daily"
	token := "new-token"

	existing := &subscription.Subscription{Email: email, IsConfirmed: true, IsUnsubscribed: true}

	d.validator.On("CityIsValid", ctx, city).Return(true, nil)
	d.repo.On("GetByEmail", ctx, email).Return(existing, nil)
	d.tokens.On("Generate", email).Return(token, nil)
	d.repo.On("Update", ctx, mock.AnythingOfType("*subscription.Subscription")).Return(nil)
	d.emails.On("SendConfirmationEmail", email, token).Maybe().Return(nil)

	err := d.service.Subscribe(ctx, email, city, frequency)

	assert.NoError(t, err)
	d.validator.AssertExpectations(t)
	d.repo.AssertExpectations(t)
	d.tokens.AssertExpectations(t)
	d.emails.AssertExpectations(t)
}

func TestSubscribe_CityValidatorFails_Error(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	city := "InvalidCity"

	d.validator.On("CityIsValid", ctx, city).Return(false, subscription.ErrCityNotFound)

	err := d.service.Subscribe(ctx, email, city, "daily")

	assert.ErrorIs(t, err, subscription.ErrCityNotFound)
	d.validator.AssertExpectations(t)
}
