package subscription

import (
	"context"
	"gorm.io/gorm"
)

type GormSubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *GormSubscriptionRepository {
	return &GormSubscriptionRepository{db: db}
}

func (r *GormSubscriptionRepository) GetByEmail(ctx context.Context, email string) (*Subscription, error) {
	var sub Subscription
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&sub).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *GormSubscriptionRepository) Create(ctx context.Context, sub *Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *GormSubscriptionRepository) Update(ctx context.Context, sub *Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *GormSubscriptionRepository) GetConfirmedByFrequency(ctx context.Context, frequency string) ([]Subscription, error) {
	var subs []Subscription
	err := r.db.WithContext(ctx).
		Where("is_confirmed = ? AND is_unsubscribed = ? AND frequency = ?", true, false, frequency).
		Find(&subs).Error
	if err != nil {
		return nil, err
	}
	return subs, nil
}
