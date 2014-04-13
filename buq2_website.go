package main

import (
    "fmt"
    "net/http"
    "runtime"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    http.HandleFunc("/", mainHandler)
    http.HandleFunc("/article/", articleHandler)
    http.Handle("/static/", fileserverHandler())
    http.ListenAndServe(":8080",  nil)
}
