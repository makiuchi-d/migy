CREATE TABLE `_migrations` (
  `id` int NOT NULL,
  `applied` datetime,
  `title` varchar(255),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;

INSERT INTO `_migrations` (`id`,`applied`,`title`) VALUES
  (1, '2025-04-19 00:33:32', 'first');

CREATE TABLE `users` (
  `id` int NOT NULL,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;

INSERT INTO `users` (`id`,`name`) VALUES
  (1, 'alice'), (2, 'bob'), (3, 'carol');

DELIMITER //

CREATE PROCEDURE _migration_exists(IN input_id INTEGER)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM _migrations WHERE id = input_id) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'migration not found';
  END IF;
END//

DELIMITER ;
