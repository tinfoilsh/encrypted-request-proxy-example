package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

const (
	// Header containing the enclave URL that the client verified
	enclaveURLHeader = "X-Tinfoil-Enclave-Url"

	// Add custom headers to allowHeaders to allow them through CORS
	allowHeaders = "Accept, Authorization, Content-Type, Ehbp-Encapsulated-Key, X-Tinfoil-Enclave-Url, Your-Custom-Request-Header"

	// Add custom headers to exposeHeaders to make them readable by the browser
	exposeHeaders = "Ehbp-Response-Nonce, Ehbp-Fallback, Your-Custom-Response-Header"
)

// These encryption headers must be preserved for the protocol to work
var (
	ehbpRequestHeaders  = []string{"Ehbp-Encapsulated-Key"}
	ehbpResponseHeaders = []string{"Ehbp-Response-Nonce", "Ehbp-Fallback"}
)

func main() {
	http.HandleFunc("/v1/chat/completions", proxyHandler)

	log.Println("proxy listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request from %s", r.Method, r.RemoteAddr)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
	w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Get upstream URL from the X-Tinfoil-Enclave-Url header
	upstreamBase := r.Header.Get(enclaveURLHeader)
	if upstreamBase == "" {
		log.Println("Error: X-Tinfoil-Enclave-Url header not provided")
		http.Error(w, "X-Tinfoil-Enclave-Url header required", http.StatusBadRequest)
		return
	}
	upstreamURL := upstreamBase + r.URL.Path

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if accept := r.Header.Get("Accept"); accept != "" {
		req.Header.Set("Accept", accept)
	}

	// Add your Tinfoil API key as the Authorization header
	apiKey := os.Getenv("TINFOIL_API_KEY")
	if apiKey == "" {
		log.Println("Error: TINFOIL_API_KEY environment variable not set")
		http.Error(w, "TINFOIL_API_KEY not set", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Required: Copy encryption headers from the client request
	copyHeaders(req.Header, r.Header, ehbpRequestHeaders...)

	// Optional: Read and strip custom headers for logging, routing, or business logic
	// These headers are not forwarded to the upstream server
	if customHeader := r.Header.Get("Your-Custom-Request-Header"); customHeader != "" {
		log.Printf("Custom request header received: %s", customHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Required: Copy encryption headers from the upstream response
	copyHeaders(w.Header(), resp.Header, ehbpResponseHeaders...)

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}

	// Optional: Add custom headers to the response based on proxy logic
	w.Header().Set("Your-Custom-Response-Header", "response-value")
	if te := resp.Header.Get("Transfer-Encoding"); te != "" {
		w.Header().Set("Transfer-Encoding", te)
		w.Header().Del("Content-Length")
	}

	w.WriteHeader(resp.StatusCode)

	if flusher, ok := w.(http.Flusher); ok {
		fw := flushWriter{ResponseWriter: w, Flusher: flusher}
		if _, copyErr := io.Copy(&fw, resp.Body); copyErr != nil {
			log.Printf("stream copy failed: %v", copyErr)
		}
		return
	}

	if _, copyErr := io.Copy(w, resp.Body); copyErr != nil {
		log.Printf("response copy failed: %v", copyErr)
	}
}

type flushWriter struct {
	http.ResponseWriter
	http.Flusher
}

func (fw *flushWriter) Write(p []byte) (int, error) {
	n, err := fw.ResponseWriter.Write(p)
	if fw.Flusher != nil {
		fw.Flush()
	}
	return n, err
}

func copyHeaders(dst, src http.Header, keys ...string) {
	for _, key := range keys {
		if value := src.Get(key); value != "" {
			dst.Set(key, value)
		}
	}
}
