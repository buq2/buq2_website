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
		if strings.HasSuffix(r.URL.Path, "/") || len(r.URL.Path) == 0 {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func noDirListingMustBeIn2ndSubdir(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numSubfolders := strings.Count(r.URL.Path, "/")
		if numSubfolders < 2 || strings.HasSuffix(r.URL.Path, "/") || len(r.URL.Path) == 0 {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func fileserverHandlerStatic() http.Handler {
	// Path to 'root' of static content.
	// Can be any folder. All file in this folder will be accessible
	static_path_string := executablePath() + "/static/"
	static_path := http.FileServer(http.Dir(static_path_string))

	// Serve static content from '/static/'. Word 'static'
	// is removed from the path before using the File server
	return http.StripPrefix("/static/", noDirListing(static_path))
}

func fileserverHandlerContentStatic() http.Handler {
	// Similar to fileserverHandlerStatic, but the file must be at
	// 2nd sub directory or deeper
	// Used as we want to store articles in
	// siteGlobal.ContentRoot + "/articles" but we do not want to give user
	// direct access to files at "/articles" folder
	// so we can store files under "/articles/fold1" etc. This way articles
	// are at the "/articles/" folder and images for each article can be
	// at the sub folders of "/articles/".
	static_path_string := siteGlobal.ContentRoot
	static_path := http.FileServer(http.Dir(static_path_string))
	return http.StripPrefix("/content_static/", noDirListingMustBeIn2ndSubdir(static_path))
}
