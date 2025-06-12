package repo

import (
	"context"
	"weatherApi/internal/subscription"

	"gorm.io/gorm"
)

type GormSubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *GormSubscriptionRepository {
	return &GormSubscriptionRepository{db: db}
}

func (r *GormSubscriptionRepository) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&sub).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *GormSubscriptionRepository) Create(ctx context.Context, sub *subscription.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *GormSubscriptionRepository) Update(ctx context.Context, sub *subscription.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}
