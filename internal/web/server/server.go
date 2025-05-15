package server

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/newbpydev/go-sentinel/internal/web/handlers"
	customMiddleware "github.com/newbpydev/go-sentinel/internal/web/middleware"
)

// Server represents the web server for Go Sentinel
type Server struct {
	router           *chi.Mux
	templates        *template.Template // master: only layouts+partials
	templatePath     string             // store the root path for templates
	staticPath       string             // store the root path for static files
	testHandler      *handlers.TestResultsHandler
	metricsHandler   *handlers.MetricsHandler
	websocketHandler *handlers.WebSocketHandler
	historyHandler   *handlers.HistoryHandler
	coverageHandler  *handlers.CoverageHandler
	settingsHandler  *handlers.SettingsHandler
}

// NewServer creates a new web server instance
func NewServer(templatePath, staticPath string) (*Server, error) {
	// 1) Create a root Template with any custom funcs
	funcMap := template.FuncMap{
		"year": func() int { return time.Now().Year() },
		"add":  func(a, b int) int { return a + b },
		"sub":  func(a, b int) int { return a - b },
		"mul":  func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}
	tmpl := template.New("").Funcs(funcMap)

	// 2) Parse in order: layouts → partials → pages
	layouts, err := filepath.Glob(filepath.Join(templatePath, "layouts", "*.tmpl"))
	if err != nil {
		return nil, err
	}
	if len(layouts) > 0 {
		tmpl = template.Must(tmpl.ParseFiles(layouts...))
	}

	partials, err := filepath.Glob(filepath.Join(templatePath, "partials", "*.tmpl"))
	if err != nil {
		return nil, err
	}
	if len(partials) > 0 {
		tmpl = template.Must(tmpl.ParseFiles(partials...))
	}

	// pages, err := filepath.Glob(filepath.Join(templatePath, "pages", "*.tmpl"))
	// if err != nil {
	// 	return nil, err
	// }
	// if len(pages) > 0 {
	// 	tmpl = template.Must(tmpl.ParseFiles(pages...))
	// }

	log.Printf("Master template loaded: layouts=%d, partials=%d",
		len(layouts), len(partials))
	// Debug: list all defined templates/blocks
	// log.Println("Defined templates:", tmpl.DefinedTemplates())

	// 3) Chi router + middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer, middleware.RealIP, customMiddleware.Logger)

	// 4) Initialize your handlers
	testHandler := handlers.NewTestResultsHandler(tmpl)
	metricsHandler := handlers.NewMetricsHandler()
	websocketHandler := handlers.NewWebSocketHandler()
	historyHandler := handlers.NewHistoryHandler()
	coverageHandler := handlers.NewCoverageHandler(tmpl)
	settingsHandler := handlers.NewSettingsHandler(tmpl)

	server := &Server{
		router:           r,
		templates:        tmpl,
		templatePath:     templatePath,
		staticPath:       staticPath,
		testHandler:      testHandler,
		metricsHandler:   metricsHandler,
		websocketHandler: websocketHandler,
		historyHandler:   historyHandler,
		coverageHandler:  coverageHandler,
		settingsHandler:  settingsHandler,
	}

	// 5) Register routes
	server.registerRoutes()

	// 6) Start WebSocket broadcaster
	websocketHandler.StartBroadcaster()

	// return &Server{
	// 	router:           r,
	// 	templates:        tmpl,
	// 	templatePath:     templatePath,
	// 	staticPath:       staticPath,
	// 	testHandler:      testHandler,
	// 	metricsHandler:   metricsHandler,
	// 	websocketHandler: websocketHandler,
	// 	historyHandler:   historyHandler,
	// 	coverageHandler:  coverageHandler,
	// 	settingsHandler:  settingsHandler,
	// }, nil
	return server, nil
}

