# Concepts

This document explains the underlying concepts and ideas of this microservice architecture.

## Channels

A key concept of the architecture is the concept of channels. Channels are a canonical representation of a hierarchical structure. Channels are identified by one or more lowercase, hyphen-delimited, plural nouns (`/[a-z-]+/`). A single dot (`.`) is used to define a new hierarchy level. The format is inspired by the [type attribute of a CloudEvent][cloud-event-type]. The table below illustrates the correlation between channels and other concepts of representing hierarchy levels in other protocols.

| Concept       | Resource action | Wildcard resource | Wildcard action |
| ------------- | --------------- | ----------------- | --------------- |
| HTTP endpoint | `POST /pets`    | `POST /*`         | `ALL /pets`     |
| MQTT topic    | `pets/create`   | `#/create`        | `pets/#`        |
| Channels      | `pets.create`   | `>.create`        | `pets.>`        |

### Special channels

_TODO: Document the concept of **persistent channels** also referred to as **streams**._

[cloud-event-type]: https://github.com/cloudevents/spec/blob/v1.0.1/spec.md#type
