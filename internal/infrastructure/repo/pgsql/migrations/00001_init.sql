-- +goose Up
CREATE TABLE api_keys (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    name        text        NOT NULL,
    key_hash    bytea       NOT NULL UNIQUE,
    source      text        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    revoked_at  timestamptz
);

CREATE TABLE event_types (
    type            text PRIMARY KEY,               -- 'auth.login'
    description     text        NOT NULL DEFAULT '',
    enabled         boolean     NOT NULL DEFAULT true
);

CREATE TABLE events (
    id           bigint GENERATED ALWAYS AS IDENTITY,
    event_id     uuid        NOT NULL,              -- генерирует клиент, для идемпотентности
    occurred_at  timestamptz NOT NULL,              -- когда событие случилось в приложении
    received_at  timestamptz NOT NULL DEFAULT now(),-- когда пришло в коллектор
    source       text        NOT NULL,              -- сервис/модуль-источник: 'api', 'worker', ...
    type         text        NOT NULL,              -- 'auth.login', 'payment.failed', ...
    level        smallint    NOT NULL DEFAULT 20,   -- 10 debug, 20 info, 30 warn, 40 error
    actor_id     text,                              -- id пользователя в основном приложении (без FK)
    request_id   text,                              -- корреляция с запросом / trace
    payload      jsonb       NOT NULL DEFAULT '{}',
    PRIMARY KEY (occurred_at, id),
    UNIQUE (occurred_at, event_id)
) PARTITION BY RANGE (occurred_at);

CREATE INDEX events_type_time_idx  ON events (type, occurred_at DESC);
CREATE INDEX events_actor_time_idx ON events (actor_id, occurred_at DESC) WHERE actor_id IS NOT NULL;
CREATE INDEX events_request_idx    ON events (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX events_id_idx ON events (id);

CREATE TABLE events_default PARTITION OF events DEFAULT;

-- +goose Down
DROP TABLE events;
DROP TABLE event_types;
DROP TABLE api_keys;