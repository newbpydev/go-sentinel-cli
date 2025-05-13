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
	testHandler := handlers.NewTestResultsHandler(nil) // We'll inject the real TestRunner later
	metricsHandler := handlers.NewMetricsHandler()
	websocketHandler := handlers.NewWebSocketHandler()

	server := &Server{
		router:          r,
		templates:       tmpl,
		staticPath:      staticPath,
		testHandler:     testHandler,
		metricsHandler:  metricsHandler,
		websocketHandler: websocketHandler,
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

	// Dashboard route
	s.router.Get("/", s.handleDashboard)

	// API routes
	s.router.Route("/api", func(r chi.Router) {
		// Test routes
		r.Get("/tests", s.testHandler.GetTestResults)
		r.Post("/run-test/{testName}", s.testHandler.RunTest)
		r.Post("/run-tests", s.testHandler.RunAllTests)
		r.Get("/tests/filter", s.testHandler.FilterTestResults)
		
		// Metrics routes
		r.Get("/metrics", s.metricsHandler.GetMetrics)
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
	}
	
	s.render(w, "pages/dashboard", data)
}

// render executes the named template with the given data
func (s *Server) render(w http.ResponseWriter, name string, data interface{}) {
	err := s.templates.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}
