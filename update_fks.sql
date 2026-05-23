ALTER TABLE comments DROP FOREIGN KEY fk_comments_review;
ALTER TABLE comments ADD CONSTRAINT fk_comments_review FOREIGN KEY (review_id) REFERENCES reviews(id) ON DELETE CASCADE;

ALTER TABLE comments DROP FOREIGN KEY fk_comments_user;
ALTER TABLE comments ADD CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE reviews DROP FOREIGN KEY fk_reviews_user;
ALTER TABLE reviews ADD CONSTRAINT fk_reviews_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE reviews DROP FOREIGN KEY fk_reviews_game;
ALTER TABLE reviews ADD CONSTRAINT fk_reviews_game FOREIGN KEY (target_id) REFERENCES games(id) ON DELETE CASCADE;

ALTER TABLE activity_logs DROP FOREIGN KEY fk_activity_logs_user;
ALTER TABLE activity_logs ADD CONSTRAINT fk_activity_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE follows DROP FOREIGN KEY fk_follows_follower;
ALTER TABLE follows ADD CONSTRAINT fk_follows_follower FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE follows DROP FOREIGN KEY fk_follows_following;
ALTER TABLE follows ADD CONSTRAINT fk_follows_following FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE game_logs DROP FOREIGN KEY fk_game_logs_user;
ALTER TABLE game_logs ADD CONSTRAINT fk_game_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE game_logs DROP FOREIGN KEY fk_game_logs_game;
ALTER TABLE game_logs ADD CONSTRAINT fk_game_logs_game FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE;

ALTER TABLE likes DROP FOREIGN KEY fk_likes_user;
ALTER TABLE likes ADD CONSTRAINT fk_likes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE lists DROP FOREIGN KEY fk_lists_user;
ALTER TABLE lists ADD CONSTRAINT fk_lists_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE list_entries DROP FOREIGN KEY fk_lists_entries;
ALTER TABLE list_entries ADD CONSTRAINT fk_lists_entries FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE;

ALTER TABLE list_entries DROP FOREIGN KEY fk_list_entries_game;
ALTER TABLE list_entries ADD CONSTRAINT fk_list_entries_game FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE;

ALTER TABLE notifications DROP FOREIGN KEY fk_notifications_receiver;
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_receiver FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE notifications DROP FOREIGN KEY fk_notifications_sender;
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_sender FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ratings DROP FOREIGN KEY fk_ratings_user;
ALTER TABLE ratings ADD CONSTRAINT fk_ratings_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ratings DROP FOREIGN KEY fk_ratings_game;
ALTER TABLE ratings ADD CONSTRAINT fk_ratings_game FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE;
