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
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// App contain the application state.
type App struct {
	Port            string
	Bind            string
	ClusterEndpoint string
	ClusterName     string
	AuthClient      oauth2.Config
	Provider        oidc.Provider
	Issuer          string
}

// GetClusterCertificate returns the Kubernetes Cluster certificate.
func (app *App) GetClusterCertificate() string {
	cert := new(bytes.Buffer)

	defaultTr := http.DefaultTransport.(*http.Transport)
	tr := &http.Transport{
		Proxy:                 defaultTr.Proxy,
		DialContext:           defaultTr.DialContext,
		MaxIdleConns:          defaultTr.MaxIdleConns,
		IdleConnTimeout:       defaultTr.IdleConnTimeout,
		ExpectContinueTimeout: defaultTr.ExpectContinueTimeout,
		TLSHandshakeTimeout:   defaultTr.TLSHandshakeTimeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	res, err := client.Get(app.ClusterEndpoint)
	if err != nil {
		log.Error(err)
		return ""
	}
	data := res.TLS.PeerCertificates[0].Raw
	block := &pem.Block{Type: "CERTIFICATE", Bytes: data}

	if err := pem.Encode(cert, block); err != nil {
		log.Error(err)
		return ""
	}

	return cert.String()
}

// GetJWTSecret rerturn the JWT HMAC secret.
func (app *App) GetJWTSecret() []byte {
	return []byte(fmt.Sprintf("%s+%s", app.AuthClient.ClientID, app.AuthClient.ClientSecret))
}

// VerifyJWT validate the JSON Web Token with RFC defined claims.
func (app *App) VerifyJWT(str string) error {
	token, err := jwt.Parse(str, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return app.GetJWTSecret(), nil
	})

	switch err.(type) {
	case nil:
		if !token.Valid {
			return errors.New("invalid_jwt")
		}
	default:
		log.Error(err)
		return errors.New("invalid_jwt")
	}
	return nil
}
