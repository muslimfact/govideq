package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func videoProxy(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "Missing ?url", 400)
		return
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		http.Error(w, "Invalid URL", 400)
		return
	}

	log.Println("▶ Streaming:", parsed.String())

	req, err := http.NewRequest("GET", parsed.String(), nil)
	if err != nil {
		http.Error(w, "Request error", 500)
		return
	}

	// HEADER WAJIB VIDOY
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://embed.vidoycdn.com/")
	req.Header.Set("Origin", "https://embed.vidoycdn.com")
	req.Header.Set("Accept", "*/*")

	// Forward Range
	if r.Header.Get("Range") != "" {
		req.Header.Set("Range", r.Header.Get("Range"))
	}

	client := &http.Client{} // ❗ TANPA TIMEOUT

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Fetch failed", 502)
		return
	}
	defer resp.Body.Close()

	// === FORWARD HEADER DULU ===
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	} else {
		w.Header().Set("Content-Type", "video/mp4")
	}

	w.Header().Set("Accept-Ranges", "bytes")

	if cl := resp.Header.Get("Content-Length"); cl != "" {
		w.Header().Set("Content-Length", cl)
	}
	if cr := resp.Header.Get("Content-Range"); cr != "" {
		w.Header().Set("Content-Range", cr)
	}

	w.Header().Set("Content-Disposition", `inline; filename="video.mp4"`)

	// BARU status code
	w.WriteHeader(resp.StatusCode)

	// STREAM
	io.Copy(w, resp.Body)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/video", videoProxy)

	log.Println("🚀 Proxy running on :" + port)
	log.Println(http.ListenAndServe(":"+port, nil))
}
