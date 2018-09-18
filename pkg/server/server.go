/*
Copyright 2018 Jobteaser.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var app App

// Start start the relay party http server.
func Start(a App) {
	app = a
	app.AuthClient.RedirectURL = "http://" + app.Bind + ":" + app.Port + "/callback" // TODO: use params

	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/callback", callbackHandler).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", app.Bind, app.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Infof("Starting kubectl oidc helper on http://%s:%s", app.Bind, app.Port)
	log.Fatal(srv.ListenAndServe())
}
