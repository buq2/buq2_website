package main

import (
	"fmt"
	"net/http"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println(getAllArticleIds())
	http.HandleFunc("/", articlesHandler)
	http.HandleFunc("/article/", articleHandler)
	http.Handle("/static/", fileserverHandler())
	http.ListenAndServe(":8080", nil)
}
