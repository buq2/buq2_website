package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "runtime"
    "html/template"
    "regexp"
    "errors"
)

var validPath = regexp.MustCompile("^/(article)/([a-zA-Z0-9]+)$")
var templates = template.Must(template.ParseFiles("article.html"))

type Page struct {
    Title string
    Body  []byte
}


func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
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
