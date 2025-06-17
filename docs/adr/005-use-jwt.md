# ADR-005: Use JWT for authentication

## Status
Accepted

## Date
2025-05-14

## Tags
- auth
- jwt
- security
- backend

## Context

The task was to find a suitable method for handling user subscriptions in a secure way — including confirmation and preference updates — without requiring a full user management system.

## Decision rationale

The application needed a simple way to authenticate users for protected actions like confirming subscriptions.

Given the scope of the project and the fact that it doesn't require full user sessions or social login, I opted for stateless token-based authentication.

Coming from a Python background, I was familiar with both session-based and token-based auth. I chose JWT because:

- It’s simple and widely used
- It doesn’t require server-side session storage
- It integrates well with APIs and mobile clients
- It allowed me to gain hands-on experience with Go’s JWT libraries

## Alternatives Considered

- **Session-based auth**: Would require server-side session store (e.g., Redis), adds complexity.
- **API key**: Simpler, but doesn’t support user-specific access or expiration.
- **OAuth2**: Too heavy for this use case.

## Final decision

Use **JWT (JSON Web Tokens)** to authenticate users in API requests.

## Consequences

- Stateless, scalable authentication
- Requires signing key management
- Tokens must be verified on every request
- Easy to use with frontend or external systems if needed

## Related Requirements

- Non-functional: "API must support authenticated access to subscription endpoints"
- Constraint: "No user interface or complex session management"