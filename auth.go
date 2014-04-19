package main

import (
	"code.google.com/p/goauth2/oauth"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"
)

// http://golangtutorials.blogspot.fi/2011/11/oauth2-3-legged-authorization-in-go-web.html
// https://gist.github.com/border/3579615
// http://code.google.com/p/goauth2/source/browse/oauth/example/oauthreq.go

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
	oauthCfg.ClientId, oauthCfg.ClientSecret = getGoogleOauthClientIdAndSecret()
	url := oauthCfg.AuthCodeURL("")

	//redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)
}

func isAdmin(cookie *SiteCookie) bool {
	return cookie.UserEmail == config.AdminEmail && cookie.UserId == config.AdminId
}

func getCookie(r *http.Request) *SiteCookie {
	ret := new(SiteCookie)
	if cookie, err := r.Cookie(cookieName); err == nil {
		// detect a deleted cookie
		if "deleted" == cookie.Value {
			return new(SiteCookie)
		}
		val := make(map[string]string)
		if err = secureCookie.Decode(cookieName, cookie.Value, &val); err != nil {
			// Ignore error
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

func checkCredentials(w http.ResponseWriter, r *http.Request, info *AuthInformation) error {
	if len(info.Emails) != 1 ||
		info.Emails[0].Value != config.AdminEmail ||
		info.Id != config.AdminId {

		return errors.New("Not an admin")
	}

	// Set cookie values to be encoded
	val := make(map[string]string)
	val["UserEmail"] = info.Emails[0].Value
	val["UserId"] = info.Id

	// Encode the data
	encoded, err := secureCookie.Encode(cookieName, val)
	if nil != err {
		return err
	}

	// Create new cookie
	expiresIndays := 1
	cookie := &http.Cookie{
		Name:    cookieName,
		Value:   encoded,
		Path:    "/",
		Expires: time.Now().AddDate(0, 0, expiresIndays),
	}

	http.SetCookie(w, cookie)

	return nil
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

	// Unmarsharl information from the request (emails + id)
	info := &AuthInformation{}
	err = json.Unmarshal(body, &info)
	if nil != err {
		log.Print("Marshaling error:", err)
		return
	}

	checkCredentials(w, r, info)

	userInfoTemplate.Execute(w, string(body))
}
