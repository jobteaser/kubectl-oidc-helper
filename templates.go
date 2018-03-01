package main

import (
	"html/template"
	"log"
	"net/http"
)

type tokenTmplData struct {
	IDToken        string
	RefreshToken   string
	RedirectURL    string
	Claims         Claim
	K8SAPI         string
	K8SName        string
	K8SClusterCert string
}

var tokenTmpl = template.Must(template.New("token.html").Parse(`<html>
  <head>
    <style>
pre {
 white-space: pre;
}
textarea {
	width: 100%;
	height: 500px;
}
		</style>
		<script>
function copyToClipboard(id) {
  var copyText = document.getElementById(id);
  copyText.select();
	document.execCommand("Copy");
	alert('copied')
}
</script>
  </head>
  <body>
	{{ if .RefreshToken }}
	<form action="{{ .RedirectURL }}" method="post">
	  <input type="hidden" name="refresh_token" value="{{ .RefreshToken }}" />
	  <input type="submit" value="Refresh" />
	</form>
	{{ end }}

	<div>
	<button onclick="copyToClipboard('helper')">Copy</button>

	<textarea readonly id="helper" >
kubectl config set-credentials '{{ .Claims.Email }}' \
--auth-provider=oidc \
--auth-provider-arg=idp-issuer-url={{ .Claims.Issuer }} \
--auth-provider-arg=client-id=kubernetes \
--auth-provider-arg=client-secret=some-secret \
{{- if .RefreshToken }}
--auth-provider-arg=refresh-token='{{ .RefreshToken }}' \
{{- end }}
--auth-provider-arg=id-token='{{ .IDToken }}'

tmp=$(mktemp)
echo "{{ .K8SClusterCert }}" | base64 --decode > $tmp

kubectl config set-cluster {{ .K8SName }} \
--server={{ .K8SAPI }} \
--certificate-authority=$tmp \
--embed-certs=true

kubectl config set-context {{ .K8SName }} --cluster={{ .K8SName }} --user='{{ .Claims.Email }}'

</textarea>

</div>
  </body>
</html>
`))

func renderToken(w http.ResponseWriter, redirectURL, idToken, refreshToken string, claims Claim, app *app) {
	renderTemplate(w, tokenTmpl, tokenTmplData{
		IDToken:        idToken,
		RefreshToken:   refreshToken,
		RedirectURL:    redirectURL,
		Claims:         claims,
		K8SAPI:         app.K8SAPI,
		K8SName:        app.K8SName,
		K8SClusterCert: app.K8SClusterCert,
	})
}

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	err := tmpl.Execute(w, data)
	if err == nil {
		return
	}

	switch err := err.(type) {
	case *template.Error:
		// An ExecError guarantees that Execute has not written to the underlying reader.
		log.Printf("Error rendering template %s: %s", tmpl.Name(), err)

		// TODO(ericchiang): replace with better internal server error.
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	default:
		// An error with the underlying write, such as the connection being
		// dropped. Ignore for now.
	}
}
