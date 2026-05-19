package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"google.golang.org/genai"
)

// LLMService wraps the Gemini client for simulation generation.
type LLMService struct {
	client *genai.Client
	model  string
}

// NewLLMService creates a new LLM service with the given API key.
func NewLLMService(apiKey string) (*LLMService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &LLMService{
		client: client,
		model:  "gemini-2.0-flash",
	}, nil
}

const systemPrompt = `You are an expert JavaScript simulation generator for an HTML5 Canvas.

When given a description of a simulation, you MUST return ONLY valid JSON (no markdown, no code fences) with this exact structure:
{
  "code": "<JavaScript code>",
  "parameterDefs": [
    { "name": "paramName", "label": "Human Label", "min": 0, "max": 100, "step": 1 }
  ],
  "defaults": { "paramName": 50 }
}

Rules for the "code" field:
1. The code is the BODY of a function that receives these arguments: (ctx, canvas, parameters, deltaTime)
   - ctx: a CanvasRenderingContext2D already set up
   - canvas: the HTMLCanvasElement (use canvas.width, canvas.height)
   - parameters: an object with the current slider values (keys match parameterDefs[].name)
   - deltaTime: milliseconds since last frame
2. The code runs EVERY animation frame. Do NOT include requestAnimationFrame — the caller handles that.
3. The code MUST clear the canvas at the start: ctx.clearRect(0, 0, canvas.width, canvas.height)
4. You may use a global object called "simState" to persist state across frames. 
   Initialize it on first call like: if (!window.simState) { window.simState = { ... }; }
   Always re-initialize simState if the structure needs to change: if (!window.simState || !window.simState.initialized) { ... }
5. Make the simulation visually impressive — use colors, gradients, particles, physics, etc.
6. Always respond to parameter changes in real-time by reading from the "parameters" object each frame.
7. Use standard Canvas 2D API only — no WebGL, no external libraries.

Rules for "parameterDefs":
1. Provide 2–6 meaningful parameters that let the user tweak the simulation.
2. Each parameter needs: name (camelCase), label (human readable), min, max, step.
3. Choose sensible ranges.

Rules for "defaults":
1. Provide a default value for every parameter defined in parameterDefs.
2. Defaults should produce a visually appealing starting state.

IMPORTANT: Return ONLY the JSON object. No explanation, no markdown fences, no extra text.`

// GenerateSimulation calls Gemini with the user's prompt and returns structured simulation data.
func (s *LLMService) GenerateSimulation(ctx context.Context, prompt string) (*GenerateResponse, error) {
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				genai.NewPartFromText(systemPrompt),
			},
		},
		Temperature: genai.Ptr(float32(0.7)),
		MaxOutputTokens: 4096,
	}

	// Attempt generation with one retry on failure
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := s.client.Models.GenerateContent(ctx, s.model, []*genai.Content{genai.NewContentFromText(prompt, "user")}, config)
		if err != nil {
			lastErr = fmt.Errorf("gemini API call failed: %w", err)
			log.Printf("Attempt %d failed: %v", attempt+1, lastErr)
			continue
		}

		if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			lastErr = fmt.Errorf("empty response from Gemini")
			log.Printf("Attempt %d: empty response", attempt+1)
			continue
		}

		// Extract the text from the response
		var textContent string
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				textContent += part.Text
			}
		}

		if textContent == "" {
			lastErr = fmt.Errorf("no text content in response")
			continue
		}

		// Clean up the response — sometimes the LLM wraps it in markdown code fences
		textContent = cleanJSON(textContent)

		// Parse the JSON
		var result GenerateResponse
		if err := json.Unmarshal([]byte(textContent), &result); err != nil {
			lastErr = fmt.Errorf("failed to parse LLM response as JSON: %w\nRaw: %s", err, textContent[:min(len(textContent), 500)])
			log.Printf("Attempt %d: parse error: %v", attempt+1, lastErr)
			continue
		}

		// Validate the response
		if result.Code == "" {
			lastErr = fmt.Errorf("LLM returned empty code")
			continue
		}

		if len(result.ParameterDefs) == 0 {
			lastErr = fmt.Errorf("LLM returned no parameter definitions")
			continue
		}

		if result.Defaults == nil {
			result.Defaults = make(map[string]float64)
		}

		// Ensure all parameters have defaults
		for _, def := range result.ParameterDefs {
			if _, ok := result.Defaults[def.Name]; !ok {
				result.Defaults[def.Name] = (def.Min + def.Max) / 2
			}
		}

		log.Printf("Successfully generated simulation with %d parameters", len(result.ParameterDefs))
		return &result, nil
	}

	return nil, lastErr
}

// cleanJSON strips markdown code fences and extra whitespace from LLM output.
func cleanJSON(s string) string {
	s = strings.TrimSpace(s)

	// Remove ```json ... ``` wrapping
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}

	return s
}
