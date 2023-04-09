package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"

	_ "github.com/joho/godotenv/autoload"
)

var (
	version   = "0.0.0"            // App version number, set at build time with -ldflags "-X 'main.version=1.2.3'"
	buildInfo = "No build details" // Build details, set at build time with -ldflags "-X 'main.buildInfo=Foo bar'"
)

func main() {
	var dir string

	flag.StringVar(&dir, "dir", "./", "the directory to serve files from")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	r := chi.NewRouter()

	// The simple config endpoint
	r.Get("/config", routeConfig)

	FileServer(r, dir)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("### 🌐 NanoMon Frontend, listening on port:", port)
	log.Println("### Version:", version)
	log.Println("### Serving app content from:", dir)
	log.Fatal(srv.ListenAndServe())
}

// Simple config endpoint, returns API_ENDPOINT & AUTH_CLIENT_ID vars to front end
func routeConfig(resp http.ResponseWriter, req *http.Request) {
	apiEndpoint := os.Getenv("API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "/"
	}

	authTenant := os.Getenv("AUTH_TENANT")
	if authTenant == "" {
		authTenant = "common"
	}

	authClientID := os.Getenv("AUTH_CLIENT_ID")

	config := map[string]string{
		"API_ENDPOINT":   apiEndpoint,
		"AUTH_CLIENT_ID": authClientID,
		"VERSION":        version,
		"BUILD_INFO":     buildInfo,
		"AUTH_TENANT":    authTenant,
	}

	configJSON, _ := json.Marshal(config)

	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.Header().Add("Content-Type", "application/json")
	_, _ = resp.Write([]byte(configJSON))
}

func FileServer(router *chi.Mux, root string) {
	fs := http.FileServer(http.Dir(root))

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(root + r.RequestURI); os.IsNotExist(err) {
			http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		} else {
			fs.ServeHTTP(w, r)
		}
	})
}
