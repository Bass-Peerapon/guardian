package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"guardian/internal/model"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
)

type Service interface {
	Health() map[string]string
	GetApps(ctx context.Context) ([]*model.Application, error)
	GetPerms(ctx context.Context) ([]*model.Permission, error)
	GetUsers(ctx context.Context) ([]*model.User, error)
	GetRoles(ctx context.Context, args *sync.Map) ([]*model.Role, error)
	GetApp(ctx context.Context, appID string) (*model.Application, error)
	GetPerm(ctx context.Context, permID string, appID string) (*model.Permission, error)
	GetUser(ctx context.Context, userName string) (*model.User, error)
	GetRole(ctx context.Context, roleID string, appID string) (*model.Role, error)
	UpsertApp(ctx context.Context, app *model.Application) error
	UpsertPerm(ctx context.Context, perm *model.Permission) error
	UpsertUser(ctx context.Context, user *model.User) error
	UpsertRole(ctx context.Context, role *model.Role) error
	DeleteApp(ctx context.Context, appID string) error
	DeletePerm(ctx context.Context, permID string, appID string) error
	DeleteUser(ctx context.Context, userName string) error
	DeleteRole(ctx context.Context, roleID string, appID string) error
}

type service struct {
	db *sql.DB
}

func (service *service) GetApps(ctx context.Context) ([]*model.Application, error) {
	sql := "SELECT * FROM applications"
	rows, err := service.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	apps := make([]*model.Application, 0)
	for rows.Next() {
		var app model.Application
		if err := rows.Scan(&app.ID, &app.Name, &app.Description); err != nil {
			return nil, err
		}
		apps = append(apps, &app)
	}
	return apps, nil
}

func (service *service) GetPerms(ctx context.Context) ([]*model.Permission, error) {
	sql := "SELECT * FROM permissions"
	rows, err := service.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	perms := make([]*model.Permission, 0)
	for rows.Next() {
		var perm model.Permission
		if err := rows.Scan(&perm.ID, &perm.AppID, &perm.Name, &perm.Description, &perm.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, &perm)
	}
	return perms, nil
}

