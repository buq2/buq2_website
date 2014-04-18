package main

import (
	"html/template"
	"net/http"
	"runtime"
)

var templates = template.Must(template.ParseFiles(
	"templates/article.html",
	"templates/articles.html",
	"templates/articles_column.html",
	"templates/header.html",
	"templates/footer.html",
))

func websiteName() string {
	return "Still Processing"
}

func websiteAddress() string {
	return "http://buq2.com"
}

func websiteDescription() string {
	return websiteAuthor() + "'s' blog about image processing"
}

func websiteAuthor() string {
	return "Matti Jukola"
}

func websiteEmail() string {
	return "spam@buq2.com"
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	http.HandleFunc("/", articlesHandler)
	http.HandleFunc("/article/", articleHandler)
	http.HandleFunc("/atom.xml", atomHandler)
	http.HandleFunc("/rss", rssHandler)
	http.HandleFunc("/login", authHandler)
	http.HandleFunc("/oauth2callback", oauth2callbackHandler)
	http.Handle("/static/", fileserverHandler())
	http.ListenAndServe(":8080", nil)
}
