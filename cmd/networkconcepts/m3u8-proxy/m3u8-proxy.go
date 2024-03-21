package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func handleTsChunkProxy(targetURL string, w http.ResponseWriter) {
	// 1. Fetch the original resource
	if targetURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, "Error fetching PNG: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 2. Check the content type
	if resp.Header.Get("Content-Type") != "image/png" {
		http.Error(w, "Resource is not a PNG", http.StatusBadRequest)
		return
	}

	// 3. Process the PNG data
	pngData, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading PNG data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(pngData) <= 91 {
		http.Error(w, "PNG data is too short", http.StatusBadRequest)
		return
	}

	strippedData := pngData[96:]

	// 4. Return the modified content
	w.Header().Set("Content-Type", "video/MP2T")
	_, err = io.Copy(w, bytes.NewReader(strippedData))
	if err != nil {
		// Handle potential error during copy
		http.Error(w, "Error sending MP2T data", http.StatusInternalServerError)
	}
}

func handleM3U8Proxy(targetURL string, w http.ResponseWriter) {
	if targetURL == "" {
		http.Error(w, "Missing 'm3u8' parameter", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, "Error fetching M3U8: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading M3U8: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(data), "\n")

	// Modify and construct the new file content
	var modifiedContent strings.Builder
	for _, line := range lines {
		if strings.HasPrefix(line, "https://") {
			// Replace the URL
			newLine := "http://localhost:8686?url=" + url.QueryEscape(strings.TrimSpace(line))
			modifiedContent.WriteString(newLine + "\n")
		} else {
			modifiedContent.WriteString(line + "\n")
		}
	}

	// Set the content type and serve the modified M3U8
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Write([]byte(modifiedContent.String()))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("m3u8") {
		var targetURL string = r.URL.Query().Get("m3u8")
		handleM3U8Proxy(targetURL, w)
	} else if r.URL.Query().Has("url") {
		var targetURL string = r.URL.Query().Get("url")
		handleTsChunkProxy(targetURL, w)
	} else {
		http.Error(w, "Missing 'm3u8' or 'url' parameter", http.StatusBadRequest)
		return
	}
}

func main() {
	http.HandleFunc("/", proxyHandler)
	log.Fatal(http.ListenAndServe(":8686", nil)) // Start the proxy server
}
