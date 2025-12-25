-- You should run this query using `mysql -u root < setup_db.sql`

DROP DATABASE IF EXISTS `database`;
CREATE DATABASE `database`;

DROP DATABASE IF EXISTS `dev`;
CREATE DATABASE `dev`;

USE `database`;

CREATE USER IF NOT EXISTS `user`@`localhost` IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON `database`.* TO `user`@`localhost`;
GRANT ALL PRIVILEGES ON `dev`.* TO `user`@`localhost`;
FLUSH PRIVILEGES;
