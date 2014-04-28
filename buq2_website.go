package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"runtime"
)

var templates = template.Must(template.ParseFiles(
	"templates/about.html",
	"templates/article.html",
	"templates/articles.html",
	"templates/articles_column.html",
	"templates/header.html",
	"templates/footer.html",
	"templates/analytics.html",
	"templates/tag.html",
	"templates/article_add_comment.html",
	"templates/article_comments.html",
	"templates/article_tags.html",
))

type SiteGlobal struct {
	UseGoogleAnalytics  bool
	GoogleAnalyticsCode string
	RecaptchaPublicKey  string
	RecaptchaPrivateKey string
	ContentRoot         string
	Scripts             template.HTML
	TitleBase           string
	Title               string
	Name                string
	Address             string
	Email               string
	// Keywords will be added to the header
	Keywords []string
}

var (
	siteGlobal = SiteGlobal{}
)

func readGlobalConfig() error {
	b, err := ioutil.ReadFile("site_config.json")
	if err != nil {
		return err
	}

	// Reasonable refault values
	siteGlobal.TitleBase = "Still Processing"
	siteGlobal.Name = "Matti Jukola"
	siteGlobal.Address = "http://buq2.com"
	siteGlobal.ContentRoot = "."

	err = json.Unmarshal(b, &siteGlobal)
	if err != nil {
		return err
	}
	return err
}

func websiteName() string {
	return siteGlobal.TitleBase
}

func websiteAddress() string {
	return siteGlobal.Address
}

func websiteDescription() string {
	return websiteAuthor() + "'s' blog about image processing"
}

func websiteAuthor() string {
	return siteGlobal.Name
}

func websiteEmail() string {
	return siteGlobal.Email
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	err := readGlobalConfig()
	if nil != err {
		fmt.Println("Failed to load site configurations: " + err.Error())
		return
	}

	err = readAuthConfig()
	if nil != err {
		fmt.Println("Failed to load auth configurations: " + err.Error())
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	http.HandleFunc("/", articlesHandler)
	http.HandleFunc("/articles/", articlesHandler)
	http.HandleFunc("/article/", articleHandler)
	http.HandleFunc("/about/", aboutHandler)
	http.HandleFunc("/atom.xml", atomHandler)
	http.HandleFunc("/rss", rssHandler)
	http.HandleFunc("/tag/", tagHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/oauth2callback", oauth2callbackHandler)
	http.Handle("/static/", fileserverHandler())
	http.ListenAndServe(":8080", nil)
}
