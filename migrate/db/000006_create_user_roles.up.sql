CREATE TABLE user_roles (
  username VARCHAR(255) NOT NULL REFERENCES users(username) ON DELETE CASCADE,
  role_id VARCHAR(255) NOT NULL ,
  app_id VARCHAR(255) NOT NULL REFERENCES applications(id) ON DELETE CASCADE,

  PRIMARY KEY (username, role_id, app_id),
  FOREIGN KEY(role_id, app_id) REFERENCES roles(id, app_id) ON DELETE CASCADE
);
