-- +goose Up
SELECT 'up SQL query';

ALTER TABLE "stages" 
ADD COLUMN "name" VARCHAR(255);

-- +goose Down
SELECT 'down SQL query';
