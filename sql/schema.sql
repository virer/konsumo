DROP TABLE IF EXISTS user_data CASCADE;
DROP TABLE IF EXISTS user;
CREATE TABLE user (
  user_id VARCHAR(255) PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  profile_pic TEXT NOT NULL,
  location VARCHAR(255) NULL,
  INDEX (user_id)
);

CREATE TABLE user_data (
  id INT AUTO_INCREMENT PRIMARY KEY,
  date DATE NOT NULL,
  type ENUM('gazoline', 'water', 'electricy', 'other'),
  value1 MEDIUMINT NOT NULL,
  value2 MEDIUMINT NOT NULL,
  user_id VARCHAR(255),
  CONSTRAINT FK_user_id FOREIGN KEY (user_id) REFERENCES user(user_id)
);
