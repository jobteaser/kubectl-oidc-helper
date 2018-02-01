package main

import (
	"html/template"
	"log"
	"net/http"
)

type tokenTmplData struct {
	IDToken      string
	RefreshToken string
	RedirectURL  string
	Claims       Claim
	K8SAPI       string
	K8SName      string
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
echo "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMyRENDQWNDZ0F3SUJBZ0lSQUlISkpDVFJzdCtQcGxweWkzK1lyL2t3RFFZSktvWklodmNOQVFFTEJRQXcKRlRFVE1CRUdBMVVFQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4TnpFeU1ETXhOak15TVRCYUZ3MHlOekV5TURNeApOak15TVRCYU1CVXhFekFSQmdOVkJBTVRDbXQxWW1WeWJtVjBaWE13Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBCkE0SUJEd0F3Z2dFS0FvSUJBUURkNFQ2TkJPUkpDTjdhNUMwVmczOC85NG1nTzZWZjRLNzUxN0t0SWlUSjhHODYKWXRONnB0MnlEaWR6VVVDWjhDVU1SSWYwRFlWY2RHaXBFWkQzclJiNmFPSmtqbElCVmN4aUJ6VGVVcytiOXNpbwpVNUxiV0ZrNGtzZ3d5UGtjeDBwU2ZaY2M5SStDcVZrR0lma2dPRXJobnB2KzV6UEtyVi9UODE0TkpOZWpaRWRUCmpIeXF0RllSU05icTJ1STVobStVQ2pNQUtrWm9CNTF1LzdvQTcwcDFUZlgrTmRMODRtOFNUTUVVT09sWEZXQ2sKTXQxS1BNOWE1RFRSRnc2SSs3RlY1U0FPek1IZ3U4UkVGZkUrSVNYMS90N0tSU1hRbGJGYkQ2UjB0ODlWeHlyOApuN0dEMW1KUUdvbExQaTMxQ3JvNURha2FET09MWXQvbkVKTzFoSDNaQWdNQkFBR2pJekFoTUE0R0ExVWREd0VCCi93UUVBd0lCQmpBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFBRXQwN3gKajZHZWNtUkVLS0lmQ3VOUXJUcnMvUlVuNjVRbHhTR2NLUnJJd2UvRDY3WVFjVzNLRGlyeVYzYlp1VkdUaU9NcgpIdW1KVG01NmYzNlRaK0E3MDRzUzlMNDR1aXZXNysyWXllOGEyNmF2cjcwMDlLTFpXTUVvVGlBQTJ5Q2hiTythCjY1cmxlZ2trbGFDVGVnNm85bE52M1ZhNHNzNkY2cm9VUm1GaURTcVR5THczcnArVjdqc2tNNXVpVmNSNDlPM3YKK3JIM00wU1VOZ2xxSzliM2ZnZlBjMGJrVlpFbW1pdzFNNk1kdGNIUXVKWXlYRVAxUmE0bkYrV1FQVkJ5RUpVdwpoRjB0TFVLZnRXb1Y4QnZ2U3AzS2RrUDVLazl4Q2xxUGNpVjNTSmVyaUJxamMra2k4cHBNdnk0K3lIeFl1L1NjCkcxL3VOWEh2RDhMMjJpc3gKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=" | base64 --decode > $tmp

kubectl config set-cluster {{ .K8SName }} \
--server={{ .K8SAPI }} \
--certificate-authority=$tmp \
--embed-certs=true

kubectl config set-context {{ .K8SName }} --cluster={{ .K8SName }} --user='{{ .Claims.Email }}' --namespace=jobteaser
kubectl config use-context {{ .K8SName }}

</textarea>

</div>
  </body>
</html>
`))

func renderToken(w http.ResponseWriter, redirectURL, idToken, refreshToken string, claims Claim, app *app) {
	renderTemplate(w, tokenTmpl, tokenTmplData{
		IDToken:      idToken,
		RefreshToken: refreshToken,
		RedirectURL:  redirectURL,
		Claims:       claims,
		K8SAPI:       app.K8SAPI,
		K8SName:      app.K8SName,
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

		http.Error(w, "Internal server error", http.StatusInternalServerError)
	default:
		// An error with the underlying write, such as the connection being
		// dropped. Ignore for now.
	}
}
