package main

import (
	"code.google.com/p/goauth2/oauth"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"
)

// http://golangtutorials.blogspot.fi/2011/11/oauth2-3-legged-authorization-in-go-web.html
// https://gist.github.com/border/3579615
// http://code.google.com/p/goauth2/source/browse/oauth/example/oauthreq.go
// https://github.com/kjk/web-blog/blob/master/go/handler_login.go

type Email struct {
	Value string
	Type  string
}

type AuthInformation struct {
	Emails []Email
	Id     string
}

type SiteCookie struct {
	UserEmail string
	UserId    string
}

func (cookie SiteCookie) IsAdmin() bool {
	// Checking for zero length ids just in case there is configuration error
	return cookie.UserEmail == config.AdminEmail &&
		cookie.UserId == config.AdminId &&
		len(cookie.UserId) > 0
}

var oauthCfg = &oauth.Config{
	ClientId:     "",
	ClientSecret: "",
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://accounts.google.com/o/oauth2/token",
	RedirectURL:  "http://127.0.0.1:8080/oauth2callback",

	// Scope 'openid' needs to be such that we get 'id' field
	// Use of 'email' only works almost, as it seems that second request
	// will always contain the 'id' field as well. By adding 'openid' we get
	// the 'id' on the first call
	Scope: "email openid",
}

// This is the URL that Google has defined so that an authenticated application
// may obtain the user's info in json format
const profileInfoURL = "https://www.googleapis.com/plus/v1/people/me"

var userInfoTemplate = template.Must(template.New("").Parse(`
<html>
<head>
<meta http-equiv="refresh" content="5;url=/">
</head>
<body>
You have logged in as: '{{.UserEmail}}'<br>
With id: '{{.UserId}}'<br>
{{if .IsAdmin}}
    You are an admin<br>
{{end}}
</body>
</html>
`))

func getCookie(r *http.Request) *SiteCookie {
	ret := new(SiteCookie)
	if cookie, err := r.Cookie(cookieName); err == nil {
		// detect a deleted cookie
		if "deleted" == cookie.Value {
			log.Printf("Cookie has been deleted")
			return new(SiteCookie)
		}
		val := make(map[string]string)
		if err = secureCookie.Decode(cookieName, cookie.Value, &val); err != nil {
			// Ignore error
			log.Printf("Could not decode cookie")
			return new(SiteCookie)
		}
		var ok bool
		if ret.UserEmail, ok = val["UserEmail"]; !ok {
			log.Printf("Error decoding cookie, no 'UserEmail' field")
			return new(SiteCookie)
		}
		if ret.UserId, ok = val["UserId"]; !ok {
			log.Printf("Error decoding cookie, no 'UserId' field")
			return new(SiteCookie)
		}
	}
	return ret
}

func createCookie(w http.ResponseWriter, info *AuthInformation) (*SiteCookie, error) {
	// Set cookie values to be encoded
	val := make(map[string]string)
	val["UserEmail"] = info.Emails[0].Value
	val["UserId"] = info.Id

	// Encode the data
	encoded, err := secureCookie.Encode(cookieName, val)
	if nil != err {

		return nil, err
	}

	// Create new cookie
	expiresIndays := 1
	http_cookie := &http.Cookie{
		Name:    cookieName,
		Value:   encoded,
		Path:    "/",
		Expires: time.Now().AddDate(0, 0, expiresIndays),
	}

	http.SetCookie(w, http_cookie)

	// Create returned SiteCookie
	cookie := new(SiteCookie)
	cookie.UserEmail = val["UserEmail"]
	cookie.UserId = val["UserId"]

	return cookie, nil
}

// Function that handles the callback from the Google server
func oauth2callbackHandler(w http.ResponseWriter, r *http.Request) {
	//Get the code from the response
	code := r.FormValue("code")

	// Create transport from config
	t := &oauth.Transport{Config: oauthCfg}

	// Exchange the received code for a token
	// Note that we do not store the actual token
	_, err := t.Exchange(code)
	if nil != err {
		log.Printf("Failed to exchange token")
		return
	}

	// Transport (t) has now valid token
	t.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}

	// Make the request to data
	req, err := t.Client().Get(profileInfoURL)
	if err != nil {
		log.Print("Request Error:", err)
		return
	}
	defer req.Body.Close()

	// Get the requested data
	body, _ := ioutil.ReadAll(req.Body)
	log.Print(string(body))

	// Unmarshal information from the request (emails + id)
	info := &AuthInformation{}
	err = json.Unmarshal(body, &info)
	if nil != err {
		log.Print("Marshaling error:", err)
		return
	}

	// Create login cookie for the users
	cookie, err := createCookie(w, info)
	if nil != err {
		log.Print("Failed to create cookie")
	}

	// Show user the login information
	userInfoTemplate.Execute(w, cookie)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Get the Google URL which shows the Authentication page to the user
	oauthCfg.ClientId, oauthCfg.ClientSecret = getGoogleOauthClientIdAndSecret()
	url := oauthCfg.AuthCodeURL("")

	// Redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Set value of the secure cookie to 'deleted' and redirect
	cookie := &http.Cookie{
		Name:    cookieName,
		Value:   "deleted",
		Expires: time.Now(),
		Path:    "/",
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusFound)
}
