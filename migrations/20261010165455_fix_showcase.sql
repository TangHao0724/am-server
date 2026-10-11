-- +goose Up
SELECT 'up SQL query';

ALTER TABLE "game_play_stages_test"
RENAME COLUMN "gameplay_id" TO"game_id";

ALTER TABLE "game_play_stages_test"
ADD FOREIGN KEY("game_id") REFERENCES "games_test"("id")
ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE "game_play_stages_test"
ADD COLUMN user_id INTEGER;


ALTER TABLE "game_play_stages_test"
ADD FOREIGN KEY("user_id") REFERENCES "users"("id")
ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE "games_test"
ADD FOREIGN KEY("creator_id") REFERENCES "users"("id")
ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
SELECT 'down SQL query';
