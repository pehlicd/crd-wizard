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
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"sync"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pehlicd/crd-wizard/internal/logger"
	"github.com/pehlicd/crd-wizard/internal/models"

	"github.com/pehlicd/crd-wizard/internal/k8s"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	K8sClient *k8s.Client
	router    *http.ServeMux
	server    *http.Server
	log       *logger.Logger
}

func NewServer(client *k8s.Client, port string, log *logger.Logger) *Server {
	r := http.NewServeMux()
	s := &Server{
		K8sClient: client,
		router:    r,
		server: &http.Server{
			Addr:         ":" + port,
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
		log: log,
	}
	s.registerHandlers()
	return s
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) registerHandlers() {
	apiRouter := s.router
	apiRouter.HandleFunc("/cluster-info", s.ClusterInfoHandler)
	apiRouter.HandleFunc("/crds", s.CrdsHandler)
	apiRouter.HandleFunc("/crs", s.CrsHandler)
	apiRouter.HandleFunc("/cr", s.CrHandler)
	apiRouter.HandleFunc("/events", s.EventsHandler)
	apiRouter.HandleFunc("/resource-graph", s.ResourceGraphHandler)
	s.router.Handle("/api/", http.StripPrefix("/api", s.log.Middleware(apiRouter)))

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

func (s *Server) ClusterInfoHandler(w http.ResponseWriter, _ *http.Request) {
	clusterName, err := s.K8sClient.GetClusterName()
	if err != nil {
		s.log.Error("error getting cluster name", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]string{"clusterName": clusterName})
}

func (s *Server) CrdsHandler(w http.ResponseWriter, _ *http.Request) {
	// Note: This re-uses the k8s.GetCRDs which returns the TUI model.
	// For the API, we want the full spec, so we fetch the raw list and convert.
	crdList, err := s.K8sClient.ExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})
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
			instanceCount := s.K8sClient.CountCRDInstances(context.Background(), crd)
			apiCrds[i] = models.ToAPICRD(crd, instanceCount)
		}(i, crd)
	}
	wg.Wait()

	s.respondWithJSON(w, http.StatusOK, apiCrds)
}

func (s *Server) CrsHandler(w http.ResponseWriter, r *http.Request) {
	crdName := r.URL.Query().Get("crdName")
	if crdName == "" {
		s.log.Error("crd name is empty")
		http.Error(w, "crdName query parameter is required", http.StatusBadRequest)
		return
	}

	crs, err := s.K8sClient.GetCRsForCRD(context.Background(), crdName)
	if err != nil {
		s.log.Error("error getting crs from wizard api", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.respondWithJSON(w, http.StatusOK, crs)
}

func (s *Server) CrHandler(w http.ResponseWriter, r *http.Request) {
	crdName := r.URL.Query().Get("crdName")
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if crdName == "" || namespace == "" || name == "" {
		s.log.Error("crd name or namespace or name is empty")
		http.Error(w, "crdName, namespace, and name query parameters are required", http.StatusBadRequest)
		return
	}

	cr, err := s.K8sClient.GetSingleCR(context.Background(), crdName, namespace, name)
	if err != nil {
		s.log.Error("error getting cr from wizard api", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s.respondWithJSON(w, http.StatusOK, cr)
}

func (s *Server) EventsHandler(w http.ResponseWriter, r *http.Request) {
	crdName := r.URL.Query().Get("crdName")
	resourceUID := r.URL.Query().Get("resourceUid")

	if crdName == "" && resourceUID == "" {
		s.log.Error("crd name or resource uid is empty")
		http.Error(w, "Either crdName or resourceUid query parameter is required", http.StatusBadRequest)
		return
	}

	events, err := s.K8sClient.GetEvents(context.Background(), crdName, resourceUID)
	if err != nil {
		s.log.Error("error getting events from wizard api", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s.respondWithJSON(w, http.StatusOK, events)
}

func (s *Server) ResourceGraphHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		s.log.Error("uid is empty")
		http.Error(w, "uid query parameter is required", http.StatusBadRequest)
		return
	}

	graph, err := s.K8sClient.GetResourceGraph(context.Background(), uid)
	if err != nil {
		s.log.Error("error getting resource graph from wizard api", "uid", uid, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s.respondWithJSON(w, http.StatusOK, graph)
}

func (s *Server) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			s.log.Error("Failed to encode JSON response", "err", err)
		}
	}
}
