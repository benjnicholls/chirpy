package main

import (
	"chirpy/internal/database"
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	platform       string
	secretKey      string
}

func main() {
	const port = "8080"
	const filePathRoot = "."

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbURL := getEnvVariable("DB_URL")
	platform := getEnvVariable("PLATFORM")
	secretString := getEnvVariable("SECRET")

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	dbQueries := database.New(conn)

	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secretKey:      secretString,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filePathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerNumberOfRequests)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.handlerGetSingleChirp)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)

	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
