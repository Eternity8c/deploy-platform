-- +goose Up
CREATE TABLE applications (
      id         bigserial PRIMARY KEY,
  tenant     text        NOT NULL DEFAULT 'shop',
  name       text        NOT NULL,
  spec       jsonb       NOT NULL,   -- весь spec как есть
  version    integer     NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant, name)
);

-- +goose Down
DROP TABLE applications;