// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// NanoMon API server
// ----------------------------------------------------------------------------

package main

import (
	"log"
	"os"
	"regexp"
	"time"

	"nanomon/services/common/database"

	"github.com/benc-uk/go-rest-api/pkg/auth"
	"github.com/benc-uk/go-rest-api/pkg/env"
	"github.com/benc-uk/go-rest-api/pkg/logging"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	_ "github.com/joho/godotenv/autoload"
)

var (
	version     = "0.0.0"            // App version number, injected at build time
	buildInfo   = "No build details" // Build details, injected at build time
	serviceName = "NanoMon"
	defaultPort = 8000
)

// This scope is used to validate access to the API. The app registration must
// - be configured to allow & expose this scope. Also see frontend/app.mjs
const authScope = "system.admin"

func main() {
	// Port to listen on, change the default as you see fit
	serverPort := env.GetEnvInt("PORT", defaultPort)

	// Core of the REST API
	router := chi.NewRouter()

	// Note this will exit the process if the DB connection fails, so no need to check for errors
	db := database.ConnectToDB()

	api, err := NewAPI(db, env.GetEnvBool("ENABLE_PROMETHEUS", false))
	if err != nil {
		log.Fatalf("### ❌ Failed to create API: %v", err)
	}

	// Some basic middleware, change as you see fit, see: https://github.com/go-chi/chi#core-middlewares
	router.Use(middleware.RealIP)
	// Filtered request logger, exclude /metrics & /health endpoints
	router.Use(logging.NewFilteredRequestLogger(regexp.MustCompile(`(^/metrics)|(^/api/health)`)))
	router.Use(middleware.Recoverer)

	// Some custom middleware for very permissive CORS policy
	router.Use(cors.AllowAll().Handler)

	// Protected routes
	router.Group(func(privateRouter chi.Router) {
		// Authentication can be switched on or off
		clientID := os.Getenv("AUTH_CLIENT_ID")
		if clientID == "" {
			log.Println("### 🚨 No AUTH_CLIENT_ID set, skipping auth validation")
		} else {
			log.Println("### 🔐 Auth enabled, validating JWT tokens")

			// Validate JWT tokens using the Microsoft common public key endpoint and our scope
			jwtValidator := auth.NewJWTValidator(
				clientID,
				"https://login.microsoftonline.com/common/discovery/v2.0/keys",
				authScope,
			)

			privateRouter.Use(jwtValidator.Middleware)
		}

		// These routes carry out create, update, delete operations
		api.addProtectedRoutes(privateRouter)
	})

	// Anonymous routes
	router.Group(func(publicRouter chi.Router) {
		// Add optional root, health & status endpoints
		api.AddHealthEndpoint(publicRouter, "api/health")
		api.AddStatusEndpoint(publicRouter, "api/status")
		api.AddOKEndpoint(publicRouter, "api/")

		// Rest of the NanoMon routes which are public
		api.addAnonymousRoutes(publicRouter)
	})

	log.Printf("### ⚓ API routes configured")

	// Start ticker to check the DB connection, and set the health status
	go func() {
		ticker := time.Tick(5 * time.Second)

		for range ticker {
			_ = db.Ping(func(dbHealthy bool) {
				api.Healthy = dbHealthy
			})
		}
	}()

	// Start the API server, this function will block until the server is stopped
	api.StartServer(serverPort, router, 10*time.Second)
}
