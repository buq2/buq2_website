package main

import (
	"fmt"
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

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println(getAllArticleIds())
	http.HandleFunc("/", articlesHandler)
	http.HandleFunc("/article/", articleHandler)
	http.Handle("/static/", fileserverHandler())
	http.ListenAndServe(":8080", nil)
}
