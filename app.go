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

	// Global CORS wrapper for all routes registered on DefaultServeMux
	cors := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, cors(http.DefaultServeMux)))
}
