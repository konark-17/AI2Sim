package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"ai2sim/backend/internal/llm"
)

// Handler holds dependencies for the API handlers.
type Handler struct {
	llmClient *llm.Client
}

// NewHandler creates a new API handler.
func NewHandler(client *llm.Client) *Handler {
	return &Handler{llmClient: client}
}

// RegisterRoutes wires up all HTTP routes on the Fiber app.
func RegisterRoutes(app *fiber.App, h *Handler) {
	v1 := app.Group("/api/v1")
	v1.Get("/health", h.Health)
	v1.Post("/parameters", h.GetParameters)
	v1.Get("/simulate", h.SimulateStream)
}

// ─── Health ───────────────────────────────────────────────────────────────────

func (h *Handler) Health(c *fiber.Ctx) error {
	ready, msg := h.llmClient.Status()
	return c.JSON(fiber.Map{
		"status":   "ok",
		"service":  "ai2sim-backend",
		"provider": h.llmClient.Provider(),
		"model":    h.llmClient.Model(),
		"ready":    ready,
		"message":  msg,
	})
}

// ─── Parameters ───────────────────────────────────────────────────────────────

type parametersRequest struct {
	Prompt string `json:"prompt"`
}

// GetParameters returns an inferred parameter schema for the user's prompt.
func (h *Handler) GetParameters(c *fiber.Ctx) error {
	var req parametersRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if req.Prompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "prompt is required",
		})
	}

	schema := llm.DefaultParamsForPrompt(req.Prompt)
	return c.JSON(fiber.Map{"parameters": schema})
}

// ─── Simulate (SSE Stream) ────────────────────────────────────────────────────

// SimulateStream streams LLM-generated simulation code via Server-Sent Events.
// Query params: prompt (required), params (optional JSON object of param values)
func (h *Handler) SimulateStream(c *fiber.Ctx) error {
	prompt := c.Query("prompt")
	if prompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "prompt query parameter is required",
		})
	}

	// Parse optional params JSON
	var paramValues map[string]any
	if paramsJSON := c.Query("params"); paramsJSON != "" {
		if err := json.Unmarshal([]byte(paramsJSON), &paramValues); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid params JSON: %v", err),
			})
		}
	}

	// Validate and fill defaults
	schema := llm.DefaultParamsForPrompt(prompt)
	if paramValues != nil {
		paramValues = llm.ValidateParams(paramValues, schema)
	} else {
		paramValues = llm.ParamsToMap(schema)
	}

	streamCtx, cancelStream := context.WithCancel(context.Background())

	// SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	capturedPrompt := prompt
	capturedParams := paramValues
	capturedClient := h.llmClient

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cancelStream()

		_, err := capturedClient.StreamSimulation(
			streamCtx,
			capturedPrompt,
			capturedParams,
			func(token string) error {
				// Send safely JSON-encoded token object: {"token":"..."}
				payload, _ := json.Marshal(map[string]string{"token": token})
				_, writeErr := fmt.Fprintf(w, "data: %s\n\n", payload)
				if writeErr != nil {
					return writeErr
				}
				return w.Flush()
			},
		)
		if err != nil {
			log.Printf("SimulateStream error for prompt %q: %v", capturedPrompt, err)
			errPayload, _ := json.Marshal(map[string]string{
				"error": fmt.Sprintf("Generation failed: %v", err),
			})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", errPayload)
			_ = w.Flush()
			return
		}

		// Signal completion
		fmt.Fprint(w, "event: done\ndata: [DONE]\n\n")
		_ = w.Flush()
	})

	return nil
}
