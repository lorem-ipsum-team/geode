-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS geo (
  id UUID PRIMARY KEY,
  location GEOGRAPHY(POINT, 4326) NOT NULL,
  geo_updated TIMESTAMP NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "geo";
-- +goose StatementEnd
