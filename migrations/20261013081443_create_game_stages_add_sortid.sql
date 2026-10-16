-- +goose Up
SELECT 'up SQL query';

ALTER TABLE "stages" 
ADD COLUMN "sort_id" INTEGER NOT NULL;

ALTER TABLE "next_stages" 
DROP COLUMN "sort_id";
-- +goose Down
SELECT 'down SQL query';
