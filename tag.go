package main

import (
	"errors"
	"log"
	"net/http"
	"regexp"
)

type TagData struct {
	SiteGlobal
	Articles Articles
	Tag      string
}

var validTag = regexp.MustCompile("^/(tag)/([a-zA-Z0-9_ ]+)$")

func getTag(r *http.Request) (string, error) {
	m := validTag.FindStringSubmatch(r.URL.Path)
	if m == nil {
		return "", errors.New("Invalid tag from request: " + r.URL.Path)
	}

	return m[2], nil // The tag is the second subexpression.
}

func tagHandler(w http.ResponseWriter, r *http.Request) {
	tag, err := getTag(r)
	if err != nil {
		http.NotFound(w, r)
		log.Print("Could not parse tag from request:" + err.Error())
		return
	}
	articles, err := GetArticlesByTag(tag)
	if err != nil {
		http.NotFound(w, r)
		log.Print("Failed to get articles by tag:" + tag)
		return
	}

	data := TagData{}
	data.Tag = tag
	data.SiteGlobal = siteGlobal
	data.Articles = SplitRawArticlesIntoColumns(articles)

	renderTemplate(w, "tag", data)
}
