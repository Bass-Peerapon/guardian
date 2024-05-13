CREATE TABLE role_permissions (
  role_id VARCHAR(255) NOT NULL ,
  permission_id VARCHAR(255) NOT NULL ,
  app_id VARCHAR(255) NOT NULL REFERENCES applications(id) ON DELETE CASCADE,

  FOREIGN KEY(role_id, app_id) REFERENCES roles(id, app_id) ON DELETE CASCADE,
  FOREIGN KEY(permission_id, app_id) REFERENCES permissions(id, app_id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id, app_id)
);
