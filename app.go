package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/et-hicks/imitation-backend/src"
)

//go:embed templates/*
var resources embed.FS

var t = template.Must(template.ParseFS(resources, "templates/*"))

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"

	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{
			"Region": os.Getenv("FLY_REGION"),
		}

		t.ExecuteTemplate(w, "index.html.tmpl", data)
	})

	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{
			"Region": os.Getenv("FLY_REGION"),
		}

		t.ExecuteTemplate(w, "about.html.tmpl", data)
	})

	// Serve JSON data from embedded file
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		content, err := resources.ReadFile("data/data.json")
		if err != nil {
			http.Error(w, "failed to read data", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(content)
	})

	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
