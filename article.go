package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"regexp"
	"strings"
)

var validArticle = regexp.MustCompile("^/(article)/([a-zA-Z0-9]+)$")

type PageMetaData struct {
	Title            string
	ShortDescription string
	CreateDate       string
	ModifiedDate     string
}

type Article struct {
	Body             template.HTML
	Title            string
	ShortDescription string
	CreateDate       string
	ModifiedDate     string
	Id               string
}

func GetAllArticles() []*Article {
	ids := getAllArticleIds()
	articles := make([]*Article, len(ids))

	for idx, id := range ids {
		articles[idx], _ = NewArticle(id)
	}

	return articles
}

func NewArticle(id string) (*Article, error) {
	// Try to find the data to the article with certain id
	filename := id + ".txt"
	article_data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	article := new(Article)
	err = parseRawTextArticleData(article_data, article)
	article.Id = id

	return article, err
}

func getAllArticleIds() []string {
	folder := "./"
	ids := []string{}

	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return ids
	}

	for _, file := range files {
		filename := file.Name()
		ext := path.Ext(filename)
		if ext == ".txt" {
			name := filename[0 : len(filename)-len(ext)]
			ids = append(ids, name)
		}
	}

	return ids
}

func parseArticleBodyToHtml(article_body_data []byte) template.HTML {
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

	body_markdown := blackfriday.Markdown(article_body_data, renderer, extensions)

	return template.HTML(body_markdown)
}

func parseRawTextArticleData(article_data []byte, article *Article) error {
	// Find meta and body separator
	article_header_separator := "---------- META END ----------"
	separator_len := len(article_header_separator)
	separator_begin := strings.Index(string(article_data), article_header_separator)

	// Separate meta and body data
	article_body_data := []byte{}
	article_meta_data := []byte{}

	if separator_begin > 0 {
		// Found separator, split the data into meta data and body data
		article_meta_data = article_data[:separator_begin-1]
		article_body_data = article_data[separator_begin+separator_len:]
	} else {
		// Did not find separator, meta data is empty
		article_body_data = article_data
	}

	// Read metadata
	meta := PageMetaData{}
	if err := json.Unmarshal(article_meta_data, &meta); err != nil {
		fmt.Println("Failed to parse article meta data. Returning empty meta data")
	}

	// Put metadata into the Article
	article.Title = meta.Title
	article.ShortDescription = meta.ShortDescription
	article.CreateDate = meta.CreateDate
	article.ModifiedDate = meta.ModifiedDate

	// Parse article body to valid HTML (which is safe)
	article.Body = parseArticleBodyToHtml(article_body_data)

	return nil
}

func getArticleId(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validArticle.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The id is the second subexpression.
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getArticleId(w, r)
	if err != nil {
		return
	}
	article, err := NewArticle(title)
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "article", *article)
}
