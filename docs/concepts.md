# Concepts

This document explains the underlying concepts and ideas of this microservice architecture.

## Endpoints

A key concept of the architecture is the concept of endpoints. Endpoints are a canonical representation of a hierarchical structure and are further used instead of:

- resource locations
- function handles
- event types
- event topics
- queues names

All endpoints start with a forward slash (`/`), which is followed by one or more lowercase, hyphen-delimited, plural nouns (`/[a-z-]+/`) terminated by a forward slash after each. Instead of a noun, the endpoint may also contain a single-level (`*`) or a multi-level (`**`) wildcard. Finally the endpoint ends either with a verb (e.g. `find`) or a verbal adjective (e.g. `found`). A few examples are shown below.

- `/endpoints/create`
- `/endpoints/found`
- `/providers/external/*`
- `/items/*/pending`
- `/**`
