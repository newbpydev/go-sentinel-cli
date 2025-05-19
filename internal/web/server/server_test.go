package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewServer_ValidPaths(t *testing.T) {
	// Create temporary directories for templates and static files
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "templates")
	staticPath := filepath.Join(tmpDir, "static")

	// Create required directories
	if err := os.MkdirAll(filepath.Join(templatePath, "layouts"), 0750); err != nil {
		t.Fatalf("Failed to create layouts directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(templatePath, "partials"), 0750); err != nil {
		t.Fatalf("Failed to create partials directory: %v", err)
	}
	if err := os.MkdirAll(staticPath, 0750); err != nil {
		t.Fatalf("Failed to create static directory: %v", err)
	}

	// Create a sample template file
	layoutContent := `{{define "base"}}{{block "content" .}}{{end}}{{end}}`
	if err := os.WriteFile(filepath.Join(templatePath, "layouts", "base.tmpl"), []byte(layoutContent), 0600); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Test server creation
	server, err := NewServer(templatePath, staticPath)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	if server == nil {
		t.Fatal("Expected non-nil server")
	}
}

func TestNewServer_InvalidPaths(t *testing.T) {
	tests := []struct {
		name        string
		templateDir string
		staticDir   string
	}{
		{"non-existent template dir", "/nonexistent/templates", "static"},
		{"non-existent static dir", "templates", "/nonexistent/static"},
		{"empty paths", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewServer(tc.templateDir, tc.staticDir)
			if err == nil {
				t.Error("Expected error for invalid paths")
			}
		})
	}
}

func TestServer_RegisterRoutes(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "templates")
	staticPath := filepath.Join(tmpDir, "static")

	// Create required directories and files
	dirs := []string{
		filepath.Join(templatePath, "layouts"),
		filepath.Join(templatePath, "partials"),
		filepath.Join(templatePath, "pages"), // Ensure pages dir exists
		staticPath,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create base template
	baseTemplate := `{{define "base"}}{{block "content" .}}{{end}}{{end}}`
	if err := os.WriteFile(filepath.Join(templatePath, "layouts", "base.tmpl"), []byte(baseTemplate), 0600); err != nil {
		t.Fatalf("Failed to create base template: %v", err)
	}

	// Create minimal page templates for each route
	pageNames := []string{"dashboard", "tests", "reports", "history", "settings", "coverage"}
	for _, name := range pageNames {
		pageContent := `{{define "content"}}Test Content{{end}}`
		pagePath := filepath.Join(templatePath, "pages", name+".tmpl")
		if err := os.WriteFile(pagePath, []byte(pageContent), 0600); err != nil {
			t.Fatalf("Failed to create page template %s: %v", name, err)
		}
	}

	// Create server
	server, err := NewServer(templatePath, staticPath)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// Test routes
	routes := []struct {
		path   string
		method string
		status int
	}{
		{"/", "GET", http.StatusOK},
		{"/tests", "GET", http.StatusOK},
		{"/reports", "GET", http.StatusOK},
		{"/history", "GET", http.StatusOK},
		{"/settings", "GET", http.StatusOK},
		{"/coverage", "GET", http.StatusOK},
		{"/static/nonexistent.css", "GET", http.StatusNotFound},
	}

	for _, route := range routes {
		t.Run(route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)
			if w.Code != route.status {
				t.Errorf("Expected status %d for %s %s, got %d", route.status, route.method, route.path, w.Code)
			}
		})
	}
}

func TestServer_RenderTemplate(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "templates")
	staticPath := filepath.Join(tmpDir, "static")

	// Create required directories
	if err := os.MkdirAll(filepath.Join(templatePath, "layouts"), 0750); err != nil {
		t.Fatalf("Failed to create layouts directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(templatePath, "pages"), 0750); err != nil {
		t.Fatalf("Failed to create pages directory: %v", err)
	}
	if err := os.MkdirAll(staticPath, 0750); err != nil {
		t.Fatalf("Failed to create static directory: %v", err)
	}

	// Create test templates
	templates := map[string]string{
		"layouts/base.tmpl": `{{define "base"}}Title: {{.Title}}{{block "content" .}}{{end}}{{end}}`,
	}

	for name, content := range templates {
		path := filepath.Join(templatePath, name)
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create template %s: %v", name, err)
		}
	}
	// Create required page template
	pageTmpl := `{{define "content"}}Content: {{.Content}}{{end}}`
	if err := os.WriteFile(filepath.Join(templatePath, "pages", "page.tmpl"), []byte(pageTmpl), 0600); err != nil {
		t.Fatalf("Failed to create page template: %v", err)
	}

	// Create server
	server, err := NewServer(templatePath, staticPath)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// Test template rendering
	data := map[string]interface{}{
		"Title":   "Test Page",
		"Content": "Test Content",
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := server.render("page", data)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Title: Test Page") {
		t.Error("Expected title in output")
	}
	if !strings.Contains(body, "Content: Test Content") {
		t.Error("Expected content in output")
	}
}

func TestServer_TemplateFunctions(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "templates")
	staticPath := filepath.Join(tmpDir, "static")

	// Create required directories
	if err := os.MkdirAll(filepath.Join(templatePath, "layouts"), 0750); err != nil {
		t.Fatalf("Failed to create layouts directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(templatePath, "pages"), 0750); err != nil {
		t.Fatalf("Failed to create pages directory: %v", err)
	}
	if err := os.MkdirAll(staticPath, 0750); err != nil {
		t.Fatalf("Failed to create static directory: %v", err)
	}

	// Create test template that uses custom functions
	tmplContent := `{{define "base"}}
		Year: {{year}}
		Add: {{add 2 3}}
		Sub: {{sub 5 2}}
		Mul: {{mul 4 3}}
		Div: {{div 6 2}}
		Div by zero: {{div 1 0}}
	{{end}}`

	if err := os.WriteFile(filepath.Join(templatePath, "layouts", "base.tmpl"), []byte(tmplContent), 0600); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}
	// Create required page template
	pageTmpl := `{{define "content"}}{{end}}`
	if err := os.WriteFile(filepath.Join(templatePath, "pages", "base.tmpl"), []byte(pageTmpl), 0600); err != nil {
		t.Fatalf("Failed to create page template: %v", err)
	}

	// Create server
	server, err := NewServer(templatePath, staticPath)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// Test template functions
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := server.render("base", nil)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	currentYear := time.Now().Year()
	expectedOutputs := map[string]string{
		"Add":         "Add: 5",
		"Sub":         "Sub: 3",
		"Mul":         "Mul: 12",
		"Div":         "Div: 3",
		"Div by zero": "Div by zero: 0",
		"Year":        fmt.Sprintf("Year: %d", currentYear),
	}

	for name, expected := range expectedOutputs {
		if !strings.Contains(strings.ReplaceAll(body, "\n", ""), expected) {
			t.Errorf("%s: expected %q in output, got: %q", name, expected, body)
		}
	}
}
