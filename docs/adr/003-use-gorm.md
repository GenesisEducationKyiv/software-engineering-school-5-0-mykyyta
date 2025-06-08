# ADR-003: Use GORM for PostgreSQL ORM

## Status
Accepted

## Date
2025-05-14

## Tags
- orm
- gorm
- database
- backend

## Context

As someone with experience in Python, I was used to working with high-level ORMs like SQLAlchemy and Django ORM.

Being new to Go, I wanted to reduce the complexity and time spent writing raw SQL queries and handling manual DB mappings.  
Although Go encourages writing SQL directly (or using lightweight tools like `sqlx`), I chose an ORM to help me focus on learning Go and completing the project efficiently.

GORM is the most widely adopted ORM in the Go ecosystem, with good documentation and active community support.

## Decision

Use **GORM** as the ORM for working with PostgreSQL.

## Alternatives Considered

- **Raw SQL with `database/sql`**: More control, idiomatic, but time-consuming and error-prone for a beginner.
- **`sqlx`**: Lightweight extension over `database/sql`, but still requires manual query writing.
- **`ent`**: Type-safe and modern, but adds complexity and a generation step.

## Consequences

- Faster development and simplified DB operations.
- Slightly less idiomatic Go code due to GORM’s abstractions.
- May need to learn how to debug ORM behavior in some edge cases.
- Easier transition for someone used to Python ORMs.

## Related Requirements

- Constraint: "Use PostgreSQL"
- Goal: Deliver a working API while learning Go