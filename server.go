package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v1"
)

var (
	addr       = flag.String("http", ":8080", "HTTP service address (e.g., '127.0.0.1:6060' or just ':6060')")
	configPath = flag.String("conf", "settings.yaml", "path to settings file")
)

type userConf struct {
	URL, Token string
	Goals      []map[string]string
}

type config map[string]userConf

func readConfig() config {
	b, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatal("Couldn't read settings file: ", err)
	}
	var c config
	if err = yaml.Unmarshal(b, &c); err != nil {
		log.Fatal("Failed to parse settings: ", err)
	}
	return c
}

func postDatapoint(k, user, token string) error {
	url := fmt.Sprintf("https://www.beeminder.com/api/v1/users/%s/goals/%s/datapoints.json?value=1&comment=entered+with+clean-bee+form&auth_token=%s", user, k, token)
	resp, err := http.PostForm(url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	b, _ := ioutil.ReadAll(resp.Body)
	return errors.New(string(b))
}

func handle(w http.ResponseWriter, r *http.Request, user, token, static string) {
	switch r.Method {
	case "GET":
		fmt.Fprint(w, static)
	case "POST":
		postData(w, r, user, token)
	default:
		http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
	}
}

func makeHandler(user string, conf userConf, tmpl *template.Template) http.HandlerFunc {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, conf.Goals); err != nil {
		log.Fatal("Error executing template: ", err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		handle(w, r, user, conf.Token, buf.String())
	}
}

func postData(w http.ResponseWriter, r *http.Request, user, token string) {
	if err := r.ParseForm(); err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	for k := range r.PostForm {
		if r.PostForm.Get(k) == "yes" {
			if err := postDatapoint(k, user, token); err != nil {
				writeError(w, err)
				return
			}
		}
	}
	http.ServeFile(w, r, "success.html")
}

func writeError(w http.ResponseWriter, e error) {
	w.Write([]byte("<body>Whoops, something went wrong! Here's the error message:\n<pre style=\"word-wrap: break-word; white-space: pre-wrap;\">"))
	w.Write([]byte(e.Error()))
	w.Write([]byte("</pre></body>"))
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	if _, err := os.Stat("success.html"); err != nil {
		log.Fatal("Couldn't read success.html file: ", err)
	}
	if _, err := os.Stat("success_baby.jpg"); err != nil {
		log.Fatal("Couldn't read success_baby.jpg file: ", err)
	}
	tmpl := template.Must(template.ParseFiles("form.html"))
	for user, conf := range readConfig() {
		http.Handle(conf.URL, makeHandler(user, conf, tmpl))
	}
	http.HandleFunc("/go/success_baby.jpg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "success_baby.jpg")
	})

	fmt.Println(http.ListenAndServe(*addr, nil).Error())
}
