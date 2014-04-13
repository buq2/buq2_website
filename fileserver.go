package main

import (
    "net/http"
    "strings"
)

// From https://groups.google.com/forum/#!topic/golang-nuts/bStLPdIVM6w
// Returns 404 if static content directory is accessed without
// a file
func noDirListing(h http.Handler) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasSuffix(r.URL.Path, "/") {
            http.NotFound(w, r)
            return
        }
        h.ServeHTTP(w, r)
    })
}

func fileserverHandler() http.Handler {
    // Path to 'root' of static content.
    // Can be any folder. All file in this folder will be accessible
    static_path_string := executablePath() + "/static/";
    static_path := http.FileServer(http.Dir(static_path_string))

    // Serve static content from '/static/'. Word 'static'
    // is removed from the path before using the File server
    return http.StripPrefix("/static/", noDirListing(static_path) );
}
