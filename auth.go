package main

import (
	"code.google.com/p/goauth2/oauth"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
)

// http://golangtutorials.blogspot.fi/2011/11/oauth2-3-legged-authorization-in-go-web.html
// https://gist.github.com/border/3579615
// http://code.google.com/p/goauth2/source/browse/oauth/example/oauthreq.go

var oauthCfg = &oauth.Config{
	ClientId:     "",
	ClientSecret: "",
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://accounts.google.com/o/oauth2/token",
	RedirectURL:  "http://127.0.0.1:8080/oauth2callback",
	Scope:        "email",
}

//This is the URL that Google has defined so that an authenticated application may obtain the user's info in json format
const profileInfoURL = "https://www.googleapis.com/plus/v1/people/me"

var userInfoTemplate = template.Must(template.New("").Parse(`
<html><body>
{{.}}
</body></html>
`))

func authHandler(w http.ResponseWriter, r *http.Request) {
	//Get the Google URL which shows the Authentication page to the user
	url := oauthCfg.AuthCodeURL("")

	//redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)
}

// Function that handles the callback from the Google server
func oauth2callbackHandler(w http.ResponseWriter, r *http.Request) {
	//Get the code from the response
	code := r.FormValue("code")

	t := &oauth.Transport{Config: oauthCfg}

	// Exchange the received code for a token
	tok, _ := t.Exchange(code)

	{
		tokenCache := oauth.CacheFile("./request.token")

		err := tokenCache.PutToken(tok)
		if err != nil {
			log.Print("Cache write:", err)
		}
		log.Printf("Token is cached in %v\n", tokenCache)
	}

	t.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}

	// Make the request.
	req, err := t.Client().Get(profileInfoURL)
	if err != nil {
		log.Print("Request Error:", err)
		return
	}
	defer req.Body.Close()

	body, _ := ioutil.ReadAll(req.Body)

	log.Println(string(body))
	userInfoTemplate.Execute(w, string(body))
}
