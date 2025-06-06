package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"weatherApi/internal/model"
	"weatherApi/pkg/jwtutil"
	"weatherApi/pkg/weatherapi"

	emailutil "weatherApi/pkg/email"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Allows replacing weatherapi.CityExists in tests.
var cityValidator = func(ctx context.Context, city string) (bool, error) {
	return weatherapi.CityExists(ctx, city)
}

type SubscribeRequest struct {
	Email     string `form:"email" binding:"required,email"`
	City      string `form:"city" binding:"required"`
	Frequency string `form:"frequency" binding:"required,oneof=daily hourly"`
}

// - sends confirmation email asynchronously.
func subscribeHandler(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := validateCity(c.Request.Context(), req.City); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingSub, err := checkExistingSubscription(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	token, err := generateToken(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}

	if existingSub != nil {
		// Update existing unconfirmed/unsubscribed subscription with new data and token
		if err := updateSubscription(existingSub, req, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
			return
		}
	} else {
		// Create new subscription
		if err := createSubscription(req, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save subscription"})
			return
		}
	}

	// Send confirmation email in a separate goroutine
	sendConfirmationEmailAsync(req.Email, token)
	c.JSON(http.StatusOK, gin.H{"message": "Subscription successful. Confirmation email sent."})
}

// validateCity checks if the requested city exists using the external weather API.
func validateCity(ctx context.Context, city string) error {
	ok, err := cityValidator(ctx, city)
	if err != nil {
		return fmt.Errorf("Failed to validate city")
	}
	if !ok {
		return fmt.Errorf("City not found")
	}
	return nil
}

// or an error if the email is already subscribed and confirmed.
func checkExistingSubscription(req SubscribeRequest) (*model.Subscription, error) {
	var existing model.Subscription
	err := DB.Where("email = ?", req.Email).First(&existing).Error
	if err == nil {
		if existing.IsConfirmed && !existing.IsUnsubscribed {
			return nil, fmt.Errorf("Email already subscribed")
		}
		return &existing, nil
	}
	return nil, nil
}

// generateToken creates a JWT for email confirmation and unsubscribe links.
func generateToken(email string) (string, error) {
	return jwtutil.Generate(email)
}

// createSubscription saves a new unconfirmed subscription to the database.
func createSubscription(req SubscribeRequest, token string) error {
	sub := model.Subscription{
		ID:             uuid.New().String(),
		Email:          req.Email,
		City:           req.City,
		Frequency:      req.Frequency,
		IsConfirmed:    false,
		IsUnsubscribed: false,
		Token:          token,
		CreatedAt:      time.Now(),
	}
	return DB.Create(&sub).Error
}

// updateSubscription updates an existing subscription with new values and resets confirmation status.
func updateSubscription(sub *model.Subscription, req SubscribeRequest, token string) error {
	sub.City = req.City
	sub.Frequency = req.Frequency
	sub.Token = token
	sub.CreatedAt = time.Now()
	sub.IsConfirmed = false
	sub.IsUnsubscribed = false
	return DB.Save(sub).Error
}

// sendConfirmationEmailAsync sends the confirmation email in a background goroutine.
func sendConfirmationEmailAsync(email, token string) {
	go func() {
		if err := emailutil.SendConfirmationEmail(email, token); err != nil {
			fmt.Printf("Failed to send confirmation email to %s: %v\n", email, err)
		}
	}()
}