// registerRoutes sets up all the HTTP routes
func (s *Server) registerRoutes() {
	// Static files
	fileServer := http.FileServer(http.Dir(s.staticPath))
	s.router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Page routes: all use block overrides in the page templates
	s.router.Get("/", s.render("dashboard", map[string]interface{}{
		"Title":                 "Dashboard",
		"ActivePage":            "dashboard",
		"ShowTestManagement":    false,
		"ShowTestConfiguration": false,
	}))
	s.router.Get("/tests", s.render("tests", map[string]interface{}{
		"Title":                 "Tests",
		"Tests":                 handlers.GetMockTestResults(),
		"ActivePage":            "tests",
		"ShowTestManagement":    true,
		"ShowTestConfiguration": true,
	}))
	s.router.Get("/reports", s.render("reports", map[string]interface{}{
		"Title":                 "Reports",
		"ActivePage":            "reports",
		"Subtitle":              "View and analyze your Go Sentinel reports",
		"ShowTestManagement":    false,
		"ShowTestConfiguration": false,
	}))
	s.router.Get("/history", s.render("history", map[string]interface{}{
		"Title":                 "Test History",
		"ActivePage":            "history",
		"ShowTestManagement":    false,
		"ShowTestConfiguration": false,
	}))
	s.router.Get("/settings", s.render("settings", map[string]interface{}{
		"Title":                 "Settings",
		"ActivePage":            "settings",
		"ShowTestManagement":    false,
		"ShowTestConfiguration": false,
		"Settings":              handlers.GetDefaultSettings(),
	}))
	s.router.Get("/coverage", s.render("coverage", map[string]interface{}{
		"Title":                 "Coverage",
		"ActivePage":            "coverage",
		"Subtitle":              "Visualize your test coverage",
		"ShowTestManagement":    false,
		"ShowTestConfiguration": false,
	}))

	// API routes
	s.router.Route("/api", func(r chi.Router) {
		r.Use(customMiddleware.ToastErrorHandler)

		// Test routes
		r.Get("/tests", s.testHandler.GetTestResults)
		r.Post("/run-test/{testName}", s.testHandler.RunTest)
		r.Post("/run-tests", s.testHandler.RunAllTests)
		r.Get("/tests/filter", s.testHandler.FilterTestResults)

		// Metrics
		r.Get("/metrics", s.metricsHandler.GetMetrics)

		// Notifications
		r.Post("/notifications/test", handlers.HandleTestNotification)

		// History
		r.Get("/history", s.historyHandler.GetTestRunHistory)
		r.Get("/history/compare", s.historyHandler.CompareTestRuns)
		r.Get("/history/{runID}", s.historyHandler.GetTestRunDetails)

		// Coverage
		r.Get("/coverage/summary", s.coverageHandler.GetCoverageSummary)
		r.Get("/coverage/files", s.coverageHandler.GetCoverageFiles)
		r.Get("/coverage/file/{id}", s.coverageHandler.GetFileDetail)
		r.Get("/coverage/search", s.coverageHandler.SearchCoverage)
		r.Get("/coverage/filter", s.coverageHandler.FilterCoverage)

		// Settings
		r.Get("/settings", s.settingsHandler.GetSettings)
		r.Post("/settings/validate", s.settingsHandler.ValidateSettings)
		r.Post("/settings/save", s.settingsHandler.SaveSettings)
		r.Get("/settings/reset", s.settingsHandler.ResetSettings)
		r.Delete("/settings/clear-cache", s.settingsHandler.ClearCache)
		r.Delete("/settings/clear-history", s.settingsHandler.ClearHistory)

		// Toast demo
		r.Get("/toast/test/{type}", s.handleToastTest)
	})

	// WebSocket
	s.router.Get("/ws", s.websocketHandler.HandleWebSocket)

	// Not found
	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})
}

// render returns a handler that injects base-template blocks
func (s *Server) render(pageName string, baseData map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) Clone the **un-executed** master
		t, err := s.templates.Clone()
		if err != nil {
			log.Printf("Template clone error: %v", err)
			http.Error(w, "Template clone error", http.StatusInternalServerError)
			return
		}

		// 2) Parse only the single page file into the clone
		pageFile := filepath.Join(s.templatePath, "pages", pageName+".tmpl")
		log.Printf("Rendering page: %s", pageFile)
		if _, err := t.ParseFiles(pageFile); err != nil {
			log.Printf("Template parse error (%s): %v", pageName, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 3) Inject year & headers (always make sure the Content-Type is set)
		baseData["Year"] = time.Now().Year()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// 4 ) Execute the base layout on the clone, which will picl the correct blocks
		if err := t.ExecuteTemplate(w, "base", baseData); err != nil {
			log.Printf("Template execution error (%s): %v", pageName, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// Start begins listening on the given address
func (s *Server) Start(addr string) error {
	log.Printf("Starting web server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}
