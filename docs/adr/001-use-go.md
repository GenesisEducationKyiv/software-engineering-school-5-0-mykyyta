# ADR-001: Use Go as the programming language

## Status
Accepted

## Date
2025-05-14

## Tags
- language
- backend

## Context

The project required building a Weather API using Go, Node.js, or PHP. Python, despite being my most familiar language, was not allowed. This led to an evaluation of the allowed options.

## Decision rationale

The task was to implement a Weather API using one of the allowed languages: Go, Node.js, or PHP.

Although I have experience with Python, it was not allowed. I considered Node.js and Go.  
I chose Go for its performance, simplicity, and relevance in modern backend and cloud-native systems.

This project also served as an opportunity to learn Go through practice.

## Alternatives Considered

- **Node.js**: Fast to prototype, familiar ecosystem, but weak typing.
- **PHP**: Less relevant for modern backend development.

## Final decision

Go was chosen as the main programming language for the project.

## Consequences

- Slower initial progress due to learning curve.
- Gained valuable experience with a modern backend language.

## Related Requirements

- Constraint: "Must use Go, Node.js, or PHP"