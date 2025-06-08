# Functional Requirements — Weather API

## 1. User Subscription
- Users can submit their email and city to subscribe
- A confirmation email is sent to verify the subscription (double opt-in)

## 2. Email Confirmation
- User must click a link with a secure token to confirm subscription
- Confirmed subscriptions are marked as active

## 3. Unsubscribe
- Users can unsubscribe using a secure token link
- Unsubscribed users are excluded from future emails

## 4. Weather Retrieval
- Public API allows users to get current weather by city

## 5. Email Delivery
- System sends weather updates to all active subscribers
- Emails are sent based on the selected frequency ("daily" or "hourly")