package main

import (
  "github.com/cloudfoundry-community/go-cfenv"
  "gopkg.in/redis.v3"
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
  PageCount uint64
}

const KEY string = "PageCount"

var client *redis.Client = (*redis.Client)(nil)
var templates *template.Template = (*template.Template)(nil)

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

func resetHandler(w http.ResponseWriter, r *http.Request) {
  if client == nil {
    return
  }

  client.Set(KEY, "0", 0)
  http.Redirect(w, r, "/hello", http.StatusFound)
  return
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
  p := loadPage()
  p.PageCount = pageCount()

  fmt.Printf("A request just came in for instance %s. How exciting!\n", p.Index)

  err := templates.ExecuteTemplate(w, "hello.html", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func pageCount() uint64 {
  if client == nil {
    return 0
  }

  count, err := client.Get(KEY).Uint64()
  if err == redis.Nil {
    count = 1
  } else {
    count++
  }

  statusCmd := client.Set(KEY, fmt.Sprintf("%d", count), 0)
  if statusCmd.Err() != nil {
    panic(statusCmd)
  }

  return count
}

func loadDB(appEnv *cfenv.App) {
  services, _ := appEnv.Services.WithTag("redis")
  if len(services) > 0 {
    creds := services[0].Credentials

    client = redis.NewClient(&redis.Options{
      Addr: fmt.Sprintf("%s:%s", creds["hostname"], creds["port"]),
      Password: creds["password"].(string),
      DB: 0,
    })

    _, err := client.Ping().Result()
    if err != nil {
      client = nil
    }
  }
}

func main() {
  appEnv, _ := cfenv.Current()

  FuncMap := template.FuncMap{
    "show": func(a *Page) bool {
      return a != nil && a.PageCount > 0
    },
    "ne": func(a, b interface{}) bool {
      return a != b
    },
  }

  templates = template.Must(template.New("SampleTemplates").Funcs(FuncMap).ParseFiles("templates/hello.html"))

  loadDB(appEnv)

  http.HandleFunc("/", rootHandler)
  http.HandleFunc("/kill", killHandler)
  http.HandleFunc("/reset", resetHandler)
  http.HandleFunc("/hello", helloHandler)

  err := http.ListenAndServe(":" + os.Getenv("PORT"), nil)
  if err != nil {
    log.Fatal(err)
  }
}
