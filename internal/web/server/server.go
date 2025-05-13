package server

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	
	"github.com/newbpydev/go-sentinel/internal/web/handlers"
	customMiddleware "github.com/newbpydev/go-sentinel/internal/web/middleware"
)

// Server represents the web server for Go Sentinel
type Server struct {
	router          *chi.Mux
	templates       *template.Template
	staticPath      string
	testHandler     *handlers.TestResultsHandler
	metricsHandler  *handlers.MetricsHandler
	websocketHandler *handlers.WebSocketHandler
	historyHandler   *handlers.HistoryHandler
}

// NewServer creates a new web server instance
func NewServer(templatePath, staticPath string) (*Server, error) {
	// Parse templates
	// First try the .tmpl extension
	tmpl := template.New("")

	// Parse layout templates first
	layoutFiles, err := filepath.Glob(filepath.Join(templatePath, "layouts/*.tmpl"))
	if err != nil {
		return nil, err
	}
	
	if len(layoutFiles) > 0 {
		tmpl, err = tmpl.ParseFiles(layoutFiles...)
		if err != nil {
			return nil, err
		}
	}
	
	// Parse page templates
	pageFiles, err := filepath.Glob(filepath.Join(templatePath, "pages/*.tmpl"))
	if err != nil {
		return nil, err
	}
	
	if len(pageFiles) > 0 {
		tmpl, err = tmpl.ParseFiles(pageFiles...)
		if err != nil {
			return nil, err
		}
	}
	
	// Parse partial templates
	partialFiles, err := filepath.Glob(filepath.Join(templatePath, "partials/*.tmpl"))
	if err != nil {
		return nil, err
	}
	
	if len(partialFiles) > 0 {
		tmpl, err = tmpl.ParseFiles(partialFiles...)
		if err != nil {
			return nil, err
		}
	}
	
	log.Printf("Loaded templates: layouts=%d, pages=%d, partials=%d", 
		len(layoutFiles), len(pageFiles), len(partialFiles))

	// Create router with middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.Logger)

	// Initialize handlers
	testHandler := handlers.NewTestResultsHandler(tmpl)
	metricsHandler := handlers.NewMetricsHandler()
	websocketHandler := handlers.NewWebSocketHandler()
	historyHandler := handlers.NewHistoryHandler()

	server := &Server{
		router:          r,
		templates:       tmpl,
		staticPath:      staticPath,
		testHandler:     testHandler,
		metricsHandler:  metricsHandler,
		websocketHandler: websocketHandler,
		historyHandler:   historyHandler,
	}

	// Register routes
	server.registerRoutes()

	// Start WebSocket broadcaster
	websocketHandler.StartBroadcaster()

	return server, nil
}

// registerRoutes sets up all the HTTP routes
func (s *Server) registerRoutes() {
	// Serve static files
	fileServer := http.FileServer(http.Dir(s.staticPath))
	s.router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Page routes
	s.router.Get("/", s.handleDashboard)
	s.router.Get("/tests", s.handleTests)
	s.router.Get("/reports", s.handleReports)
	s.router.Get("/history", s.handleHistory)
	s.router.Get("/settings", s.handleSettings)

	// API routes
	s.router.Route("/api", func(r chi.Router) {
		// Test routes
		r.Get("/tests", s.testHandler.GetTestResults)
		r.Post("/run-test/{testName}", s.testHandler.RunTest)
		r.Post("/run-tests", s.testHandler.RunAllTests)
		r.Get("/tests/filter", s.testHandler.FilterTestResults)
		
		// Metrics routes
		r.Get("/metrics", s.metricsHandler.GetMetrics)

		// Notification test route
		r.Post("/notifications/test", handlers.HandleTestNotification)
		
		// History routes
		r.Get("/history", s.historyHandler.GetTestRunHistory)
		r.Get("/history/compare", s.historyHandler.CompareTestRuns)
		r.Get("/history/{runID}", s.historyHandler.GetTestRunDetails)
	})

	// WebSocket route
	s.router.Get("/ws", s.websocketHandler.HandleWebSocket)
}

// Start begins listening on the given address
func (s *Server) Start(addr string) error {
	log.Printf("Starting web server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// handleDashboard renders the main dashboard
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Get initial data for the dashboard
	metrics := handlers.GetMetricsData()
	tests := handlers.GetMockTestResults()
	
	data := map[string]interface{}{
		"Title": "Test Dashboard",
		"Stats": metrics,
		"Tests": tests,
		"ActivePage": "dashboard",
	}
	
	s.render(w, "pages/dashboard", data)
}

// handleHistory renders the test history page
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Test History",
		"ActivePage": "history",
	}
	
	s.render(w, "pages/history", data)
}

// handleSettings renders the settings page
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Settings",
		"ActivePage": "settings",
	}
	
	s.render(w, "pages/settings", data)
}

// handleTests renders the tests page
func (s *Server) handleTests(w http.ResponseWriter, r *http.Request) {
	// Get test data for the page
	tests := handlers.GetMockTestResults()
	
	data := map[string]interface{}{
		"Title": "Tests",
		"ActivePage": "tests",
		"Tests": tests,
	}
	
	s.render(w, "pages/tests", data)
}

// handleReports renders the reports page
func (s *Server) handleReports(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Reports",
		"ActivePage": "reports",
		"Subtitle": "View and analyze your Go Sentinel reports",
	}
	
	s.render(w, "pages/reports", data)
}

// render executes the named template with the given data
func (s *Server) render(w http.ResponseWriter, name string, data interface{}) {
	err := s.templates.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}
