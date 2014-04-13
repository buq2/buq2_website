package main

import (
    "io/ioutil"
    "net/http"
    "html/template"
    "regexp"
    "errors"
    "github.com/russross/blackfriday"
    "strings"
    "encoding/json"
    "fmt"
    "path"
)

var validPath = regexp.MustCompile("^/(article)/([a-zA-Z0-9]+)$")
var templates = template.Must(template.ParseFiles("article.html"))

type PageMetaData struct {
    Title string;
    ShortDescription string;
    CreateDate string;
    ModifiedDate string;
}

type Page struct {
    Meta PageMetaData
    Body template.HTML
}

func getAllArticleIds() ([]string) {
    folder := "./"
    ids := []string{};
    
    files,err := ioutil.ReadDir(folder)
    if err != nil {
        return ids;
    }

    for _, file := range files {
        filename := file.Name();
        ext := path.Ext(filename);
        if ext == ".txt" {
            name := filename[0:len(filename)-len(ext)];
            ids = append(ids, name)
        }
    }

    return ids;
}

func parseArticleData(article_data []byte) ([]byte, PageMetaData) {
    article_header_separator := "---------- META END ----------";
    separator_len := len(article_header_separator)
    separator_begin := strings.Index(string(article_data), article_header_separator)

    article_body_data := []byte{};
    article_meta_data := []byte{};

    if separator_begin > 0 {
        // Found separator, split the data into meta data and body data
        article_meta_data = article_data[:separator_begin-1];
        article_body_data = article_data[separator_begin+separator_len:];
    } else {
        // Did not find separator, meta data is empty
        article_body_data = article_data
    }

    meta := PageMetaData{}
    if err := json.Unmarshal(article_meta_data, &meta); err != nil {
        fmt.Println("Failed to parse article meta data")
    }
    
    return article_body_data, meta
}

func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    article_data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    body, meta := parseArticleData(article_data)

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
        Meta: meta,
        Body: template.HTML(body_markdown),
    }, nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page Title")
    }
    return m[2], nil // The title is the second subexpression.
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
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
