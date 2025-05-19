package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// errorTemplate returns a *template.Template that always errors on ExecuteTemplate
func errorTemplate(name string) *template.Template {
	t := template.New(name)
	t.Funcs(template.FuncMap{"error": func() (string, error) { return "", errors.New("template error") }})
	tmpl, _ := t.Parse(`{{error}}`)
	return tmpl
}

func TestNewCoverageHandler_NotNil(t *testing.T) {
	h := NewCoverageHandler(template.New("test"))
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.templates == nil {
		t.Error("expected templates to be set")
	}
}

func TestGetCoverageSummary_SuccessAndError(t *testing.T) {
	tmpl := template.Must(template.New("partials/coverage-summary").Parse(`{{.TotalCoverage}}`))
	h := &CoverageHandler{templates: tmpl}
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.GetCoverageSummary(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "78.5") {
		t.Errorf("expected summary output, got %q", w.Body.String())
	}
	// Error path
	h.templates = errorTemplate("partials/coverage-summary")
	w2 := httptest.NewRecorder()
	h.GetCoverageSummary(w2, r)
	if w2.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w2.Code)
	}
}

func TestGetCoverageFiles_AllFiltersAndSearch(t *testing.T) {
	tmpl := template.Must(template.New("partials/coverage-file-list").Parse(`{{len .Files}}`))
	h := &CoverageHandler{templates: tmpl}
	baseURL := "/?page=1"
	tests := []struct {
		name   string
		query  string
		expect int
	}{
		{"all", "", 5},
		{"low", "filter=low", 1},
		{"medium", "filter=medium", 2},
		{"high", "filter=high", 2},
		{"search-match", "search=runner", 1},
		{"search-none", "search=notfound", 0},
		{"page-out-of-bounds", "page=99", 5},
	}
	for _, tc := range tests {
		r := httptest.NewRequest("GET", baseURL+"&"+tc.query, nil)
		w := httptest.NewRecorder()
		h.GetCoverageFiles(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", tc.name, w.Code)
		}
		if !strings.Contains(w.Body.String(), strconv.Itoa(tc.expect)) {
			t.Errorf("%s: expected %d files, got %q", tc.name, tc.expect, w.Body.String())
		}
	}
	// Error path
	h.templates = errorTemplate("partials/coverage-file-list")
	r := httptest.NewRequest("GET", baseURL, nil)
	w := httptest.NewRecorder()
	h.GetCoverageFiles(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetFileDetail_EdgeCases(t *testing.T) {
	tmpl := template.Must(template.New("partials/coverage-file-detail").Parse(`{{.FileID}}`))
	h := &CoverageHandler{templates: tmpl}
	// Valid file
	r := httptest.NewRequest("GET", "/?id=file1", nil)
	w := httptest.NewRecorder()
	h.GetFileDetail(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "file1") {
		t.Errorf("expected file1 in output, got %q", w.Body.String())
	}
	// Invalid file
	r2 := httptest.NewRequest("GET", "/?id=notfound", nil)
	w2 := httptest.NewRecorder()
	h.GetFileDetail(w2, r2)
	if w2.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w2.Code)
	}
	// Error path
	h.templates = errorTemplate("partials/coverage-file-detail")
	r3 := httptest.NewRequest("GET", "/?id=file1", nil)
	w3 := httptest.NewRecorder()
	h.GetFileDetail(w3, r3)
	if w3.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w3.Code)
	}
}

func TestSearchCoverage(t *testing.T) {
	tmpl := template.Must(template.New("partials/coverage-file-list").Parse(`search`))
	h := &CoverageHandler{templates: tmpl}
	r := httptest.NewRequest("GET", "/?search=runner", nil)
	w := httptest.NewRecorder()
	h.SearchCoverage(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	// Error path
	h.templates = errorTemplate("partials/coverage-file-list")
	w2 := httptest.NewRecorder()
	h.SearchCoverage(w2, r)
	if w2.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w2.Code)
	}
}

func TestFilterCoverage(t *testing.T) {
	tmpl := template.Must(template.New("partials/coverage-file-list").Parse(`filter`))
	h := &CoverageHandler{templates: tmpl}
	r := httptest.NewRequest("GET", "/?filter=low", nil)
	w := httptest.NewRecorder()
	h.FilterCoverage(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	// Error path
	h.templates = errorTemplate("partials/coverage-file-list")
	w2 := httptest.NewRecorder()
	h.FilterCoverage(w2, r)
	if w2.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w2.Code)
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	cases := []struct {
		s, substr string
		expect    bool
	}{
		{"HelloWorld", "world", true},
		{"HelloWorld", "WORLD", true},
		{"HelloWorld", "nope", false},
		{"", "", true},
	}
	for _, c := range cases {
		if got := containsIgnoreCase(c.s, c.substr); got != c.expect {
			t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", c.s, c.substr, got, c.expect)
		}
	}
}
