package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "runtime"
    "html/template"
    "regexp"
    "errors"
    "github.com/russross/blackfriday"
)

var validPath = regexp.MustCompile("^/(article)/([a-zA-Z0-9]+)$")
var templates = template.Must(template.ParseFiles("article.html"))

type Page struct {
    Title string
    Body  template.HTML
}


func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile(filename)
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
    
    body_markdown := blackfriday.Markdown(body, renderer, extensions);
    
    return &Page{
        Title: title,
        Body: template.HTML(body_markdown),
    }, nil
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page Title")
    }
    return m[2], nil // The title is the second subexpression.
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
    title, err := getTitle(w, r)
    if err != nil {
        return
    }
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "article", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    http.HandleFunc("/", mainHandler)
    http.HandleFunc("/article/", viewHandler)
    http.ListenAndServe(":8080", nil)
}
