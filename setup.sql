--- You should run this query using `mysql -u root < setup_db.sql`

DROP DATABASE IF EXISTS database;
CREATE DATABASE database;

USE database

CREATE USER 'user'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON database.* TO 'user'@'localhost' IDENTIFIED BY 'password';
