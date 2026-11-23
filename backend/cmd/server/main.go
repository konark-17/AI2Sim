package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"ai2sim/backend/internal/api"
	"ai2sim/backend/internal/llm"
	"ai2sim/backend/internal/middleware"
)

func main() {
	loadDotEnv()

	// Initialize LLM client (auto-detects Gemini, Groq, OpenAI, or Ollama)
	cfg := llm.Config{
		Provider:      getEnv("LLM_PROVIDER", "auto"),
		BaseURL:       getEnv("LLM_BASE_URL", ""),
		Model:         getEnv("LLM_MODEL", ""),
		APIKey:        getEnv("LLM_API_KEY", ""),
		MaxConcurrent: 5,
	}

	llmClient, err := llm.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize LLM client: %v", err)
	}

	// Warm up model in background if local Ollama
	go llmClient.WarmUp()

	app := fiber.New(fiber.Config{
		AppName:               "AI2Sim Backend v1.1",
		ReadTimeout:           0, // No timeout for streaming
		WriteTimeout:          0,
		IdleTimeout:           0,
		DisableStartupMessage: false,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	// CORS — allow all local dev ports (5173, 5174, 4173) and configured origins
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174,http://localhost:4173,http://127.0.0.1:5173,http://127.0.0.1:5174")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
	}))

	// Concurrency limiter
	app.Use(middleware.ConcurrencyLimiter(llmClient))

	// Register routes
	handler := api.NewHandler(llmClient)
	api.RegisterRoutes(app, handler)

	port := getEnv("PORT", "8080")
	ready, statusMsg := llmClient.Status()
	log.Printf("🚀 AI2Sim backend starting on :%s", port)
	log.Printf("   Provider : %s", strings.ToUpper(llmClient.Provider()))
	log.Printf("   Model    : %s", llmClient.Model())
	log.Printf("   Status   : ready=%v (%s)", ready, statusMsg)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func loadDotEnv() {
	for _, path := range []string{".env", "../.env"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, `"'`)
				if os.Getenv(key) == "" {
					os.Setenv(key, val)
				}
			}
		}
		break
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
