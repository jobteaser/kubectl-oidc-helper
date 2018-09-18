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
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
	log "github.com/sirupsen/logrus"
)

var callbackTmpl = template.Must(template.New("callbackTmpl").Parse(`<!doctype html>
<html>
  <head>
    <title>Kubernetes OIDC Portal</title>
    <meta name="robots" content="noindex, nofollow, noarchive, noodp, nosnippet">
    <meta charset="utf-8">
  </head>
  <body>
		<div>
			<h2>Configure the Kubernetes cluster crendentials</h2>
<pre>
kubectl config set-credentials "{{ .Email }}" \
	--auth-provider=oidc \
	--auth-provider-arg=idp-issuer-url={{ .Issuer }} \
	--auth-provider-arg=client-id={{ .ClientID }} \
	--auth-provider-arg=client-secret={{ .ClientSecret }} \
	{{- if .RefreshToken }}
	--auth-provider-arg=refresh-token={{ .RefreshToken }} \
	{{- end }}
	--auth-provider-arg=id-token={{ .IDToken }}
</pre>
		</div>
		{{ if .ClusterCrt }}
		<div>
			<h2>Configure the Kubernetes cluster cluster</h2>
<pre>
CRT=$(mktemp)
cat {{ "<<EOF" }} >> $CRT
{{ .ClusterCrt }}EOF

kubectl config set-cluster {{ .ClusterName }} \
	--server={{ .ClusterServer }} \
	--certificate-authority=$CRT \
	--embed-certs=true
</pre>
		</div>
		{{ end }}
		<div>
			<h2>Change your current context</h2>
<pre>
kubectl config set-context {{ .ClusterName }} --cluster={{ .ClusterName }} --user={{ .Email }}
</pre>
		</div>
  </body>
</html>
`))

type callbackView struct {
	Email         string
	Issuer        string
	ClientID      string
	ClientSecret  string
	RefreshToken  string
	IDToken       string
	ClusterName   string
	ClusterServer string
	ClusterCrt    template.HTML
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	view := callbackView{}
	view.ClusterCrt = template.HTML(strings.Replace(app.GetClusterCertificate(), "\n", "<br>", -1))
	view.ClusterServer = app.ClusterEndpoint
	view.ClientID = app.AuthClient.ClientID
	view.ClientSecret = app.AuthClient.ClientSecret
	view.ClusterName = app.ClusterName
	view.Issuer = app.Issuer

	w.Header().Add("Content-Type:", "text/html")

	if err := r.FormValue("error"); err != "" {
		http.Error(w, fmt.Sprintf("%s: %s", err, r.FormValue("error_description")), http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	if err := app.VerifyJWT(r.Form.Get("state")); err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	ctx := oidc.ClientContext(context.Background(), http.DefaultClient)
	verifier := app.Provider.Verifier(&oidc.Config{ClientID: app.AuthClient.ClientID})

	oauth2token, err := app.AuthClient.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		log.Error(err)
		http.Error(w, "invalid_request", http.StatusUnauthorized)
		return
	}
	view.RefreshToken = oauth2token.RefreshToken

	rawIDToken, ok := oauth2token.Extra("id_token").(string)
	if !ok {
		log.Error("missing IDToken")
		http.Error(w, "invalid_request", http.StatusUnauthorized)
		return
	}
	view.IDToken = rawIDToken

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Error("missing IDToken")
		http.Error(w, "invalid_request", http.StatusUnauthorized)
		return
	}

	var claims struct {
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
	}

	if err := idToken.Claims(&claims); err != nil {
		log.Error(err)
		http.Error(w, "invalid_request", http.StatusUnauthorized)
		return
	}

	view.Email = claims.Email

	if err := callbackTmpl.Execute(w, view); err != nil {
		log.Debug(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
