package middleware

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"go-gin-api/config"
	"go-gin-api/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func rbacFreshDB(t *testing.T) *gorm.DB {
	// unique in-memory DB per test for full isolation
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.AutoMigrate(&models.Permission{}, &models.Role{}, &models.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	config.DB = db
	return db
}

// helpers ──────────────────────────────

func createRoleInDB(t *testing.T, db *gorm.DB, name string) models.Role {
	r := models.Role{Name: name}
	if err := db.Create(&r).Error; err != nil {
		t.Fatalf("create role %s: %v", name, err)
	}
	return r
}

func createUserWithRole(t *testing.T, db *gorm.DB, email string, role *models.Role) models.User {
	u := models.User{Email: email, Name: "Tester"}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user %s: %v", email, err)
	}
	if role != nil {
		if err := db.Model(&u).Association("Roles").Append(role); err != nil {
			t.Fatalf("append role: %v", err)
		}
	}
	return u
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestRequireRole_AllowedRole — user has a matching role; handler must proceed.
func TestRequireRole_AllowedRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := rbacFreshDB(t)

	role := createRoleInDB(t, db, "editor")
	user := createUserWithRole(t, db, "rbac_allowed@example.com", &role)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("user_id", user.ID)

	RequireRole("editor", "super_admin")(c)

	if c.IsAborted() {
		t.Fatalf("expected handler to proceed for allowed role, got abort: code=%d body=%s", w.Code, w.Body.String())
	}
}

// TestRequireRole_ForbiddenRole — user has a different role; must get 403.
func TestRequireRole_ForbiddenRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := rbacFreshDB(t)

	role := createRoleInDB(t, db, "viewer")
	user := createUserWithRole(t, db, "rbac_forbidden@example.com", &role)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("user_id", user.ID)

	RequireRole("super_admin")(c)

	if !c.IsAborted() {
		t.Fatalf("expected 403 abort for missing role")
	}
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

// TestRequireRole_NoRoleAtAll — user exists but has no roles assigned.
func TestRequireRole_NoRoleAtAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := rbacFreshDB(t)

	user := createUserWithRole(t, db, "rbac_norole@example.com", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("user_id", user.ID)

	RequireRole("admin")(c)

	if !c.IsAborted() {
		t.Fatalf("expected 403 for user with no roles")
	}
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

// TestRequireRole_NoUserIDInContext — user_id not set; must get 401.
func TestRequireRole_NoUserIDInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rbacFreshDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	RequireRole("admin")(c)

	if !c.IsAborted() {
		t.Fatalf("expected abort when user_id missing from context")
	}
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// TestRequireRole_UserNotFoundInDB — user_id set but user doesn't exist in DB.
func TestRequireRole_UserNotFoundInDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rbacFreshDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("user_id", uint(99999))

	RequireRole("admin")(c)

	if !c.IsAborted() {
		t.Fatalf("expected abort for non-existent user")
	}
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// TestRequireRole_MultipleAllowedRoles — any one matching role is sufficient.
func TestRequireRole_MultipleAllowedRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := rbacFreshDB(t)

	role := createRoleInDB(t, db, "moderator")
	user := createUserWithRole(t, db, "rbac_multi@example.com", &role)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("user_id", user.ID)

	RequireRole("super_admin", "moderator", "editor")(c)

	if c.IsAborted() {
		t.Fatalf("expected handler to proceed when user has one of allowed roles")
	}
}
