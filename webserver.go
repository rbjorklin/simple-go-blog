package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var templates = template.Must(template.ParseFiles("templates/posts.html", "templates/root.html", "templates/about.html", "templates/generic.html"))
var validPath = regexp.MustCompile("^(/|/about/|/posts/(20[0-9]{2}/[0-1][0-9]/[0-3][0-9]/[a-zA-Z0-9]+)?)$")
var posts []string
var config = Configuration{SiteName: "example.com", Port: "8080", AboutDesc: "This is the about string", RootDesc: "This is the root string", PostsDesc: "This is the posts string"}

func main() {
	fd, err := os.Open("config.json")
	defer fd.Close()
	if err != nil {
		log.Print(err)
		log.Print("Writing defaults to config.json and continuing")
		fdw, err := os.Create("config.json")
		check(err)
		defer fdw.Close()
		enc := json.NewEncoder(fdw)
		if err := enc.Encode(&config); err != nil {
			log.Println(err)
		}
	} else {
		dec := json.NewDecoder(fd)
		if err := dec.Decode(&config); err != nil {
			log.Print(err)
			os.Exit(1)
		}
	}
	updatePosts()
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/posts/", makeHandler(postsHandler))
	http.HandleFunc("/about/", makeHandler(aboutHandler))
	http.HandleFunc("/", makeHandler(rootHandler))
	http.ListenAndServe(strings.Join([]string{":", config.Port}, ""), nil)
}

type Configuration struct {
	SiteName  string `json:"sitename,omitempty"`
	Port      string `json:"port,omitempty"`
	AboutDesc string `json:"about,omitempty"`
	RootDesc  string `json:"root,omitempty"`
	PostsDesc string `json:"posts,omitempty"`
}

type Page struct {
	SiteName    string
	PageTitle   string
	PostTitle   string
	Body        template.HTML
	PostDate    string
	Description string
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.URL.Path)
		m := validPath.FindStringSubmatch(r.URL.Path)
		log.Print(m)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, strings.TrimRight(m[0], "/"))
	}
}

func check(err error) {
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func updatePosts() {
	posts = []string{}
	anon := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		posts = append(posts, path)
		return nil
	}
	filepath.Walk("posts", anon)
}

func rootHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(posts[len(posts)-1], config.RootDesc, config.SiteName)
	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	renderTemplate(w, "root", p)
}

func postsHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Print("Title is: " + title)
	log.Print(len(strings.Split(title, "/")))
	if len(strings.Split(title, "/")) > 2 {
		p, err := loadPage(title+".html", config.RootDesc, config.SiteName)
		if err != nil {
			log.Print(err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		renderTemplate(w, "root", p)
	} else {
		r := strings.NewReplacer("_", " ")
		var data bytes.Buffer
		for i := len(posts) - 1; i >= 0; i-- {
			s := posts[i]
			row := fmt.Sprintf("<p><a href=\"/%s\">%s</a>\n\t\t\t\t<span class=\"blog-post-meta\">%s</span>\n\t\t\t\t</p>\n",
				strings.Split(s, ".html")[0], r.Replace(strings.Split(filepath.Base(s), ".html")[0]), dateFromPath(s))
			log.Print(s)
			data.Write([]byte(row))
		}
		p := &Page{SiteName: config.SiteName, PageTitle: "blargh", Body: template.HTML(data.String()), PostDate: "N/A", PostTitle: "plupp", Description: config.PostsDesc}
		renderTemplate(w, "posts", p)
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title+".html", config.AboutDesc, title)
	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	renderTemplate(w, "about", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loadPage(path string, desc string, pageTitle string) (*Page, error) {
	path = strings.TrimLeft(path, "/")
	log.Print("loadPage path: " + path)
	log.Print(config.SiteName)
	date := dateFromPath(path)
	// When did I add this? Is this for the PostTitle?
	r := strings.NewReplacer("_", " ")
	postTitle := r.Replace(strings.Split(filepath.Base(path), ".html")[0])
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &Page{SiteName: config.SiteName, PageTitle: pageTitle, Body: template.HTML(body), PostDate: date, PostTitle: postTitle, Description: desc}, nil
}

func dateFromPath(path string) string {
	split_path := strings.Split(path, "/")
	if len(split_path) != 5 {
		return "N/A"
	}
	return strings.Join(split_path[1:4], "-")
}
