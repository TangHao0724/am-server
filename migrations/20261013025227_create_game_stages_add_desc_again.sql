-- +goose Up
SELECT 'up SQL query';

ALTER TABLE "stages" 
ADD COLUMN "description" TEXT;
-- +goose Down
SELECT 'down SQL query';
