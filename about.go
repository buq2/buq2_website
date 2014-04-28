package main

import (
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

type About struct {
	SiteGlobal
	Body  template.HTML
	Title string
}

func getAbout() (*About, error) {
	about_data, err := ioutil.ReadFile(siteGlobal.ContentRoot + "/about/about.md")
	if err != nil {
		return nil, err
	}

	// Convert markdown to thml
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_USE_XHTML
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	// Can't use HTML_SANITIZE_OUTPUT as it will remove custom
	// classes from code blocks, etc
	//htmlFlags |= blackfriday.HTML_SANITIZE_OUTPUT
	// Can't use GitHub style as highlight.js does not support it
	//htmlFlags |= blackfriday.HTML_GITHUB_BLOCKCODE
	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	extensions := 0
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	extensions |= blackfriday.EXTENSION_HEADER_IDS

	body_markdown := blackfriday.Markdown(about_data, renderer, extensions)

	about := new(About)
	about.SiteGlobal = siteGlobal
	about.Body = template.HTML(body_markdown)
	about.Title = "About"

	return about, nil
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	about, err := getAbout()

	if nil != err {
		log.Print("Could not parse about page: " + err.Error())
		http.NotFound(w, r)
		return
	}

	renderTemplate(w, "about", about)
}
