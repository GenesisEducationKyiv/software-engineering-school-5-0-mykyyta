package subscription

import (
	"context"
	"errors"
	"time"
	"weatherApi/internal/domain"

	"gorm.io/gorm"
)

type SubscriptionRecord struct {
	ID             string `gorm:"primaryKey"`
	Email          string `gorm:"not null;uniqueIndex"`
	City           string `gorm:"not null"`
	Frequency      string `gorm:"type:text;not null"`
	IsConfirmed    bool   `gorm:"default:false"`
	IsUnsubscribed bool   `gorm:"default:false"`
	Token          string `gorm:"not null"`
	CreatedAt      time.Time
}

func toRecord(s domain.Subscription) SubscriptionRecord {
	return SubscriptionRecord{
		ID:             s.ID,
		Email:          s.Email,
		City:           s.City,
		Frequency:      string(s.Frequency),
		IsConfirmed:    s.IsConfirmed,
		IsUnsubscribed: s.IsUnsubscribed,
		Token:          s.Token,
		CreatedAt:      s.CreatedAt,
	}
}

func fromRecord(r SubscriptionRecord) domain.Subscription {
	return domain.Subscription{
		ID:             r.ID,
		Email:          r.Email,
		City:           r.City,
		Frequency:      domain.Frequency(r.Frequency),
		IsConfirmed:    r.IsConfirmed,
		IsUnsubscribed: r.IsUnsubscribed,
		Token:          r.Token,
		CreatedAt:      r.CreatedAt,
	}
}

type GormSubscriptionRepository struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *GormSubscriptionRepository {
	return &GormSubscriptionRepository{db: db}
}

func (r *GormSubscriptionRepository) GetByEmail(ctx context.Context, email string) (*domain.Subscription, error) {
	var rec SubscriptionRecord
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}
	sub := fromRecord(rec)
	return &sub, nil
}

func (r *GormSubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	rec := toRecord(*sub)
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *GormSubscriptionRepository) Update(ctx context.Context, sub *domain.Subscription) error {
	rec := toRecord(*sub)
	return r.db.WithContext(ctx).Save(&rec).Error
}

func (r *GormSubscriptionRepository) GetConfirmedByFrequency(ctx context.Context, freq string) ([]domain.Subscription, error) {
	var recs []SubscriptionRecord
	err := r.db.WithContext(ctx).
		Where("is_confirmed = ? AND is_unsubscribed = ? AND frequency = ?", true, false, freq).
		Find(&recs).Error
	if err != nil {
		return nil, err
	}

	subs := make([]domain.Subscription, 0, len(recs))
	for _, r := range recs {
		subs = append(subs, fromRecord(r))
	}
	return subs, nil
}

func (SubscriptionRecord) TableName() string {
	return "subscriptions"
}
