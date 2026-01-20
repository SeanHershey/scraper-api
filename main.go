package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var randomStrings = []string{
	"glitch art",
	"dark aesthetic",
	"cyberpunk",
	"neon lights",
	"vaporwave",
	"synthwave",
	"dark mode",
	"high contrast",
	"abstract dark",
	"digital art",
	"pixel art",
	"retro dark",
	"noir",
	"monochrome",
	"dark minimalism",
	"glitch effect",
	"dark fantasy",
	"gothic",
	"dark architecture",
	"neon aesthetic",
	"dark city",
	"night photography",
	"dark abstract",
	"contrast photography",
	"dark textures",
	"glitch aesthetic",
	"dark patterns",
	"neon aesthetic",
	"dark mood",
	"high contrast photography",
	"dark cyber",
	"glitchy",
	"dark neon",
	"minimalist dark",
	"dark geometric",
	"neon glow",
	"dark surreal",
	"glitchcore",
	"dark futuristic",
	"neon city",
	"dark minimal",
	"high contrast art",
	"dark digital",
	"neon art",
	"glitch photography",
	"dark modern",
	"neon abstract",
	"dark tech",
	"contrast art",
	"dark visual",
	"red neon",
	"blue cyberpunk",
	"purple glitch",
	"green neon",
	"cyan aesthetic",
	"magenta dark",
	"orange glow",
	"yellow neon",
	"pink cyber",
	"red cyberpunk",
	"blue neon lights",
	"purple vaporwave",
	"green glitch",
	"cyan synthwave",
	"magenta aesthetic",
	"orange dark",
	"yellow glow",
	"pink neon",
	"red dark aesthetic",
	"blue glitch art",
	"purple cyberpunk",
	"green neon city",
	"cyan dark mode",
	"magenta high contrast",
	"orange abstract dark",
	"yellow digital art",
	"pink pixel art",
	"red retro dark",
	"blue noir",
	"purple monochrome",
	"green dark minimalism",
	"cyan glitch effect",
	"magenta dark fantasy",
	"orange gothic",
	"yellow dark architecture",
	"pink neon aesthetic",
	"red dark city",
	"blue night photography",
	"purple dark abstract",
	"green contrast photography",
	"cyan dark textures",
	"magenta glitch aesthetic",
	"orange dark patterns",
	"yellow dark mood",
	"pink high contrast photography",
	"red dark cyber",
	"blue glitchy",
	"purple dark neon",
	"green minimalist dark",
	"cyan dark geometric",
	"magenta neon glow",
	"orange dark surreal",
	"yellow glitchcore",
	"pink dark futuristic",
	"red neon city",
	"blue dark minimal",
	"purple high contrast art",
	"green dark digital",
	"cyan neon art",
	"magenta glitch photography",
	"orange dark modern",
	"yellow neon abstract",
	"pink dark tech",
	"red contrast art",
	"blue dark visual",
	"electric blue",
	"neon red",
	"cyber purple",
	"glitch green",
	"dark cyan",
	"neon magenta",
	"vaporwave orange",
	"synthwave yellow",
	"aesthetic pink",
	"dark red",
	"neon blue",
	"cyber green",
	"glitch cyan",
	"dark magenta",
	"neon orange",
	"vaporwave yellow",
	"synthwave pink",
}

// Allowed image sources
var allowedSources = []string{
	"cosmos.so",
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ImageResponse struct {
	Query     string `json:"query"`
	ImageURL  string `json:"image_url"`
	Title     string `json:"title,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

type GoogleSearchResponse struct {
	Items []struct {
		Title       string `json:"title"`
		Link        string `json:"link"`
		DisplayLink string `json:"displayLink"`
		Image       struct {
			ContextLink     string `json:"contextLink"`
			Height          int    `json:"height"`
			Width           int    `json:"width"`
			ByteSize        int    `json:"byteSize"`
			ThumbnailLink   string `json:"thumbnailLink"`
			ThumbnailHeight int    `json:"thumbnailHeight"`
			ThumbnailWidth  int    `json:"thumbnailWidth"`
		} `json:"image"`
	} `json:"items"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	port := getEnv("PORT", "8080")

	http.HandleFunc("/api/image", requireAuth(imageHandler))
	http.HandleFunc("/", rootHandler)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := MessageResponse{
		Message: "Welcome to the Go API",
	}
	json.NewEncoder(w).Encode(response)
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := getEnv("API_KEY", "")
		if apiKey == "" {
			http.Error(w, "API key not configured on server", http.StatusInternalServerError)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format. Use: Bearer <token>", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token != apiKey {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := getRandomString()

	apiKey := getEnv("GOOGLE_API_KEY", "")
	searchEngineID := getEnv("GOOGLE_SEARCH_ENGINE_ID", "")

	if apiKey == "" || searchEngineID == "" {
		http.Error(w, "Google API credentials not configured", http.StatusInternalServerError)
		return
	}

	imageURL, title, thumbnail, source, err := searchGoogleImage(query, apiKey, searchEngineID)
	if err != nil {
		log.Printf("Error searching for image: %v", err)
		http.Error(w, fmt.Sprintf("Error searching for image: %v", err), http.StatusInternalServerError)
		return
	}

	if !isAllowedSource(source) {
		log.Printf("Image from disallowed source: %s", source)
		http.Error(w, fmt.Sprintf("No images found from allowed sources for query: %s", query), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := ImageResponse{
		Query:     query,
		ImageURL:  imageURL,
		Title:     title,
		Thumbnail: thumbnail,
	}
	json.NewEncoder(w).Encode(response)
}

func searchGoogleImage(query, apiKey, searchEngineID string) (string, string, string, string, error) {
	baseURL := "https://www.googleapis.com/customsearch/v1"

	siteFilter := buildSiteFilter()

	params := url.Values{}
	params.Add("key", apiKey)
	params.Add("cx", searchEngineID)
	params.Add("q", query)
	params.Add("searchType", "image")
	params.Add("num", "1")
	params.Add("safe", "active")
	if siteFilter != "" {
		params.Add("siteSearch", siteFilter)
		params.Add("siteSearchFilter", "i")
	}

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(requestURL)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", "", "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var searchResponse GoogleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return "", "", "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(searchResponse.Items) == 0 {
		return "", "", "", "", fmt.Errorf("no images found for query: %s", query)
	}

	item := searchResponse.Items[0]
	source := extractDomain(item.DisplayLink)
	return item.Link, item.Title, item.Image.ThumbnailLink, source, nil
}

func buildSiteFilter() string {
	if len(allowedSources) == 0 {
		return ""
	}
	filter := ""
	for i, source := range allowedSources {
		if i > 0 {
			filter += " "
		}
		filter += source
	}
	return filter
}

func extractDomain(displayLink string) string {
	domain := displayLink
	if len(domain) > 4 && domain[:4] == "www." {
		domain = domain[4:]
	}
	return domain
}

func isAllowedSource(source string) bool {
	for _, allowed := range allowedSources {
		if source == allowed {
			return true
		}
		if len(source) > len(allowed) && source[len(source)-len(allowed)-1:] == "."+allowed {
			return true
		}
	}
	return false
}

func getRandomString() string {
	return randomStrings[rand.Intn(len(randomStrings))]
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
