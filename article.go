package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
)

type PageMetaData struct {
	Title        string
	LongTitle    string
	Description  string
	DateCreated  string
	DateModified string
	Icon         string
}

type Article struct {
	SiteGlobal
	Scripts      template.HTML
	Body         template.HTML
	Title        string
	LongTitle    string
	Description  string
	DateCreated  time.Time
	DateModified time.Time
	Id           string
	Icon         string
	Link         string
}

var validArticle = regexp.MustCompile("^/(article)/([a-zA-Z0-9]+)$")

var articleScripts = template.HTML(`
<!--***********
 Mathjax
************-->
<script type="text/x-mathjax-config">
    MathJax.Hub.Config({"HTML-CSS": {
        preferredFont: "TeX", availableFonts: ["STIX","TeX"], linebreaks: { automatic:true }},
        tex2jax: { inlineMath: [ ["$", "$"], ["\\\\(","\\\\)"] ], displayMath: [ ["$$","$$"], ["\\[", "\\]"] ], processEscapes: true, ignoreClass: "tex2jax_ignore|dno" },
        TeX: {
                equationNumbers: { autoNumber: "AMS" },
                noUndefined: { attributes: { mathcolor: "red", mathbackground: "#FFEEEE", mathsize: "90%" } }
            },
        messageStyle: "none"
        });
</script>
<script type="text/javascript" src="http://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS_HTML"></script>
<!--***********
 Highlighting
************-->
<link rel="stylesheet" href="//cdnjs.cloudflare.com/ajax/libs/highlight.js/8.0/styles/default.min.css">
<script src="//cdnjs.cloudflare.com/ajax/libs/highlight.js/8.0/highlight.min.js"></script>
<!--***********
 jquery
************-->
<script src="http://cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
<!--***********
 Image captions
************-->
<script>
    function InitAutoImageCaption() {
        // Every image referenced from a Markdown document
        $("img").each(function() {
            // Let's put a caption if there is not one
            if($(this).attr("alt"))
                $(this).wrap('<figure class="image"></figure>')
                    .after('<figcaption>'+$(this).attr("alt")+'</figcaption>');
        });
    }
</script>
<!--***********
 Initialization code
************-->
<script>
    $(document).ready(function(){
        InitAutoImageCaption();
        hljs.initHighlightingOnLoad();
    });
</script>
`)

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

	// Other
	article.SiteGlobal = siteGlobal
	article.Id = id
	article.Scripts = articleScripts
	article.Link = websiteAddress() + "/article/" + article.Id

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
	err := json.Unmarshal(article_meta_data, &meta)
	if err != nil {
		fmt.Println("Failed to parse article meta data. Returning empty meta data: " + err.Error())
	}

	// Put metadata into the Article
	article.Title = meta.Title
	article.Description = meta.Description
	article.LongTitle = meta.LongTitle
	article.DateCreated, err = time.Parse("2006-01-02 15:04", meta.DateCreated)
	if err != nil {
		fmt.Println("Failed to parse DateCreated: " + err.Error())
	}
	article.DateModified, err = time.Parse("2006-01-02 15:04", meta.DateModified)
	if err != nil {
		article.DateModified = article.DateCreated
	}
	article.Icon = meta.Icon

	// Parse article body to valid HTML (which is safe)
	article.Body = parseArticleBodyToHtml(article_body_data)

	return nil
}

func getArticleId(r *http.Request) (string, error) {
	m := validArticle.FindStringSubmatch(r.URL.Path)
	if m == nil {
		return "", errors.New("Invalid Article Id with request: " + r.URL.Path)
	}

	return m[2], nil // The id is the second subexpression.
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getArticleId(r)
	if err != nil {
		http.NotFound(w, r)
		log.Print("Could not parse article Id from request:" + err.Error())
		return
	}
	article, err := NewArticle(id)
	if err != nil {
		http.NotFound(w, r)
		log.Print("Unknown article Id:" + id)
		return
	}
	renderTemplate(w, "article", *article)
}
