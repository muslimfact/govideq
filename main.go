package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

func videoProxy(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "Missing ?url parameter", http.StatusBadRequest)
		return
	}

	// PARSE & ENCODE ULANG (INI KUNCI)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		http.Error(w, "Invalid video URL", http.StatusBadRequest)
		return
	}
	videoURL := parsed.String()

	log.Println("▶ Streaming:", videoURL)

	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		http.Error(w, "Request error", 500)
		return
	}

	// HEADER WAJIB VIDOY
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://embed.vidoycdn.com/")
	req.Header.Set("Origin", "https://embed.vidoycdn.com")
	req.Header.Set("Accept", "*/*")

	if r.Header.Get("Range") != "" {
		req.Header.Set("Range", r.Header.Get("Range"))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Fetch failed", 502)
		return
	}
	defer resp.Body.Close()

	// === HEADER DULU ===
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Disposition", `inline; filename="video.mp4"`)

	if cl := resp.Header.Get("Content-Length"); cl != "" {
		w.Header().Set("Content-Length", cl)
	}
	if cr := resp.Header.Get("Content-Range"); cr != "" {
		w.Header().Set("Content-Range", cr)
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	http.HandleFunc("/video", videoProxy)

	log.Println("🚀 Proxy running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
