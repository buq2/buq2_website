package main

import (
	"bytes"
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
	"sort"
	"strings"
)

type Article struct {
	SiteGlobal

	// Following are not from the file
	Id   string
	Link string

	// Following are parsed from the files
	Body         template.HTML
	Title        string
	LongTitle    string
	Description  string
	DateCreated  ParsableTime
	DateModified ParsableTime
	Icon         string
	Tags         []string
	Comments     *[]Comment

	// If user tries to add comment, this will be
	// filled with data
	NewComment NewComment

	// Options which read from the file, and affect how the data is processed
	// but should not be displayed on the final HTML
	CreateToc bool
}

var validArticle = regexp.MustCompile("^/(article)/([a-zA-Z0-9_]+)$")

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
            if($(this).attr("alt")) {
                if($(this).attr("class")) {
                    // Img has class, use this as figures class. Remove
                    // class from Img
                    $(this).wrap('<figure class="' + $(this).attr("class") + '"></figure>')
                        .after('<figcaption>'+$(this).attr("alt")+'</figcaption>');
                    $(this).removeAttr('class');
                } else {
                    $(this).wrap('<figure></figure>')
                        .after('<figcaption>'+$(this).attr("alt")+'</figcaption>');
                }
            }
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
		showRecaptcha();
    });
</script>
<!--***********
 Recaptcha
************-->
<script type="text/javascript" src="http://www.google.com/recaptcha/api/js/recaptcha_ajax.js"></script>
<script>
function showRecaptcha() {
	Recaptcha.create("6LdVRvcSAAAAAAOqeWSPRmZmvBSZZpZ2spv6L_fd",
    "captchadiv",
    {
      theme: "red"
    }
  );
}
</script>
`)

const (
	articleFolder    = "/articles/"
	articleExtension = ".md"
)

// Helper type for sorting
type ByCreationDateNewestFirst []*Article

// Helper funcition for sorting
func (this ByCreationDateNewestFirst) Len() int {
	return len(this)
}

// Helper funcition for sorting
func (this ByCreationDateNewestFirst) Less(i, j int) bool {
	return this[i].DateCreated.After(this[j].DateCreated.Time)
}

// Helper funcition for sorting
func (this ByCreationDateNewestFirst) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func GetArticleFolder() string {
	return siteGlobal.ContentRoot + "/" + articleFolder
}

func GetAllArticles() []*Article {
	ids := getAllArticleIds()
	articles := make([]*Article, len(ids))

	for idx, id := range ids {
		articles[idx], _ = NewArticle(id)
	}

	sort.Sort(ByCreationDateNewestFirst(articles))

	return articles
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GetArticlesByTag(tag string) ([]*Article, error) {
	articles_all := GetAllArticles()

	articles := []*Article{}
	for _, article := range articles_all {
		if stringInSlice(tag, article.Tags) {
			articles = append(articles, article)
		}
	}

	sort.Sort(ByCreationDateNewestFirst(articles))

	return articles, nil
}

func SplitRawArticlesIntoColumns(articles_raw []*Article) Articles {
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

	return articles
}

func NewArticle(id string) (*Article, error) {
	// Try to find the data to the article with certain id
	filename := GetArticleFolder() + "/" + id + articleExtension
	article_data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	article := new(Article)
	err = parseRawTextArticleData(article_data, article)
	// Save HeadAfterScripts from overwriting
	additional_scripts := article.HeadAfterScripts;

	// Other
	article.SiteGlobal = siteGlobal
	article.Id = id
	article.Scripts = articleScripts
	article.Link = websiteAddress() + "/article/" + article.Id
	article.Keywords = article.Tags
	article.Comments, _ = GetComments(id)
	article.HeadAfterScripts = additional_scripts;

	return article, err
}

func getAllArticleIds() []string {
	ids := []string{}

	files, err := ioutil.ReadDir(GetArticleFolder())
	if err != nil {
		return ids
	}

	for _, file := range files {
		filename := file.Name()
		ext := path.Ext(filename)
		if ext == articleExtension {
			name := filename[0 : len(filename)-len(ext)]
			ids = append(ids, name)
		}
	}

	return ids
}

func escapeLatexInner(input []byte) []byte {

	// Replace all "_"" with escaped version ("\_")
	tmp := bytes.Replace(input, []byte("_"), []byte("\\_"), -1)
	// Replace "\\" (matrix line change) with escaped version
	// ("\\\\")
	tmp = bytes.Replace(tmp, []byte("\\\\"), []byte("\\\\\\\\"), -1)
	// Remove white space around "=" (as other wise it might be taken as
	// a start of a header)
	re := regexp.MustCompile("[\\s]*=[\\s]*")
	tmp = re.ReplaceAll(tmp, []byte("="))
	return tmp
}

func escapeLatex(input []byte) []byte {
	// (?s) sets dot to match new lines
	// Escape all text which is between $+ tags
	// This is not perfect, but good enough for my use
	re := regexp.MustCompile("(?s)\\$+(.+?)\\$+")
	output := re.ReplaceAllFunc(input, escapeLatexInner)
	return output
}

func parseArticleBodyToHtml(article_body_data []byte, article Article) template.HTML {
	// Convert markdown to thml
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_USE_XHTML
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	if article.CreateToc {
		// Only create TOC if it is specifically requested
		htmlFlags |= blackfriday.HTML_TOC
	}
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
	// Footnotes allow use of "[^1]" style footnotes
	extensions |= blackfriday.EXTENSION_FOOTNOTES

	// Escape latex blocks
	article_body_data_escaped := escapeLatex(article_body_data)

	body_markdown := blackfriday.Markdown(article_body_data_escaped, renderer, extensions)

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
	err := json.Unmarshal(article_meta_data, article)
	if err != nil {
		fmt.Println("Failed to parse article meta data. Returning empty meta data: " + err.Error())
	}

	// Parse article body to valid HTML (which is safe)
	article.Body = parseArticleBodyToHtml(article_body_data, *article)

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

	article, err = CheckNewComment(r, article)

	renderTemplate(w, "article", *article)
}
