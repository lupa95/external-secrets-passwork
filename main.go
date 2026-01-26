// docker/main.go
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lupa95/passwork-client-go"
)

// jsonError provides a tiny, consistent error payload.
type jsonError struct {
	Error string `json:"error"`
}

func main() {
	// ----- Configuration & client initialization -----
	host := os.Getenv("PASSWORK_HOST")       // e.g. https://my‑passwork-instance.com/api/v4
	apiKey := os.Getenv("PASSWORK_API_KEY") // Passwork API key
	if host == "" || apiKey == "" {
		log.Fatal("PASSWORK_HOST and PASSWORK_API_KEY must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create a Passwork client with a 30‑second timeout.
	client := passwork.NewClient(host, apiKey, 30*time.Second)

	// Perform a single login – the SDK re‑uses the session token.
	if err := client.Login(); err != nil {
		log.Fatalf("Passwork login failed: %v", err)
	}
	log.Println("Passwork client authenticated")

	// ----- Handlers -----------------------------------------------------

	// Fetch a secret (password entry) from Passwork.
	http.HandleFunc("/passwords/", func(w http.ResponseWriter, r *http.Request) {
		secretID := strings.TrimPrefix(r.URL.Path, "/passwords/")
		if secretID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(jsonError{Error: "missing secret ID"})
			log.Println("missing secret ID")
			return
		}

		start := time.Now()
		resp, err := client.GetPassword(secretID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(jsonError{Error: "secret not found"})
			log.Printf("error fetching secret %s: %v (took %s)", secretID, err, time.Since(start))
			return
		}

		// Convert to map to modify CryptedPassword
		var data map[string]interface{}
		jsonBytes, _ := json.Marshal(resp.Data)
		json.Unmarshal(jsonBytes, &data)

		// Decode CryptedPassword from base64
		if cryptedPw, ok := data["CryptedPassword"].(string); ok && cryptedPw != "" {
			decoded, err := base64.StdEncoding.DecodeString(cryptedPw)
			if err == nil {
				data["CryptedPassword"] = string(decoded)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if encErr := json.NewEncoder(w).Encode(data); encErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(jsonError{Error: "failed to encode response"})
			log.Printf("encode error for secret %s: %v (took %s)", secretID, encErr, time.Since(start))
			return
		}
		log.Printf("served secret %s (took %s)", secretID, time.Since(start))
	})

	// Simple health‑check endpoint.
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// ----- Server with graceful shutdown -----------------------------

	srv := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	// Listen for OS signals (Ctrl‑C, Docker stop, etc.).
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Run the server in a goroutine.
	go func() {
		log.Printf("Passwork proxy listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until a termination signal arrives.
	<-stop
	log.Println("shutdown signal received")

	// Graceful shutdown with a 5‑second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
