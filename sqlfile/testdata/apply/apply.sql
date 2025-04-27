
CREATE TABLE memo (
  id      int NOT NULL,
  title   varchar(255),
  created datetime,
  PRIMARY KEY (id)
);

INSERT INTO memo (id, title, created) VALUES
  (1, 'first', '2025-04-27 10:00:00');

INSERT INTO memo (id, title, created) VALUES
  (2, 'secnd', '2025-04-27 11:00:00'),
  (3, 'third', '2025-04-27 12:00:00');

