package server

import (
	"guardian/internal/model"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", s.HelloWorldHandler)
	e.GET("/health", s.healthHandler)

	e.GET("/apps", s.GetAppsHandler)
	e.GET("/apps/:appID", s.GetAppHandler)
	e.POST("/apps", s.UpsertAppHandler)
	e.DELETE("/apps/:appID", s.DeleteAppHandler)

	e.GET("/permissions", s.GetPermsHandler)
	e.GET("/permissions/:permID/:appID", s.GetPermHandler)
	e.POST("/permissions", s.UpsertPermHandler)
	e.DELETE("/permissions/:permID/:appID", s.DeletePermHandler)

	e.GET("/roles", s.GetRolesHandler)
	e.GET("/roles/:roleID/:appID", s.GetRoleHandler)
	e.POST("/roles", s.UpsertRoleHandler)
	e.DELETE("/roles/:roleID/:appID", s.DeleteRoleHandler)

	e.GET("/users", s.GetUsersHandler)
	e.GET("/users/:userName", s.GetUserHandler)
	e.POST("/users", s.UpsertUserHandler)
	e.DELETE("/users/:userName", s.DeleteUserHandler)

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) UpsertAppHandler(c echo.Context) error {
	app := new(model.Application)
	if err := c.Bind(app); err != nil {
		return err
	}

	if err := s.db.UpsertApp(c.Request().Context(), app); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]string{
		"message": "ok",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) GetAppsHandler(c echo.Context) error {
	apps, err := s.db.GetApps(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, apps)
}

func (s *Server) DeleteAppHandler(c echo.Context) error {
	appID := c.Param("appID")
	if err := s.db.DeleteApp(c.Request().Context(), appID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]string{
		"message": "ok",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) GetAppHandler(c echo.Context) error {
	appID := c.Param("appID")
	app, err := s.db.GetApp(c.Request().Context(), appID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, app)
}

func (s *Server) UpsertPermHandler(c echo.Context) error {
	perm := new(model.Permission)
	if err := c.Bind(perm); err != nil {
		return err
	}

	if err := s.db.UpsertPerm(c.Request().Context(), perm); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]string{
		"message": "ok",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) GetPermsHandler(c echo.Context) error {
	perms, err := s.db.GetPerms(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, perms)
}

func (s *Server) DeletePermHandler(c echo.Context) error {
	permID := c.Param("permID")
	appID := c.Param("appID")
	if err := s.db.DeletePerm(c.Request().Context(), permID, appID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]string{
		"message": "ok",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) GetPermHandler(c echo.Context) error {
	permID := c.Param("permID")
	appID := c.Param("appID")
	perm, err := s.db.GetPerm(c.Request().Context(), permID, appID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, perm)
}

func (s *Server) GetRolesHandler(c echo.Context) error {
	appID := c.QueryParam("app_id")
	args := new(sync.Map)
	if appID != "" {
		args.Store("app_id", appID)
	}
	roles, err := s.db.GetRoles(c.Request().Context(), args)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, roles)
}

func (s *Server) UpsertRoleHandler(c echo.Context) error {
	role := new(model.Role)
	if err := c.Bind(role); err != nil {
		return err
	}

	if err := s.db.UpsertRole(c.Request().Context(), role); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]string{
		"message": "ok",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) DeleteRoleHandler(c echo.Context) error {
	roleID := c.Param("roleID")
	appID := c.Param("appID")
	if err := s.db.DeleteRole(c.Request().Context(), roleID, appID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]string{
		"message": "ok",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) GetRoleHandler(c echo.Context) error {
	roleID := c.Param("roleID")
	appID := c.Param("appID")
	role, err := s.db.GetRole(c.Request().Context(), roleID, appID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, role)
}

func (s *Server) GetUsersHandler(c echo.Context) error {
	users, err := s.db.GetUsers(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, users)
}

func (s *Server) UpsertUserHandler(c echo.Context) error {
	user := new(model.User)
	if err := c.Bind(user); err != nil {
		return err
	}

	if err := s.db.UpsertUser(c.Request().Context(), user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]string{
		"message": "ok",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) DeleteUserHandler(c echo.Context) error {
	userName := c.Param("userName")
	if err := s.db.DeleteUser(c.Request().Context(), userName); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]string{
		"message": "ok",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) GetUserHandler(c echo.Context) error {
	userName := c.Param("userName")
	user, err := s.db.GetUser(c.Request().Context(), userName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}
