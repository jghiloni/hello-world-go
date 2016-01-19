package main

import (
  "log"
  "fmt"
  "net/http"
  "html/template"
  "os"
)

type Page struct {
  IP string
  Port string
  Index string
}

var templates = template.Must(template.ParseFiles("templates/hello.html"))

func loadPage() *Page {
  return &Page {
    IP: os.Getenv("CF_INSTANCE_IP"),
    Port: os.Getenv("CF_INSTANCE_PORT"),
    Index: os.Getenv("CF_INSTANCE_INDEX"),
  }
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path != "/" {
    http.NotFound(w,r)
    return
  }

  http.Redirect(w, r, "/hello", http.StatusFound)
  return
}

func killHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("About to kill this instance")
  os.Exit(1)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
  p := loadPage()

  fmt.Printf("A request just came in for instance %s. How exciting!\n", p.Index)

  err := templates.ExecuteTemplate(w, "hello.html", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func main() {
  p := loadPage()

  fmt.Printf("%+v\n", p)

  http.HandleFunc("/", rootHandler)
  http.HandleFunc("/kill", killHandler)
  http.HandleFunc("/hello", helloHandler)

  err := http.ListenAndServe(":" + os.Getenv("PORT"), nil)
  if err != nil {
    log.Fatal(err)
  }
}
