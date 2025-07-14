package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	expectedEmail string
	returnToken   string
	returnEmail   string
	genCalled     bool
	parseCalled   bool
}

func (m *mockProvider) Generate(email string) (string, error) {
	m.genCalled = true
	m.expectedEmail = email
	return m.returnToken, nil
}

func (m *mockProvider) Parse(token string) (string, error) {
	m.parseCalled = true
	return m.returnEmail, nil
}

func TestTokenService_Generate(t *testing.T) {
	mock := &mockProvider{returnToken: "mocktoken"}
	svc := NewService(mock)

	token, err := svc.Generate("test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "mocktoken", token)
	assert.True(t, mock.genCalled)
	assert.Equal(t, "test@example.com", mock.expectedEmail)
}

func TestTokenService_Parse(t *testing.T) {
	mock := &mockProvider{returnEmail: "parsed@example.com"}
	svc := NewService(mock)

	email, err := svc.Parse("sometoken")
	require.NoError(t, err)
	assert.Equal(t, "parsed@example.com", email)
	assert.True(t, mock.parseCalled)
}
