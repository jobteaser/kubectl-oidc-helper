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
	"html/template"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

var homeTmpl = template.Must(template.New("homeTmpl").Parse(`<!doctype html>
<html>
  <head>
    <title>kubectl oidc helper</title>
    <meta name="robots" content="noindex, nofollow, noarchive, noodp, nosnippet">
    <meta charset="utf-8">
  </head>
  <body>
    <a href="{{ . }}" hreflang="en" rel="external" type="text/html">Connect</a>
  </body>
</html>
`))

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html")

	state, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{"exp": time.Now().UTC().Add(time.Minute * 30).Unix()},
	).SignedString(app.GetJWTSecret())

	if err != nil {
		log.Error(err)
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}

	authURL := app.AuthClient.AuthCodeURL(state)

	if err := homeTmpl.Execute(w, authURL); err != nil {
		log.Error(err)
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}
}
