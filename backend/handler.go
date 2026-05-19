package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler holds the dependencies for HTTP handlers.
type Handler struct {
	llm *LLMService
}

// NewHandler creates a new Handler with the given LLM service.
func NewHandler(llm *LLMService) *Handler {
	return &Handler{llm: llm}
}

// HealthHandler responds with a simple status check.
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"service": "ai2sim-backend",
	})
}

// GenerateHandler accepts a prompt and returns generated simulation code.
func (h *Handler) GenerateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Only POST is allowed")
		return
	}

	// Parse request body
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	defer r.Body.Close()

	// Validate prompt
	if req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "Prompt cannot be empty")
		return
	}

	log.Printf("Generating simulation for prompt: %q", req.Prompt)

	// Call the LLM service
	result, err := h.llm.GenerateSimulation(r.Context(), req.Prompt)
	if err != nil {
		log.Printf("Generation failed: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to generate simulation. Please try again.")
		return
	}

	// Return the result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
