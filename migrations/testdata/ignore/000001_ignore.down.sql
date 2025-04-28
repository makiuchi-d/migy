
/*
 * migy:ignore table1.column1, table1.column2
 */
ALTER TABLE table1 ADD COLUMN column1 ADD COLUMN column2;

CREATE TABLE table2 ( -- migy:ignore table2.*
  `id` int NOT NULL, 
  `str` text,
  PRIMARY KEY (`id`)
)
