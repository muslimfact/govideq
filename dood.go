package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	http.HandleFunc("/video", videoProxy)

	log.Println("🚀 Go video proxy running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func videoProxy(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "missing ?url parameter", http.StatusBadRequest)
		return
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	client := &http.Client{
		Timeout: 0, // streaming
	}

	req, err := http.NewRequest("GET", parsed.String(), nil)
	if err != nil {
		http.Error(w, "request error", 500)
		return
	}

	// 🔥 PENTING: samakan header dengan fetch script
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://dood.video/")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "fetch error", 500)
		return
	}
	defer resp.Body.Close()

	// forward headers
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(resp.StatusCode)

	// stream body
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, _ = w.Write(buf[:n])
			w.(http.Flusher).Flush()
		}
		if err != nil {
			break
		}
	}

	log.Println("▶ stream done:", time.Now())
}
