/*
 * Copyright 2023 Marco Confalonieri.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package server

import (
	"net"
	"net/http"

	"github.com/bsm/openmetrics"
	"github.com/bsm/openmetrics/omhttp"
	log "github.com/sirupsen/logrus"
)

// PublicServer is the liveness and readiness server.
type PublicServer struct {
	status *HealthStatus
	reg    *openmetrics.Registry
}

func NewPublicServer(status *HealthStatus, reg *openmetrics.Registry) *PublicServer {
	return &PublicServer{
		status: status,
		reg:    reg,
	}
}

// livenessHandler checks if the server is healthy. It writes 200/OK if the
// healthy flag is set to "true" and 503/Service Unavailable otherwise.
func (s PublicServer) livenessHandler(w http.ResponseWriter, r *http.Request) {
	healthy := s.status.IsHealthy()
	var err error
	if healthy {
		_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err = w.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
	}
	if err != nil {
		log.Warn("Could not answer to a liveness probe: ", err.Error())
	}
}

// readinessHandler checks if the server is ready. It writes 200/OK if the
// healthy flag is set to "true" and 503/Service Unavailable otherwise.
func (s PublicServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
	ready := s.status.IsReady()
	var err error
	if ready {
		_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err = w.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
	}
	if err != nil {
		log.Warn("Could not answer to a readiness probe: ", err.Error())
	}
}

// Start starts the liveness and readiness server.
func (s *PublicServer) Start(startedChan chan struct{}, options ServerOptions) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.readinessHandler)
	mux.HandleFunc("/ready", s.readinessHandler)
	mux.HandleFunc("/health", s.livenessHandler)
	mux.Handle("/metrics", omhttp.NewHandler(s.reg))

	address := options.GetHealthAddress()

	srv := &http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  options.GetReadTimeout(),
		WriteTimeout: options.GetWriteTimeout(),
	}

	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	if startedChan != nil {
		startedChan <- struct{}{}
	}

	if err := srv.Serve(l); err != nil {
		log.Fatal(err)
	}
}