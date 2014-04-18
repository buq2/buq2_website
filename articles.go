package main

import (
	"net/http"
)

type Articles struct {
	Articles []*Article
}

func articlesHandler(w http.ResponseWriter, r *http.Request) {
	articles_raw := GetAllArticles()
	articles := Articles{articles_raw}
	renderTemplate(w, "articles", articles)
}
