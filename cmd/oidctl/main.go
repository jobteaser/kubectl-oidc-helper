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

package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/jobteaser/kubectl-oidc-helper/pkg/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = cobra.Command{
	Use:   "oidctl",
	Short: "OpenID Connect client that simplifies authentication to a Kubernetes cluster.",
}

var app = server.App{}

var scopes string

var serverCmd = cobra.Command{
	Use:   "server",
	Short: "Start OpenID Connect Relay Party.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := oidc.ClientContext(context.Background(), http.DefaultClient)
		log.Info("Reading OpenID Connect Provider metadata")
		provider, err := oidc.NewProvider(ctx, app.Issuer)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		app.Provider = *provider
		app.AuthClient.Endpoint = app.Provider.Endpoint()
		app.AuthClient.Scopes = strings.Split(scopes, ",")

		server.Start(app)
	},
}

func init() {
	rootCmd.AddCommand(&serverCmd)

	serverCmd.Flags().StringVarP(&app.Bind, "bind", "b", "localhost", "Binds OIDC to the specified IP")

	serverCmd.Flags().StringVarP(&app.AuthClient.ClientID, "client-id", "", "", "OpenID Connect client ID (required)")
	serverCmd.MarkFlagRequired("client-id")

	serverCmd.Flags().StringVarP(&app.AuthClient.ClientSecret, "client-secret", "", "", "OpenID Connect client secret (required)")
	serverCmd.MarkFlagRequired("client-secret")

	serverCmd.Flags().StringVarP(&app.ClusterEndpoint, "cluster-endpoint", "", "", "Kubernetes Cluster API endpoint (required)")
	serverCmd.MarkFlagRequired("cluster-endpoint")

	serverCmd.Flags().StringVarP(&app.ClusterName, "cluster-name", "", "", "Kubernetes Cluster configuration name (required)")
	serverCmd.MarkFlagRequired("cluster-name")

	serverCmd.Flags().StringVarP(&app.Issuer, "issuer", "", "", "OpenID Connect issuer (required)")
	serverCmd.MarkFlagRequired("issuer")

	serverCmd.Flags().StringVarP(&app.Port, "port", "p", "3000", "Run OIDC on the specified port")

	serverCmd.Flags().StringVarP(&scopes, "scopes", "", "openid", "OpenID Connect requested scopes")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
