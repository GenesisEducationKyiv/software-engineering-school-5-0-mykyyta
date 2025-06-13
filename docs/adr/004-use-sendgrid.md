# ADR-004: Use SendGrid for email delivery

## Status
Accepted

## Date
2025-05-14

## Tags
- email
- sendgrid
- infrastructure
- communication

## Context

The task was to choose an email delivery solution for sending subscription confirmation and notification emails.

## Decision rationale

The project required sending emails for subscription confirmation and notifications.

In previous personal projects, I used basic SMTP setups, which work but are not ideal for production — they can be unreliable, hard to configure securely, and lack analytics or rate control.

For this project, I wanted to explore a production-grade email delivery service. SendGrid offers:

- A free tier suitable for development and testing
- An official Go SDK
- A clear API for managing email templates, sender identity, and API keys
- Built-in support for domain verification and email deliverability tools

## Alternatives Considered

- **SMTP**: Already used in the past, but would require self-managed email server or Gmail SMTP (less reliable).
- **Amazon SES**: Powerful and production-ready, but more complex to set up for this small project.
- **Mailgun**: Also good, but SendGrid had clearer documentation and better SDK support for Go.

## Final decision

Use **SendGrid** to handle outgoing email in the project.

## Consequences

- Faster and more reliable delivery than SMTP
- More realistic experience with a modern SaaS email tool

## Related Requirements

- Functional: "Send confirmation email after user subscribes"
- Constraint: "Avoid hardcoded secrets, use secure delivery"