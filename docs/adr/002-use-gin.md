# ADR-002: Use Gin as the web framework

## Status
Accepted

## Date
2025-05-14

## Tags
- framework
- gin
- backend

## Context

The task was to choose an appropriate web framework for implementing the HTTP layer of the Weather API in Go.

## Decision rationale

Coming from a Python background, I'm used to working with web frameworks like Flask Django or FastAPI.  
As a beginner in Go, I didn’t feel confident building HTTP routing, middleware, and request handling from scratch using `net/http`.

Although Go encourages minimalism and writing everything manually, I opted for a framework to speed up development and reduce boilerplate.

Gin is one of the most popular and well-documented Go frameworks, with built-in support for routing, middleware, JSON binding, and request handling.

## Alternatives Considered

- **net/http**: Fully standard and idiomatic, but too low-level for my current experience.
- **Echo**: Also fast and lightweight, but smaller community.
- **Fiber**: Inspired by Express.js, but Gin has more documentation and examples.

## Final decision

Use the **Gin** web framework for building the HTTP API.

## Consequences

- Faster development and less boilerplate thanks to built-in features.
- Some abstraction over standard library, which may hide some Go internals.
- Slight deviation from Go’s philosophy of “just use the standard library”, but worth it at this learning stage.

## Related Requirements

- Functional: "Expose a REST API for weather and subscription features"
- Constraint: "Use Go as the primary language"