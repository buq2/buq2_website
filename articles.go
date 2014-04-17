package main

import (
	"html/template"
	"net/http"
)

var articles_templates = template.Must(template.ParseFiles("articles.html"))

type Articles struct {
	Articles []*Article
}

func renderArticlesTemplate(w http.ResponseWriter, tmpl string, articles *Articles) {
	err := articles_templates.ExecuteTemplate(w, tmpl+".html", articles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func articlesHandler(w http.ResponseWriter, r *http.Request) {
	articles_raw := GetAllArticles()
	articles := Articles{articles_raw}
	renderArticlesTemplate(w, "articles", &articles)
}
