/*
Copyright © 2025 Furkan Pehlivan furkanpehlivan34@gmail.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package web

import (
	"archive/zip"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"sync"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/pehlicd/crd-wizard/internal/ai"
	"github.com/pehlicd/crd-wizard/internal/generator"
	"github.com/pehlicd/crd-wizard/internal/giturl"
	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/logger"
	"github.com/pehlicd/crd-wizard/internal/models"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	ClusterManager *k8s.ClusterManager
	router         *http.ServeMux
	server         *http.Server
	aiClient       *ai.Client
	log            *logger.Logger
	startTime      time.Time
}

func NewServer(clusterManager *k8s.ClusterManager, port string, aiClient *ai.Client, log *logger.Logger) *Server {
	r := http.NewServeMux()
	s := &Server{
		ClusterManager: clusterManager,
		router:         r,
		server: &http.Server{
			Addr:         ":" + port,
			Handler:      r,
			ReadTimeout:  15 * time.Minute,
			WriteTimeout: 15 * time.Minute,
			IdleTimeout:  15 * time.Minute,
		},
		aiClient:  aiClient,
		log:       log,
		startTime: time.Now(),
	}
	s.registerHandlers()
	return s
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) registerHandlers() {
	apiRouter := s.router
	apiRouter.HandleFunc("/clusters", s.ClustersHandler)
	apiRouter.HandleFunc("/cluster-info", s.ClusterInfoHandler)
	apiRouter.HandleFunc("/crds", s.CrdsHandler)
	apiRouter.HandleFunc("/crs", s.CrsHandler)
	apiRouter.HandleFunc("/cr", s.CrHandler)
	apiRouter.HandleFunc("/events", s.EventsHandler)
	apiRouter.HandleFunc("/resource-graph", s.ResourceGraphHandler)
	if s.aiClient != nil {
		apiRouter.HandleFunc("/crd/generate-context", s.GenerateCrdContextHandler)
	}
	apiRouter.HandleFunc("/status", s.Status)
	apiRouter.HandleFunc("/export", s.ExportHandler)
	apiRouter.HandleFunc("/export-all", s.ExportAllHandler)
	apiRouter.HandleFunc("/generate", s.GenerateHandler)
	s.router.Handle("/api/", http.StripPrefix("/api", s.log.Middleware(apiRouter)))

	// Health endpoint is registered without logging middleware to avoid noise in logs
	s.router.HandleFunc("/health", s.HealthHandler)

	staticFS, _ := fs.Sub(staticFiles, "static")
	uiFile := http.FS(staticFS)
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFiles(uiFile, w, r, "index.html")
	})
	s.router.HandleFunc("/instances", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFiles(uiFile, w, r, "instances.html")
	})
	s.router.HandleFunc("/resource", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFiles(uiFile, w, r, "resource.html")
	})
	s.router.HandleFunc("/generator", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFiles(uiFile, w, r, "generator.html")
	})
}

func serveStaticFiles(staticFS http.FileSystem, w http.ResponseWriter, r *http.Request, defaultFile string) {
	path := r.URL.Path
	if path == "/" {
		path = "/" + defaultFile
	}

	file, err := staticFS.Open(path)
	if err != nil {
		file, err = staticFS.Open("/" + defaultFile)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	defer file.Close()

	// Get the file information
	fileInfo, err := file.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeContent(w, r, path, fileInfo.ModTime(), file)
}

type statusResponse struct {
	Uptime    string `json:"uptime"`
	AIEnabled bool   `json:"aiEnabled"`
}

func (s *Server) Status(w http.ResponseWriter, _ *http.Request) {
	resp := statusResponse{
		Uptime:    time.Since(s.startTime).String(),
		AIEnabled: s.aiClient != nil,
	}
	s.respondWithJSON(w, http.StatusOK, resp)
}

// generateContextRequest defines the expected JSON body for the AI context generation endpoint.
type generateContextRequest struct {
	Group      string `json:"group"`
	Version    string `json:"version"`
	Kind       string `json:"kind"`
	SchemaJSON string `json:"schemaJSON"`
}

func (s *Server) GenerateCrdContextHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight requests
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set CORS header for the actual request
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		s.respondWithJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Only POST method is allowed"})
		return
	}

	var reqPayload generateContextRequest
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		s.log.Error("error decoding generate-context request body", "err", err)
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	generatedText, err := s.aiClient.GenerateCrdContext(
		r.Context(),
		reqPayload.Group,
		reqPayload.Version,
		reqPayload.Kind,
		reqPayload.SchemaJSON,
	)
	if err != nil {
		s.log.Error("error generating crd context from ollama", "err", err)
		http.Error(w, "Error communicating with AI service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If success, just return the content as text strings
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(generatedText))
}

// getClientForRequest returns the appropriate K8s client based on the X-Cluster-Name header.
// If no header is provided, it returns the current default client.
func (s *Server) getClientForRequest(r *http.Request) (*k8s.Client, error) {
	clusterName := r.Header.Get("X-Cluster-Name")
	if clusterName == "" {
		return s.ClusterManager.GetCurrentClient(), nil
	}
	return s.ClusterManager.GetClient(clusterName)
}

// ClustersHandler returns a list of all available clusters.
func (s *Server) ClustersHandler(w http.ResponseWriter, _ *http.Request) {
	clusters := s.ClusterManager.ListClusters()
	s.respondWithJSON(w, http.StatusOK, clusters)
}

func (s *Server) ClusterInfoHandler(w http.ResponseWriter, r *http.Request) {
	client, err := s.getClientForRequest(r)
	if err != nil {
		s.log.Error("cluster not found", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	clusterInfo, err := client.GetClusterInfo()
	if err != nil {
		s.log.Error("error getting cluster info", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.respondWithJSON(w, http.StatusOK, clusterInfo)
}

func (s *Server) CrdsHandler(w http.ResponseWriter, r *http.Request) {
	client, err := s.getClientForRequest(r)
	if err != nil {
		s.log.Error("cluster not found", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Note: This re-uses the k8s.GetCRDs which returns the TUI model.
	// For the API, we want the full spec, so we fetch the raw list and convert.
	crdList, err := client.ExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.log.Error("error listing CRDs", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	apiCrds := make([]models.APICRD, len(crdList.Items))
	var wg sync.WaitGroup
	for i, crd := range crdList.Items {
		wg.Add(1)
		go func(i int, crd apiextensionsv1.CustomResourceDefinition) {
			defer wg.Done()
			// This is a bit inefficient as it recounts, but for correctness with the new model.
			instanceCount := client.CountCRDInstances(context.Background(), crd)
			apiCrds[i] = models.ToAPICRD(crd, instanceCount)
		}(i, crd)
	}
	wg.Wait()

	s.respondWithJSON(w, http.StatusOK, apiCrds)
}

func (s *Server) CrsHandler(w http.ResponseWriter, r *http.Request) {
	client, err := s.getClientForRequest(r)
	if err != nil {
		s.log.Error("cluster not found", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	crdName := r.URL.Query().Get("crdName")
	if crdName == "" {
		s.log.Error("crd name is empty")
		http.Error(w, "crdName query parameter is required", http.StatusBadRequest)
		return
	}

	crs, err := client.GetCRsForCRD(context.Background(), crdName)
	if err != nil {
		s.log.Error("error getting crs from wizard api", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.respondWithJSON(w, http.StatusOK, crs)
}

func (s *Server) CrHandler(w http.ResponseWriter, r *http.Request) {
	client, err := s.getClientForRequest(r)
	if err != nil {
		s.log.Error("cluster not found", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	crdName := r.URL.Query().Get("crdName")
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if crdName == "" || namespace == "" || name == "" {
		s.log.Error("crd name or namespace or name is empty")
		http.Error(w, "crdName, namespace, and name query parameters are required", http.StatusBadRequest)
		return
	}

	cr, err := client.GetSingleCR(context.Background(), crdName, namespace, name)
	if err != nil {
		s.log.Error("error getting cr from wizard api", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s.respondWithJSON(w, http.StatusOK, cr)
}

func (s *Server) EventsHandler(w http.ResponseWriter, r *http.Request) {
	client, err := s.getClientForRequest(r)
	if err != nil {
		s.log.Error("cluster not found", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	crdName := r.URL.Query().Get("crdName")
	resourceUID := r.URL.Query().Get("resourceUid")

	if crdName == "" && resourceUID == "" {
		s.log.Error("crd name or resource uid is empty")
		http.Error(w, "Either crdName or resourceUid query parameter is required", http.StatusBadRequest)
		return
	}

	events, err := client.GetEvents(context.Background(), crdName, resourceUID)
	if err != nil {
		s.log.Error("error getting events from wizard api", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s.respondWithJSON(w, http.StatusOK, events)
}

func (s *Server) ResourceGraphHandler(w http.ResponseWriter, r *http.Request) {
	client, err := s.getClientForRequest(r)
	if err != nil {
		s.log.Error("cluster not found", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uid := r.URL.Query().Get("uid")
	if uid == "" {
		s.log.Error("uid is empty")
		http.Error(w, "uid query parameter is required", http.StatusBadRequest)
		return
	}

	graph, err := client.GetResourceGraph(context.Background(), uid)
	if err != nil {
		s.log.Error("error getting resource graph from wizard api", "uid", uid, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s.respondWithJSON(w, http.StatusOK, graph)
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	clusterCount := s.ClusterManager.ClusterCount()

	client, err := s.getClientForRequest(r)
	if err != nil {
		s.log.Error("cluster not found", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var status models.Health
	err = client.CheckHealth(r.Context())
	if err != nil {
		s.log.Error("health check failed", "err", err)
		status = models.Health{
			Status:       models.StatusUnhealthy.String(),
			ClusterCount: clusterCount,
			Message:      err.Error(),
		}
		s.respondWithJSON(w, http.StatusServiceUnavailable, status)
		return
	}

	status = models.Health{
		Status:       models.StatusHealthy.String(),
		ClusterCount: clusterCount,
	}
	s.respondWithJSON(w, http.StatusOK, status)
}

func (s *Server) respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			s.log.Error("Failed to encode JSON response", "err", err)
		}
	}
}

// ExportHandler handles the export of CRD documentation.
func (s *Server) ExportHandler(w http.ResponseWriter, r *http.Request) {
	client, err := s.getClientForRequest(r)
	if err != nil {
		s.log.Error("cluster not found", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	crdName := r.URL.Query().Get("crdName")
	format := r.URL.Query().Get("format")

	if crdName == "" {
		http.Error(w, "crdName query parameter is required", http.StatusBadRequest)
		return
	}
	if format == "" {
		format = "html"
	}

	s.log.Info("exporting CRD", "crd", crdName, "format", format, "cluster", client.ClusterName)

	crd, err := client.GetFullCRD(r.Context(), crdName)
	if err != nil {
		s.log.Error("failed to get CRD", "name", crdName, "err", err)
		http.Error(w, "Failed to get CRD: "+err.Error(), http.StatusInternalServerError)
		return
	}

	gen := generator.NewGenerator()
	apiCRD := models.ToAPICRD(*crd, 0)
	content, err := gen.Generate(apiCRD, format)
	if err != nil {
		s.log.Error("failed to generate documentation", "name", crdName, "err", err)
		http.Error(w, "Failed to generate documentation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers for download
	contentType := "text/html"
	if format == "markdown" || format == "md" {
		contentType = "text/markdown"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.%s\"", crdName, getExtension(format)))
	_, _ = w.Write(content)
}

// ExportAllHandler handles the batch export of all CRD documentation as a ZIP file.
func (s *Server) ExportAllHandler(w http.ResponseWriter, r *http.Request) {
	client, err := s.getClientForRequest(r)
	if err != nil {
		s.log.Error("cluster not found", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "html"
	}

	s.log.Info("exporting all CRDs", "format", format, "cluster", client.ClusterName)

	// List all CRDs
	crdList, err := client.ExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.log.Error("failed to list CRDs", "err", err)
		http.Error(w, "Failed to list CRDs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers for ZIP download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"crd_docs_%s.zip\"", time.Now().Format("20060102_150405")))

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Concurrency control
	concurrencyLimit := 5
	semaphore := make(chan struct{}, concurrencyLimit)
	var wg sync.WaitGroup

	// Mutex to synchronize zip writes (zip.Writer is not thread-safe)
	var zipMutex sync.Mutex

	gen := generator.NewGenerator()

	for _, crdItem := range crdList.Items {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire token

		go func(name string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release token

			// Fetch full CRD to ensure we have all details
			crd, err := client.GetFullCRD(r.Context(), name)
			if err != nil {
				s.log.Error("failed to get CRD", "name", name, "err", err)
				return // Skip this CRD on error
			}

			apiCRD := models.ToAPICRD(*crd, 0)
			content, err := gen.Generate(apiCRD, format)
			if err != nil {
				s.log.Error("failed to generate documentation", "name", name, "err", err)
				return
			}

			fileName := fmt.Sprintf("%s.%s", name, getExtension(format))

			// Write to ZIP safely
			zipMutex.Lock()
			defer zipMutex.Unlock()

			f, err := zipWriter.Create(fileName)
			if err != nil {
				s.log.Error("failed to create zip entry", "name", fileName, "err", err)
				return
			}
			if _, err := f.Write(content); err != nil {
				s.log.Error("failed to write zip entry content", "name", fileName, "err", err)
			}

		}(crdItem.Name)
	}

	wg.Wait()
}

// GenerateHandler handles the generation of documentation from uploaded content.
func (s *Server) GenerateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string `json:"content"`
		URL     string `json:"url"`
		Format  string `json:"format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	crdContent := []byte(req.Content)

	// If content is empty but URL is provided, fetch it
	if len(crdContent) == 0 && req.URL != "" {
		rawURL := giturl.ConvertGitURLToRaw(req.URL)
		s.log.Info("fetching CRD from URL", "original", req.URL, "raw", rawURL)

		resp, err := http.Get(rawURL) //nolint:gosec // user supplied url is intended
		if err != nil {
			s.log.Error("failed to fetch CRD from URL", "url", rawURL, "err", err)
			http.Error(w, "Failed to fetch CRD: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			s.log.Error("failed to fetch CRD from URL", "url", rawURL, "status", resp.Status)
			http.Error(w, "Failed to fetch CRD: "+resp.Status, http.StatusBadRequest)
			return
		}

		// Read limited amount to prevent abuse
		const maxFileSize = 10 * 1024 * 1024 // 10MB
		content, err := io.ReadAll(io.LimitReader(resp.Body, maxFileSize))
		if err != nil {
			s.log.Error("failed to read CRD content", "err", err)
			http.Error(w, "Failed to read CRD content", http.StatusInternalServerError)
			return
		}
		crdContent = content
	}

	if len(crdContent) == 0 {
		http.Error(w, "Content or URL is required", http.StatusBadRequest)
		return
	}

	// Try to unmarshal as CRD
	var crd apiextensionsv1.CustomResourceDefinition
	// Use yaml.Unmarshal as it handles both YAML and JSON
	if err := yaml.Unmarshal(crdContent, &crd); err != nil {
		http.Error(w, "Invalid CRD content: "+err.Error(), http.StatusBadRequest)
		return
	}

	gen := generator.NewGenerator()
	apiCRD := models.ToAPICRD(crd, 0)

	format := req.Format
	if format == "" {
		format = "html"
	}

	content, err := gen.Generate(apiCRD, format)
	if err != nil {
		s.log.Error("failed to generate documentation", "err", err)
		http.Error(w, "Failed to generate documentation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain") // Return content directly in body
	_, _ = w.Write(content)
}

func getExtension(format string) string {
	if format == "markdown" || format == "md" {
		return "md"
	}
	return "html"
}
