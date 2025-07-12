package service_test

import (
	"context"
	"errors"
	"testing"

	"subscription/internal/domain"

	"subscription/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCKS ---

type mockRepo struct{ mock.Mock }

func (m *mockRepo) GetByEmail(ctx context.Context, email string) (*domain.Subscription, error) {
	args := m.Called(ctx, email)
	if s := args.Get(0); s != nil {
		return s.(*domain.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *mockRepo) Update(ctx context.Context, sub *domain.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *mockRepo) GetConfirmedByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error) {
	args := m.Called(ctx, frequency)
	return args.Get(0).([]domain.Subscription), args.Error(1)
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

func (m *mockEmailService) SendWeatherReport(email string, weatherReport domain.Report, city, token string) error {
	return m.Called(email, weatherReport, city, token).Error(0)
}

type mockCityValidator struct{ mock.Mock }

func (m *mockCityValidator) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(domain.Report), args.Error(1)
}

func (m *mockCityValidator) CityIsValid(ctx context.Context, city string) (bool, error) {
	args := m.Called(ctx, city)
	return args.Bool(0), args.Error(1)
}

type testDeps struct {
	repo      *mockRepo
	tokens    *mockTokenService
	emails    *mockEmailService
	validator *mockCityValidator
	service   service.Service
}

func createTestService() *testDeps {
	repo := new(mockRepo)
	tokens := new(mockTokenService)
	emails := new(mockEmailService)
	validator := new(mockCityValidator)
	service := service.NewService(repo, emails, validator, tokens)

	return &testDeps{repo, tokens, emails, validator, service}
}

// --- SUBSCRIBE ---

func TestSubscribe_SendsConfirmationEmail_Success(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "test@example.com"
	city := "Kyiv"
	frequency := domain.FreqDaily
	token := "abc-token"

	d.validator.On("CityIsValid", ctx, city).Return(true, nil)
	d.repo.On("GetByEmail", ctx, email).Return(nil, service.ErrSubscriptionNotFound)
	d.tokens.On("Generate", email).Return(token, nil)
	d.repo.On("Create", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil)

	d.emails.On("SendConfirmationEmail", email, token).Return(nil).Once()

	err := d.service.Subscribe(ctx, email, city, frequency)

	assert.NoError(t, err)
	d.emails.AssertCalled(t, "SendConfirmationEmail", email, token)
	d.emails.AssertExpectations(t)
}

func TestSubscribe_EmailSendFails_ButSubscribeStillSuccess(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "fail@example.com"
	city := "Lviv"
	frequency := domain.FreqDaily
	token := "fail-token"

	d.validator.On("CityIsValid", ctx, city).Return(true, nil)
	d.repo.On("GetByEmail", ctx, email).Return(nil, service.ErrSubscriptionNotFound)
	d.tokens.On("Generate", email).Return(token, nil)
	d.repo.On("Create", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil)

	d.emails.On("SendConfirmationEmail", email, token).Return(errors.New("smtp timeout")).Once()

	err := d.service.Subscribe(ctx, email, city, frequency)

	assert.NoError(t, err) // ми не ламаємо логіку підписки
	d.emails.AssertCalled(t, "SendConfirmationEmail", email, token)
}

func TestSubscribe_RenewsUnsubscribedUser(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	city := "Kyiv"
	frequency := domain.FreqDaily
	token := "new-token"

	existing := &domain.Subscription{Email: email, IsConfirmed: true, IsUnsubscribed: true}

	d.validator.On("CityIsValid", ctx, city).Return(true, nil)
	d.repo.On("GetByEmail", ctx, email).Return(existing, nil)
	d.tokens.On("Generate", email).Return(token, nil)
	d.repo.On("Update", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil)
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

	d.validator.On("CityIsValid", ctx, city).Return(false, service.ErrCityNotFound)

	err := d.service.Subscribe(ctx, email, city, "daily")

	assert.ErrorIs(t, err, service.ErrCityNotFound)
	d.validator.AssertExpectations(t)
}

func TestSubscribe_CityValidatorUnexpectedError(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	city := "ValidCity"

	validatorErr := errors.New("validator service down")
	d.validator.On("CityIsValid", ctx, city).Return(false, validatorErr)

	err := d.service.Subscribe(ctx, email, city, "daily")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate city")
	assert.ErrorIs(t, err, validatorErr)

	d.validator.AssertExpectations(t)
}

func TestSubscribe_AlreadySubscribed_ReturnsErr(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	city := "Kyiv"

	existing := &domain.Subscription{
		Email:          email,
		IsConfirmed:    true,
		IsUnsubscribed: false,
	}

	d.validator.On("CityIsValid", ctx, city).Return(true, nil)
	d.repo.On("GetByEmail", ctx, email).Return(existing, nil)

	err := d.service.Subscribe(ctx, email, city, "daily")

	assert.ErrorIs(t, err, service.ErrEmailAlreadyExists)
}

func TestSubscribe_GetByEmailUnexpectedError_ReturnsErr(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	city := "Kyiv"

	d.validator.On("CityIsValid", ctx, city).Return(true, nil)
	d.repo.On("GetByEmail", ctx, email).Return(nil, assert.AnError)

	err := d.service.Subscribe(ctx, email, city, "daily")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check existing subscription")
}

func TestSubscribe_TokenGenerationFails_ReturnsErr(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	city := "Kyiv"

	existing := &domain.Subscription{
		Email:          email,
		IsConfirmed:    true,
		IsUnsubscribed: true,
	}

	d.validator.On("CityIsValid", ctx, city).Return(true, nil)
	d.repo.On("GetByEmail", ctx, email).Return(existing, nil)
	d.tokens.On("Generate", email).Return("", assert.AnError)

	err := d.service.Subscribe(ctx, email, city, "daily")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not generate token")
}

func TestSubscribe_UpdateFails_ReturnsErr(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	city := "Kyiv"
	token := "token123"

	existing := &domain.Subscription{
		Email:          email,
		IsConfirmed:    true,
		IsUnsubscribed: true,
	}

	d.validator.On("CityIsValid", ctx, city).Return(true, nil)
	d.repo.On("GetByEmail", ctx, email).Return(existing, nil)
	d.tokens.On("Generate", email).Return(token, nil)
	d.repo.On("Update", ctx, mock.AnythingOfType("*domain.Subscription")).Return(assert.AnError)

	err := d.service.Subscribe(ctx, email, city, "daily")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update subscription")
}

func TestSubscribe_CreateFails_ReturnsErr(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "new@example.com"
	city := "Kyiv"
	token := "token456"

	d.validator.On("CityIsValid", ctx, city).Return(true, nil)
	d.repo.On("GetByEmail", ctx, email).Return(nil, service.ErrSubscriptionNotFound)
	d.tokens.On("Generate", email).Return(token, nil)
	d.repo.On("Create", ctx, mock.AnythingOfType("*domain.Subscription")).Return(assert.AnError)

	err := d.service.Subscribe(ctx, email, city, "daily")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create subscription")
}

// --- CONFIRM ---

func TestConfirm_ValidToken_Success(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "valid-token"
	sub := &domain.Subscription{Email: email, IsConfirmed: false}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)
	d.repo.On("Update", ctx, sub).Return(nil)

	err := d.service.Confirm(ctx, token)

	assert.NoError(t, err)
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestListConfirmedByFrequency(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	frequency := "weekly"
	subs := []domain.Subscription{
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
	d.repo.On("GetByEmail", ctx, email).Return(nil, service.ErrSubscriptionNotFound)

	err := d.service.Confirm(ctx, token)

	assert.ErrorIs(t, err, service.ErrSubscriptionNotFound)
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestConfirm_UnexpectedGetByEmailError(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	token := "token123"
	email := "user@example.com"
	fakeErr := errors.New("db timeout")

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(nil, fakeErr)

	err := d.service.Confirm(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get subscription")
	assert.ErrorIs(t, err, fakeErr)

	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestConfirm_InvalidToken(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	token := "invalid-token"

	d.tokens.On("Parse", token).Return("", service.ErrInvalidToken)

	err := d.service.Confirm(ctx, token)

	assert.ErrorIs(t, err, service.ErrInvalidToken)
	d.tokens.AssertExpectations(t)
}

func TestConfirm_AlreadyConfirmed(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "valid-token"
	sub := &domain.Subscription{Email: email, IsConfirmed: true}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)

	err := d.service.Confirm(ctx, token)

	assert.NoError(t, err)
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestConfirm_UpdateFails(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "token"
	sub := &domain.Subscription{Email: email, IsConfirmed: false}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)
	d.repo.On("Update", ctx, sub).Return(assert.AnError)

	err := d.service.Confirm(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to confirm subscription")
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

// --- UNSUBSCRIBE ---

func TestUnsubscribe_ValidToken_Success(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "valid-token"
	sub := &domain.Subscription{Email: email, IsUnsubscribed: false}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)
	d.repo.On("Update", ctx, sub).Return(nil)

	err := d.service.Unsubscribe(ctx, token)

	assert.NoError(t, err)
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestUnsubscribe_AlreadyUnsubscribed(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "token"
	sub := &domain.Subscription{Email: email, IsUnsubscribed: true}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)

	err := d.service.Unsubscribe(ctx, token)

	assert.NoError(t, err)
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestUnsubscribe_InvalidToken_ReturnsErr(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	token := "bad-token"

	d.tokens.On("Parse", token).Return("", service.ErrInvalidToken)

	err := d.service.Unsubscribe(ctx, token)

	assert.ErrorIs(t, err, service.ErrInvalidToken)
	d.tokens.AssertExpectations(t)
}

func TestUnsubscribe_GetByEmailFails_ReturnsErr(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "valid-token"

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(nil, assert.AnError)

	err := d.service.Unsubscribe(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get subscription")
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

func TestUnsubscribe_UpdateFails_ReturnsErr(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	email := "user@example.com"
	token := "valid-token"
	sub := &domain.Subscription{Email: email, IsUnsubscribed: false}

	d.tokens.On("Parse", token).Return(email, nil)
	d.repo.On("GetByEmail", ctx, email).Return(sub, nil)
	d.repo.On("Update", ctx, sub).Return(assert.AnError)

	err := d.service.Unsubscribe(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unsubscribe")
	d.tokens.AssertExpectations(t)
	d.repo.AssertExpectations(t)
}

// ---GENERATE_WEATHER_REPORT_TASKS ---

func TestGenerateWeatherReportTasks_Success(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	frequency := "daily"

	subs := []domain.Subscription{
		{Email: "a@example.com", City: "Kyiv", Token: "token1"},
		{Email: "b@example.com", City: "Lviv", Token: "token2"},
	}

	d.repo.On("GetConfirmedByFrequency", ctx, frequency).Return(subs, nil)

	tasks, err := d.service.GenerateWeatherReportTasks(ctx, frequency)

	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, "a@example.com", tasks[0].Email)
	assert.Equal(t, "Kyiv", tasks[0].City)
	assert.Equal(t, "token1", tasks[0].Token)
}

func TestGenerateWeatherReportTasks_ListFails_ReturnsError(t *testing.T) {
	d := createTestService()
	ctx := context.Background()
	frequency := "daily"

	d.repo.On("GetConfirmedByFrequency", ctx, frequency).
		Return([]domain.Subscription(nil), assert.AnError)

	tasks, err := d.service.GenerateWeatherReportTasks(ctx, frequency)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	d.repo.AssertExpectations(t)
}
