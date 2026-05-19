package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// Load API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required.\n" +
			"Get one from: https://aistudio.google.com/apikey")
	}

	// Initialize the LLM service
	llmService, err := NewLLMService(apiKey)
	if err != nil {
		log.Fatalf("Failed to initialize LLM service: %v", err)
	}
	log.Println("LLM service initialized successfully")

	// Create handler
	handler := NewHandler(llmService)

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handler.HealthHandler)
	mux.HandleFunc("/api/generate", handler.GenerateHandler)

	// Apply middleware
	var finalHandler http.Handler = mux
	finalHandler = CORSMiddleware(finalHandler)
	finalHandler = LoggingMiddleware(finalHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("AI2Sim backend starting on :%s", port)
	if err := http.ListenAndServe(":"+port, finalHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
