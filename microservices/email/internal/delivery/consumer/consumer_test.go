package consumer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"email/internal/delivery/consumer"
	"email/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
)

type mockSource struct {
	mock.Mock
}

func (m *mockSource) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan amqp.Delivery), args.Error(1)
}

type mockIdempotency struct {
	mock.Mock
}

func (m *mockIdempotency) IsProcessed(ctx context.Context, messageID string) (bool, error) {
	args := m.Called(ctx, messageID)
	return args.Bool(0), args.Error(1)
}

func (m *mockIdempotency) MarkAsProcessing(ctx context.Context, messageID string) (bool, error) {
	args := m.Called(ctx, messageID)
	return args.Bool(0), args.Error(1)
}

func (m *mockIdempotency) MarkAsProcessed(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *mockIdempotency) ClearProcessing(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

type mockUseCase struct {
	mock.Mock
}

func (m *mockUseCase) Send(req domain.SendEmailRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Printf(format string, v ...interface{}) {
	m.Called(format, v)
}

func (m *mockLogger) Println(v ...interface{}) {
	m.Called(v)
}

type MockAcknowledger struct {
	mock.Mock
}

func (m *MockAcknowledger) Ack(tag uint64, multiple bool) error {
	args := m.Called(tag, multiple)
	return args.Error(0)
}

func (m *MockAcknowledger) Nack(tag uint64, multiple, requeue bool) error {
	args := m.Called(tag, multiple, requeue)
	return args.Error(0)
}

func (m *MockAcknowledger) Reject(tag uint64, requeue bool) error {
	args := m.Called(tag, requeue)
	return args.Error(0)
}

type mockBreaker struct {
	mock.Mock
}

func (m *mockBreaker) CanExecute() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockBreaker) RecordSuccess() {
	m.Called()
}

func (m *mockBreaker) RecordFailure() {
	m.Called()
}

func TestConsumer_Handle_Success(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgCh := make(chan amqp.Delivery, 1)
	defer close(msgCh)

	req := domain.SendEmailRequest{
		To:       "test@example.com",
		Template: "confirmation",
	}
	body, _ := json.Marshal(req)

	ack := new(MockAcknowledger)
	ack.On("Ack", mock.Anything, false).Return(nil)

	msgCh <- amqp.Delivery{
		Body:         body,
		MessageId:    "msg-1",
		Acknowledger: ack,
		DeliveryTag:  1,
	}

	source := new(mockSource)
	source.On("Consume", mock.Anything).Return((<-chan amqp.Delivery)(msgCh), nil)

	idem := new(mockIdempotency)
	idem.On("IsProcessed", mock.Anything, "msg-1").Return(false, nil)
	idem.On("MarkAsProcessing", mock.Anything, "msg-1").Return(true, nil)
	idem.On("MarkAsProcessed", mock.Anything, "msg-1").Return(nil)
	idem.On("ClearProcessing", mock.Anything, "msg-1").Return(nil)

	useCase := new(mockUseCase)
	useCase.On("Send", req).Return(nil)

	logger := new(mockLogger)
	logger.On("Println", mock.Anything).Maybe()
	logger.On("Printf", mock.Anything, mock.Anything).Maybe()

	breaker := new(mockBreaker)
	breaker.On("CanExecute").Return(true)
	breaker.On("RecordSuccess").Return()
	breaker.On("RecordFailure").Return()

	c := consumer.NewConsumer(source, useCase, idem, logger, breaker)

	go func() {
		_ = c.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond) // wait for processing

	source.AssertExpectations(t)
	idem.AssertExpectations(t)
	useCase.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestConsumer_AlreadyProcessed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgCh := make(chan amqp.Delivery, 1)

	ack := new(MockAcknowledger)
	ack.On("Ack", uint64(1), false).Return(nil)

	msgCh <- amqp.Delivery{
		MessageId:    "msg-1",
		Acknowledger: ack,
		DeliveryTag:  1,
	}

	source := new(mockSource)
	source.On("Consume", mock.Anything).Return((<-chan amqp.Delivery)(msgCh), nil)

	idem := new(mockIdempotency)
	idem.On("IsProcessed", mock.Anything, "msg-1").Return(true, nil)

	useCase := new(mockUseCase)
	logger := new(mockLogger)
	logger.On("Println", mock.Anything).Maybe()
	logger.On("Printf", mock.Anything, mock.Anything).Maybe()

	breaker := new(mockBreaker)
	breaker.On("CanExecute").Return(true)
	breaker.On("RecordSuccess").Return()
	breaker.On("RecordFailure").Return()

	c := consumer.NewConsumer(source, useCase, idem, logger, breaker)
	go func() { _ = c.Start(ctx) }()
	time.Sleep(50 * time.Millisecond)

	idem.AssertExpectations(t)
	useCase.AssertNotCalled(t, "Send", mock.Anything)
}

func TestConsumer_InvalidJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ack := new(MockAcknowledger)
	ack.On("Nack", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	msg := amqp.Delivery{
		MessageId:    "msg-2",
		Body:         []byte("invalid json"),
		Acknowledger: ack,
		DeliveryTag:  2,
	}
	msgCh := make(chan amqp.Delivery, 1)
	msgCh <- msg

	source := new(mockSource)
	source.On("Consume", mock.Anything).Return((<-chan amqp.Delivery)(msgCh), nil)

	idem := new(mockIdempotency)
	idem.On("IsProcessed", mock.Anything, "msg-2").Return(false, nil)
	idem.On("MarkAsProcessing", mock.Anything, "msg-2").Return(true, nil)
	idem.On("ClearProcessing", mock.Anything, "msg-2").Return(nil)

	useCase := new(mockUseCase)
	logger := new(mockLogger)
	logger.On("Printf", mock.Anything, mock.Anything).Maybe()
	logger.On("Println", mock.Anything).Maybe()

	breaker := new(mockBreaker)
	breaker.On("CanExecute").Return(true)
	breaker.On("RecordSuccess").Return()
	breaker.On("RecordFailure").Return()

	c := consumer.NewConsumer(source, useCase, idem, logger, breaker)
	go func() { _ = c.Start(ctx) }()
	time.Sleep(50 * time.Millisecond)

	useCase.AssertNotCalled(t, "Send", mock.Anything)
}

func TestConsumer_SendFails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := domain.SendEmailRequest{To: "a@b.com", Template: "welcome"}
	body, _ := json.Marshal(req)

	ack := new(MockAcknowledger)
	ack.On("Nack", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	msg := amqp.Delivery{
		MessageId:    "msg-3",
		Body:         body,
		Acknowledger: ack,
		DeliveryTag:  3,
	}
	msgCh := make(chan amqp.Delivery, 1)
	msgCh <- msg

	source := new(mockSource)
	source.On("Consume", mock.Anything).Return((<-chan amqp.Delivery)(msgCh), nil)

	idem := new(mockIdempotency)
	idem.On("IsProcessed", mock.Anything, "msg-3").Return(false, nil)
	idem.On("MarkAsProcessing", mock.Anything, "msg-3").Return(true, nil)
	idem.On("ClearProcessing", mock.Anything, "msg-3").Return(nil)

	useCase := new(mockUseCase)
	useCase.On("Send", req).Return(fmt.Errorf("failed"))

	logger := new(mockLogger)
	logger.On("Printf", mock.Anything, mock.Anything).Maybe()
	logger.On("Println", mock.Anything).Maybe()

	breaker := new(mockBreaker)
	breaker.On("CanExecute").Return(true)
	breaker.On("RecordSuccess").Return()
	breaker.On("RecordFailure").Return()

	c := consumer.NewConsumer(source, useCase, idem, logger, breaker)
	go func() { _ = c.Start(ctx) }()
	time.Sleep(50 * time.Millisecond)

	idem.AssertNotCalled(t, "MarkAsProcessed", mock.Anything, "msg-3")
}

func TestConsumer_ConcurrentProcessing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgCh := make(chan amqp.Delivery, 3)
	defer close(msgCh)

	requests := []domain.SendEmailRequest{
		{To: "a@example.com", Template: "t1"},
		{To: "b@example.com", Template: "t2"},
		{To: "c@example.com", Template: "t3"},
	}

	for i, req := range requests {
		body, _ := json.Marshal(req)
		ack := new(MockAcknowledger)
		tag := uint64(i + 1)

		ack.On("Ack", tag, false).Return(nil)

		msgCh <- amqp.Delivery{
			Body:         body,
			MessageId:    fmt.Sprintf("msg-%d", i+1),
			Acknowledger: ack,
			DeliveryTag:  tag,
		}
	}

	source := new(mockSource)
	source.On("Consume", mock.Anything).Return((<-chan amqp.Delivery)(msgCh), nil)

	idem := new(mockIdempotency)
	for i := range requests {
		msgID := fmt.Sprintf("msg-%d", i+1)
		idem.On("IsProcessed", mock.Anything, msgID).Return(false, nil)
		idem.On("MarkAsProcessing", mock.Anything, msgID).Return(true, nil)
		idem.On("MarkAsProcessed", mock.Anything, msgID).Return(nil)
		idem.On("ClearProcessing", mock.Anything, msgID).Return(nil)
	}

	useCase := new(mockUseCase)
	for _, req := range requests {
		useCase.On("Send", req).Return(nil)
	}

	logger := new(mockLogger)
	logger.On("Println", mock.Anything).Maybe()
	logger.On("Printf", mock.Anything, mock.Anything).Maybe()

	breaker := new(mockBreaker)
	breaker.On("CanExecute").Return(true)
	breaker.On("RecordSuccess").Return()
	breaker.On("RecordFailure").Return()

	c := consumer.NewConsumer(source, useCase, idem, logger, breaker)

	go func() {
		_ = c.Start(ctx)
	}()

	time.Sleep(200 * time.Millisecond) // дати часу обробити

	source.AssertExpectations(t)
	idem.AssertExpectations(t)
	useCase.AssertExpectations(t)
	logger.AssertExpectations(t)
}
