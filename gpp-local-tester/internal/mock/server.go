package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/project-piper/gpp-local-tester/internal/config"
)

type Server struct {
	config     config.MockServicesConfig
	httpServer *http.Server
	stats      *Stats
	mu         sync.RWMutex
}

type Stats struct {
	TotalRequests int
	MethodCounts  map[string]int
	PathCounts    map[string]int
	Requests      []RequestLog
}

type RequestLog struct {
	Timestamp time.Time
	Method    string
	Path      string
	Query     string
	Headers   http.Header
	Body      string
}

func NewServer(cfg config.MockServicesConfig) *Server {
	return &Server{
		config: cfg,
		stats: &Stats{
			MethodCounts: make(map[string]int),
			PathCounts:   make(map[string]int),
			Requests:     make([]RequestLog, 0),
		},
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start server in background
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Mock server error: %v\n", err)
		}
	}()

	// Wait a bit to ensure server started
	time.Sleep(100 * time.Millisecond)

	return nil
}

func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Log request
	blue := color.New(color.FgBlue).SprintFunc()
	timestamp := time.Now()
	fmt.Printf("[%s] %s %s\n", timestamp.Format("15:04:05"), blue(r.Method), r.URL.Path)

	// Record stats
	s.mu.Lock()
	s.stats.TotalRequests++
	s.stats.MethodCounts[r.Method]++
	s.stats.PathCounts[r.URL.Path]++
	s.mu.Unlock()

	// Handle CORS preflight
	if r.Method == "OPTIONS" {
		s.setCORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Find matching endpoint
	endpoint := s.matchEndpoint(r.URL.Path, r.Method)
	if endpoint == nil {
		s.handleNotFound(w, r)
		return
	}

	// Set headers
	s.setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(endpoint.Response.Status)

	// Write response
	if err := json.NewEncoder(w).Encode(endpoint.Response.Body); err != nil {
		fmt.Printf("  → Error encoding response: %v\n", err)
	} else {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("  → %s (matched: %s)\n", green(endpoint.Response.Status), endpoint.Path)
	}
}

func (s *Server) matchEndpoint(requestPath, requestMethod string) *config.MockEndpoint {
	for i := range s.config.Endpoints {
		endpoint := &s.config.Endpoints[i]

		// Check method
		if endpoint.Method != "*" && !strings.EqualFold(endpoint.Method, requestMethod) {
			continue
		}

		// Convert path pattern to regex
		pattern := endpoint.Path
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		pattern = "^" + pattern + "$"

		matched, err := regexp.MatchString(pattern, requestPath)
		if err != nil {
			continue
		}

		if matched {
			return endpoint
		}
	}

	return nil
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	red := color.New(color.FgRed).SprintFunc()
	fmt.Printf("  → %s (no matching endpoint)\n", red("404"))

	s.setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)

	response := map[string]string{
		"error":   "Not Found",
		"message": fmt.Sprintf("No mock endpoint configured for %s %s", r.Method, r.URL.Path),
		"hint":    "Add this endpoint to config.yaml mockServices.endpoints",
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func (s *Server) GetStats() *Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	statsCopy := &Stats{
		TotalRequests: s.stats.TotalRequests,
		MethodCounts:  make(map[string]int),
		PathCounts:    make(map[string]int),
	}

	for k, v := range s.stats.MethodCounts {
		statsCopy.MethodCounts[k] = v
	}
	for k, v := range s.stats.PathCounts {
		statsCopy.PathCounts[k] = v
	}

	return statsCopy
}
