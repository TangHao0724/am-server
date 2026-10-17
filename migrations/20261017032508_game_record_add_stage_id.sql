-- +goose Up
SELECT 'up SQL query';

ALTER TABLE "game_plays" 
ADD COLUMN "current_stage_status_id" INTEGER;

ALTER TABLE "game_plays" 
ADD FOREIGN KEY("current_stage_status_id") REFERENCES "game_play_stages"("id")
ON UPDATE NO ACTION ON DELETE NO ACTION;
-- +goose Down
SELECT 'down SQL query';
