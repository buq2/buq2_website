package main

import (
	"net/http"
)

type Articles struct {
	SiteGlobal
	ArticlesLeft  []*Article
	ArticlesRight []*Article
}

func articlesHandler(w http.ResponseWriter, r *http.Request) {
	articles_raw := GetAllArticles()
	articles := Articles{}
	articles.SiteGlobal = siteGlobal

	// Every other goes to left column, every other to right column
	for idx, article := range articles_raw {
		if idx%2 == 0 {
			articles.ArticlesLeft = append(articles.ArticlesLeft, article)
		} else {
			articles.ArticlesRight = append(articles.ArticlesRight, article)
		}
	}

	renderTemplate(w, "articles", articles)
}