func (service *service) GetUsers(ctx context.Context) ([]*model.User, error) {
	sql := `
	SELECT 
		users.username, 
		users.created_at, 
		users.updated_at,
	json_agg(json_build_object('id', roles.id, 'app_id', roles.app_id, 'name', roles.name, 'description', roles.description, 'created_at', roles.created_at::text, 'permissions', roles.permissions)) AS roles
	FROM 
		users 
	LEFT JOIN 
		user_roles 
	ON 
		users.username = user_roles.username 
	LEFT JOIN
		(
			SELECT
				roles.id,
				roles.app_id,
				roles.name,
				roles.description,
				roles.created_at,
	json_agg(json_build_object('id', permissions.id, 'app_id', permissions.app_id, 'name', permissions.name, 'description', permissions.description,'created_at', permissions.created_at::text)) AS permissions
			FROM
				roles
			JOIN
				role_permissions ON roles.id = role_permissions.role_id AND roles.app_id = role_permissions.app_id
			JOIN
				permissions ON permissions.id = role_permissions.permission_id AND permissions.app_id = role_permissions.app_id
			GROUP BY
				roles.id, roles.app_id, roles.name, roles.description, roles.created_at
		) AS roles 
	ON 
		user_roles.role_id = roles.id AND user_roles.app_id = roles.app_id
	GROUP BY
		users.username , users.created_at , users.updated_at
	ORDER BY
		users.updated_at DESC , users.username
	`
	rows, err := service.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]*model.User, 0)
	for rows.Next() {
		var user model.User
		user.Roles = make([]*model.Role, 0)
		var roles string
		if err := rows.Scan(&user.UserName, &user.CreatedAt, &user.UpdatedAt, &roles); err != nil {
			return nil, err
		}
		user.Roles = make([]*model.Role, 0)
		if err := json.Unmarshal([]byte(roles), &user.Roles); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func (service *service) GetRoles(ctx context.Context, args *sync.Map) ([]*model.Role, error) {
	var conds []string
	var vals []interface{}
	if args == nil {
		args = &sync.Map{}
	}
	if v, ok := args.Load("app_id"); ok {
		conds = append(conds, "roles.app_id = ?")
		vals = append(vals, v)
	}

	var where string
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}
	sql := fmt.Sprintf(`
	SELECT 
		roles.id,
		roles.app_id,
		roles.name,
		roles.description,
		roles.created_at,
	json_agg(json_build_object('id', permissions.id, 'app_id', permissions.app_id, 'name', permissions.name, 'description', permissions.description, 'created_at' , permissions.created_at::text)) AS permissions
	FROM 
		roles
	JOIN
		role_permissions ON roles.id = role_permissions.role_id AND roles.app_id = role_permissions.app_id
	JOIN
		permissions ON permissions.id = role_permissions.permission_id AND permissions.app_id = role_permissions.app_id
	%s
	GROUP BY
		roles.id, roles.app_id, roles.name, roles.description, roles.created_at
	`,
		where,
	)
	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	rows, err := service.db.QueryContext(ctx, sql, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := make([]*model.Role, 0)
	for rows.Next() {
		var role model.Role
		var perms string
		if err := rows.Scan(&role.ID, &role.AppID, &role.Name, &role.Description, &role.CreatedAt, &perms); err != nil {
			return nil, err
		}
		role.Permissions = make([]*model.Permission, 0)
		if err := json.Unmarshal([]byte(perms), &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

func (service *service) GetApp(ctx context.Context, appID string) (*model.Application, error) {
	sql := "SELECT * FROM applications WHERE id = $1"
	row := service.db.QueryRowContext(ctx, sql, appID)
	var app model.Application
	if err := row.Scan(&app.ID, &app.Name, &app.Description); err != nil {
		return nil, err
	}
	return &app, nil
}

func (service *service) GetPerm(ctx context.Context, permID string, appID string) (*model.Permission, error) {
	sql := "SELECT * FROM permissions WHERE id = $1 AND app_id = $2"
	row := service.db.QueryRowContext(ctx, sql, permID, appID)
	var perm model.Permission
	if err := row.Scan(&perm.ID, &perm.AppID, &perm.Name, &perm.Description, &perm.CreatedAt); err != nil {
		return nil, err
	}
	return &perm, nil
}

func (service *service) GetUser(ctx context.Context, userName string) (*model.User, error) {
	sql := `
	SELECT 
		users.username, 
		users.created_at, 
		users.updated_at,
	json_agg(json_build_object('id', roles.id, 'app_id', roles.app_id, 'name', roles.name, 'description', roles.description, 'created_at', roles.created_at::text, 'permissions', roles.permissions)) AS roles
	FROM 
		users 
	LEFT JOIN 
		user_roles 
	ON 
		users.username = user_roles.username 
	LEFT JOIN
		(
			SELECT
				roles.id,
				roles.app_id,
				roles.name,
				roles.description,
				roles.created_at,
	json_agg(json_build_object('id', permissions.id, 'app_id', permissions.app_id, 'name', permissions.name, 'description', permissions.description, 'created_at', permissions.created_at::text)) AS permissions
			FROM
				roles
			JOIN
				role_permissions ON roles.id = role_permissions.role_id AND roles.app_id = role_permissions.app_id
			JOIN
				permissions ON permissions.id = role_permissions.permission_id AND permissions.app_id = role_permissions.app_id
			GROUP BY
				roles.id, roles.app_id, roles.name, roles.description, roles.created_at
		) AS roles 
	ON 
		user_roles.role_id = roles.id AND user_roles.app_id = roles.app_id
	WHERE
		users.username = $1
	GROUP BY
		users.username , users.created_at , users.updated_at
	`
	row := service.db.QueryRowContext(ctx, sql, userName)
	var user model.User
	var roles string
	if err := row.Scan(&user.UserName, &user.CreatedAt, &user.UpdatedAt, &roles); err != nil {
		return nil, err
	}
	user.Roles = make([]*model.Role, 0)
	if err := json.Unmarshal([]byte(roles), &user.Roles); err != nil {
		return nil, err
	}
	return &user, nil
}

func (service *service) GetRole(ctx context.Context, roleID string, appID string) (*model.Role, error) {
	sql := `
	SELECT 
		roles.id,
		roles.app_id,
		roles.name,
		roles.description,
		roles.created_at,
		json_agg(json_build_object('id', permissions.id, 'app_id', permissions.app_id, 'name', permissions.name, 'description', permissions.description,'created_at' , permissions.created_at::text)) AS permissions
	FROM 
		roles
	JOIN
		role_permissions ON roles.id = role_permissions.role_id AND roles.app_id = role_permissions.app_id
	JOIN
		permissions ON permissions.id = role_permissions.permission_id AND permissions.app_id = role_permissions.app_id
	WHERE
		roles.id = $1 AND roles.app_id = $2
	GROUP BY
		roles.id, roles.app_id, roles.name, roles.description, roles.created_at
	`
	row := service.db.QueryRowContext(ctx, sql, roleID, appID)
	var role model.Role
	role.Permissions = make([]*model.Permission, 0)
	var perms string
	if err := row.Scan(&role.ID, &role.AppID, &role.Name, &role.Description, &role.CreatedAt, &perms); err != nil {
		return nil, err
	}
	role.Permissions = make([]*model.Permission, 0)
	if err := json.Unmarshal([]byte(perms), &role.Permissions); err != nil {
		return nil, err
	}

	return &role, nil
}

func (service *service) UpsertApp(ctx context.Context, app *model.Application) error {
	sql := "INSERT INTO applications (id, name, description) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET name = $2, description = $3"
	_, err := service.db.ExecContext(ctx, sql, app.ID, app.Name, app.Description)
	return err
}

func (service *service) UpsertPerm(ctx context.Context, perm *model.Permission) error {
	sql := "INSERT INTO permissions (id, app_id, name, description) VALUES ($1, $2, $3, $4) ON CONFLICT (id, app_id) DO UPDATE SET name = $3, description = $4"
	_, err := service.db.ExecContext(ctx, sql, perm.ID, perm.AppID, perm.Name, perm.Description)
	return err
}

func (service *service) UpsertUser(ctx context.Context, user *model.User) error {
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	sql := "INSERT INTO users (username) VALUES ($1) ON CONFLICT (username) DO UPDATE SET updated_at = NOW()"
	if _, err := tx.ExecContext(ctx, sql, user.UserName); err != nil {
		return err
	}

	sql = "DELETE FROM user_roles WHERE username = $1"
	if _, err := tx.ExecContext(ctx, sql, user.UserName); err != nil {
		return err
	}

	sql = "INSERT INTO user_roles (username, role_id, app_id) VALUES ($1, $2, $3)"
	for _, role := range user.Roles {
		if _, err := tx.ExecContext(ctx, sql, user.UserName, role.ID, role.AppID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (service *service) UpsertRole(ctx context.Context, role *model.Role) error {
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	sql := "INSERT INTO roles (id, app_id, name, description) VALUES ($1, $2, $3, $4) ON CONFLICT (id, app_id) DO UPDATE SET name = $3, description = $4"
	if _, err := tx.ExecContext(ctx, sql, role.ID, role.AppID, role.Name, role.Description); err != nil {
		return err
	}

	sql = "DELETE FROM role_permissions WHERE role_id = $1 AND app_id = $2"
	if _, err := tx.ExecContext(ctx, sql, role.ID, role.AppID); err != nil {
		return err
	}

	sql = "INSERT INTO role_permissions (role_id, permission_id, app_id) VALUES ($1, $2, $3)"
	for _, perm := range role.Permissions {
		if _, err := tx.ExecContext(ctx, sql, role.ID, perm.ID, role.AppID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (service *service) DeleteApp(ctx context.Context, appID string) error {
	sql := "DELETE FROM applications WHERE id = $1"
	_, err := service.db.ExecContext(ctx, sql, appID)
	return err
}

func (service *service) DeletePerm(ctx context.Context, permID string, appID string) error {
	sql := "DELETE FROM permissions WHERE id = $1 AND app_id = $2"
	_, err := service.db.ExecContext(ctx, sql, permID, appID)
	return err
}

func (service *service) DeleteUser(ctx context.Context, userName string) error {
	sql := "DELETE FROM users WHERE username = $1"
	_, err := service.db.ExecContext(ctx, sql, userName)
	return err
}

func (service *service) DeleteRole(ctx context.Context, roleID string, appID string) error {
	sql := "DELETE FROM roles WHERE id = $1 AND app_id = $2"
	_, err := service.db.ExecContext(ctx, sql, roleID, appID)
	return err
}

var (
	database = os.Getenv("DB_DATABASE")
	password = os.Getenv("DB_PASSWORD")
	username = os.Getenv("DB_USERNAME")
	port     = os.Getenv("DB_PORT")
	host     = os.Getenv("DB_HOST")
)

func New() Service {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	s := &service{db: db}
	return s
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.PingContext(ctx)
	if err != nil {
		log.Fatalf(fmt.Sprintf("db down: %v", err))
	}

	return map[string]string{
		"message": "It's healthy",
	}
}
