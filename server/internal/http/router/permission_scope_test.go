package router

import (
	"io/fs"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go/ast"
	"go/parser"
	"go/token"
)

func TestAdminGroupScopeDoesNotAffectCommentCreate(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	v2 := app.Group("/api").Group("/v2")

	requireAuth := func(c *fiber.Ctx) error {
		if strings.TrimSpace(c.Get("Authorization")) == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.Next()
	}
	requireAdmin := func(c *fiber.Ctx) error {
		if c.Get("X-Admin") != "1" {
			return c.Status(fiber.StatusForbidden).SendString("需要管理员权限")
		}
		return c.Next()
	}

	commentGroup := v2.Group("/comments", requireAuth)
	commentGroup.Post("/areas/:areaId", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	adminGroup := v2.Group("/admin", requireAuth, requireAdmin)
	adminGroup.Put("/comments/areas/:areaId/close", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	commentReq := httptest.NewRequest("POST", "/api/v2/comments/areas/34", strings.NewReader(`{"content":"hi"}`))
	commentReq.Header.Set("Authorization", "Bearer user-token")
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("X-Admin", "0")
	commentResp, err := app.Test(commentReq, -1)
	if err != nil {
		t.Fatalf("comment request failed: %v", err)
	}
	if commentResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected comment create status 200, got %d", commentResp.StatusCode)
	}

	adminReq := httptest.NewRequest("PUT", "/api/v2/admin/comments/areas/34/close", strings.NewReader(`{"isClosed":true}`))
	adminReq.Header.Set("Authorization", "Bearer user-token")
	adminReq.Header.Set("Content-Type", "application/json")
	adminReq.Header.Set("X-Admin", "0")
	adminResp, err := app.Test(adminReq, -1)
	if err != nil {
		t.Fatalf("admin request failed: %v", err)
	}
	if adminResp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected admin endpoint status 403, got %d", adminResp.StatusCode)
	}
}

func TestNoEmptyPrefixGroupOnV2(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to locate current file")
	}
	dir := filepath.Dir(currentFile)

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(info fs.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse dir failed: %v", err)
	}

	var violations []string
	for _, pkg := range pkgs {
		for fileName, fileNode := range pkg.Files {
			ast.Inspect(fileNode, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok || len(call.Args) == 0 {
					return true
				}

				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel == nil || selector.Sel.Name != "Group" {
					return true
				}

				baseIdent, ok := selector.X.(*ast.Ident)
				if !ok || baseIdent.Name != "v2" {
					return true
				}

				pathLit, ok := call.Args[0].(*ast.BasicLit)
				if !ok || pathLit.Kind != token.STRING {
					return true
				}
				path, err := strconv.Unquote(pathLit.Value)
				if err != nil || path != "" {
					return true
				}

				pos := fset.Position(call.Pos())
				violations = append(violations, fileName+":"+strconv.Itoa(pos.Line))
				return true
			})
		}
	}

	if len(violations) > 0 {
		t.Fatalf("found empty-prefix v2 groups, please avoid v2.Group(\"\"): %v", violations)
	}
}
