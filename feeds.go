package main

import (
	"github.com/gorilla/feeds"
	"net/http"
	"time"
)

func getFeed() *feeds.Feed {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       websiteName(),
		Link:        &feeds.Link{Href: websiteAddress()},
		Description: websiteDescription(),
		Author:      &feeds.Author{websiteAuthor(), websiteEmail()},
		Created:     now,
	}

	articles := GetAllArticles()
	for _, article := range articles {
		item := &feeds.Item{
			Title:       article.Title,
			Link:        &feeds.Link{Href: article.Link},
			Description: article.Description,
			Author:      feed.Author,
			Created:     article.DateCreated.Time,
		}
		feed.Add(item)
	}

	return feed
}

func atomHandler(w http.ResponseWriter, r *http.Request) {
	feed := getFeed()
	atom, err := feed.ToAtom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(atom))
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	feed := getFeed()
	rss, err := feed.ToRss()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(rss))
}
