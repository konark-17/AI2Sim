package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// BuildSystemPrompt returns the master system prompt for the LLM.
// It instructs the model to produce only valid, self-contained canvas JavaScript.
func BuildSystemPrompt() string {
	return `You are AI2Sim — a scientific simulation code generator.

OUTPUT ONLY raw JavaScript. NO markdown. NO code fences. NO backticks. NO explanations.

Define exactly ONE function named "createSimulation" with this exact signature:

function createSimulation(canvas, params) {
  var ctx = canvas.getContext('2d');
  var animId;
  var gravity = (params && params.gravity !== undefined) ? params.gravity : 9.8;
  // ... extract all params with defaults like above ...

  function draw() {
    // clear + draw frame
    animId = requestAnimationFrame(draw);
  }

  draw();

  return {
    stop: function() { if (animId) cancelAnimationFrame(animId); }
  };
}

RULES (violating any will break the app):
- Start IMMEDIATELY with "function createSimulation" — no text before it
- End IMMEDIATELY after the closing "}" — no text after it
- Extract EVERY param at the top using: var x = (params && params.x !== undefined) ? params.x : DEFAULT;
- Use var (not const/let) for variables that need to persist across animation frames
- Use a dark background (#07070f) and bright neon colors (#8a63ff, #3ecfff, #ff6b9d)
- Write COMPACT code — no verbose comments, no unnecessary whitespace
- No import, require, fetch, WebSocket, or external dependencies
- The function must call requestAnimationFrame internally and return { stop: fn }
- Do NOT use template literals (backtick strings)`
}

// ParameterSchema describes a single simulation parameter.
type ParameterSchema struct {
	Name        string  `json:"name"`
	Label       string  `json:"label"`
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	Default     float64 `json:"default"`
	Step        float64 `json:"step"`
	Unit        string  `json:"unit"`
}

// SimulationCategory groups prompt examples and parameter defaults.
type SimulationCategory string

const (
	CategoryPhysics   SimulationCategory = "physics"
	CategoryMath      SimulationCategory = "math"
	CategoryBiology   SimulationCategory = "biology"
	CategoryChemistry SimulationCategory = "chemistry"
	CategoryAuto      SimulationCategory = "auto"
)

// BuildUserMessage constructs the user turn of the LLM conversation.
// It includes the user's description and the current parameter values.
func BuildUserMessage(userPrompt string, params map[string]any) string {
	var sb strings.Builder

	sb.WriteString("## Simulation Request\n")
	sb.WriteString(userPrompt)
	sb.WriteString("\n\n")

	if len(params) > 0 {
		paramsJSON, err := json.MarshalIndent(params, "", "  ")
		if err == nil {
			sb.WriteString("## Parameter Values\n")
			sb.WriteString("The 'params' object passed to createSimulation will contain:\n")
			sb.WriteString("```json\n")
			sb.Write(paramsJSON)
			sb.WriteString("\n```\n\n")
			sb.WriteString("Use EXACTLY these parameter names inside your createSimulation function.\n")
		}
	}

	sb.WriteString("\n## Output Instructions\n")
	sb.WriteString("Output the JavaScript function NOW. Start directly with `function createSimulation`.")

	return sb.String()
}

// DefaultParamsForPrompt attempts to infer sensible default parameters
// from the user's natural language prompt. Used when no params are provided.
func DefaultParamsForPrompt(prompt string) []ParameterSchema {
	lower := strings.ToLower(prompt)

	switch {
	case containsAny(lower, "pendulum", "swing", "oscillat"):
		return pendulumParams()
	case containsAny(lower, "particle", "gravity", "orbit", "planet", "solar"):
		return particleParams()
	case containsAny(lower, "wave", "ripple", "interference"):
		return waveParams()
	case containsAny(lower, "fractal", "mandelbrot", "julia", "attractor", "chaos"):
		return fractalParams()
	case containsAny(lower, "population", "predator", "prey", "boid", "flocking"):
		return populationParams()
	case containsAny(lower, "reaction", "diffusion", "chemical", "turing"):
		return reactionDiffusionParams()
	case containsAny(lower, "gas", "pressure", "brownian", "molecule"):
		return gasParams()
	default:
		return genericParams()
	}
}

func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// Parameter presets for common simulation types

func pendulumParams() []ParameterSchema {
	return []ParameterSchema{
		{Name: "gravity", Label: "Gravity", Min: 0.5, Max: 20, Default: 9.8, Step: 0.1, Unit: "m/s²"},
		{Name: "length1", Label: "Arm 1 Length", Min: 50, Max: 250, Default: 150, Step: 1, Unit: "px"},
		{Name: "length2", Label: "Arm 2 Length", Min: 50, Max: 250, Default: 120, Step: 1, Unit: "px"},
		{Name: "mass1", Label: "Mass 1", Min: 5, Max: 40, Default: 15, Step: 0.5, Unit: "kg"},
		{Name: "mass2", Label: "Mass 2", Min: 5, Max: 40, Default: 10, Step: 0.5, Unit: "kg"},
		{Name: "damping", Label: "Damping", Min: 0, Max: 0.05, Default: 0.001, Step: 0.001, Unit: ""},
		{Name: "trailLength", Label: "Trail Length", Min: 0, Max: 2000, Default: 500, Step: 10, Unit: "pts"},
	}
}

func particleParams() []ParameterSchema {
	return []ParameterSchema{
		{Name: "particleCount", Label: "Particles", Min: 10, Max: 500, Default: 100, Step: 5, Unit: ""},
		{Name: "gravity", Label: "Gravity", Min: -5, Max: 5, Default: 0.1, Step: 0.05, Unit: ""},
		{Name: "speed", Label: "Initial Speed", Min: 0.5, Max: 10, Default: 2, Step: 0.1, Unit: ""},
		{Name: "bounce", Label: "Elasticity", Min: 0, Max: 1, Default: 0.85, Step: 0.01, Unit: ""},
		{Name: "radius", Label: "Particle Size", Min: 2, Max: 20, Default: 5, Step: 0.5, Unit: "px"},
	}
}

func waveParams() []ParameterSchema {
	return []ParameterSchema{
		{Name: "frequency", Label: "Frequency", Min: 0.5, Max: 10, Default: 2, Step: 0.1, Unit: "Hz"},
		{Name: "amplitude", Label: "Amplitude", Min: 10, Max: 200, Default: 80, Step: 1, Unit: "px"},
		{Name: "speed", Label: "Wave Speed", Min: 0.5, Max: 10, Default: 3, Step: 0.1, Unit: ""},
		{Name: "sources", Label: "Wave Sources", Min: 1, Max: 5, Default: 2, Step: 1, Unit: ""},
		{Name: "damping", Label: "Damping", Min: 0, Max: 0.1, Default: 0.01, Step: 0.001, Unit: ""},
	}
}

func fractalParams() []ParameterSchema {
	return []ParameterSchema{
		{Name: "maxIterations", Label: "Max Iterations", Min: 20, Max: 500, Default: 100, Step: 5, Unit: ""},
		{Name: "zoom", Label: "Zoom", Min: 0.5, Max: 5, Default: 1, Step: 0.05, Unit: "×"},
		{Name: "colorCycles", Label: "Color Cycles", Min: 1, Max: 20, Default: 5, Step: 0.5, Unit: ""},
		{Name: "animationSpeed", Label: "Animation Speed", Min: 0, Max: 0.02, Default: 0.003, Step: 0.001, Unit: ""},
	}
}

func populationParams() []ParameterSchema {
	return []ParameterSchema{
		{Name: "preyCount", Label: "Initial Prey", Min: 10, Max: 200, Default: 80, Step: 5, Unit: ""},
		{Name: "predatorCount", Label: "Initial Predators", Min: 5, Max: 50, Default: 20, Step: 1, Unit: ""},
		{Name: "preySpeed", Label: "Prey Speed", Min: 0.5, Max: 5, Default: 1.5, Step: 0.1, Unit: ""},
		{Name: "predatorSpeed", Label: "Predator Speed", Min: 0.5, Max: 5, Default: 2, Step: 0.1, Unit: ""},
		{Name: "visionRange", Label: "Vision Range", Min: 20, Max: 200, Default: 80, Step: 5, Unit: "px"},
		{Name: "reproductionRate", Label: "Prey Reproduction", Min: 0.001, Max: 0.05, Default: 0.01, Step: 0.001, Unit: ""},
	}
}

func reactionDiffusionParams() []ParameterSchema {
	return []ParameterSchema{
		{Name: "feedRate", Label: "Feed Rate (F)", Min: 0.01, Max: 0.1, Default: 0.055, Step: 0.001, Unit: ""},
		{Name: "killRate", Label: "Kill Rate (k)", Min: 0.04, Max: 0.07, Default: 0.062, Step: 0.001, Unit: ""},
		{Name: "diffuseA", Label: "Diffusion A", Min: 0.5, Max: 1.5, Default: 1.0, Step: 0.05, Unit: ""},
		{Name: "diffuseB", Label: "Diffusion B", Min: 0.1, Max: 0.8, Default: 0.5, Step: 0.01, Unit: ""},
		{Name: "speed", Label: "Simulation Speed", Min: 1, Max: 20, Default: 5, Step: 1, Unit: "steps/frame"},
	}
}

func gasParams() []ParameterSchema {
	return []ParameterSchema{
		{Name: "particleCount", Label: "Molecule Count", Min: 10, Max: 300, Default: 100, Step: 5, Unit: ""},
		{Name: "temperature", Label: "Temperature", Min: 100, Max: 2000, Default: 500, Step: 50, Unit: "K"},
		{Name: "mass", Label: "Molecular Mass", Min: 1, Max: 100, Default: 20, Step: 1, Unit: "amu"},
		{Name: "radius", Label: "Molecule Radius", Min: 3, Max: 15, Default: 6, Step: 0.5, Unit: "px"},
	}
}

func genericParams() []ParameterSchema {
	return []ParameterSchema{
		{Name: "speed", Label: "Speed", Min: 0.1, Max: 10, Default: 1, Step: 0.1, Unit: ""},
		{Name: "count", Label: "Count", Min: 1, Max: 500, Default: 50, Step: 1, Unit: ""},
		{Name: "intensity", Label: "Intensity", Min: 0.1, Max: 5, Default: 1, Step: 0.1, Unit: ""},
		{Name: "scale", Label: "Scale", Min: 0.1, Max: 5, Default: 1, Step: 0.05, Unit: ""},
	}
}

// ParamsToMap converts a schema slice to a flat map of name→default for the LLM.
func ParamsToMap(schema []ParameterSchema) map[string]any {
	m := make(map[string]any, len(schema))
	for _, p := range schema {
		m[p.Name] = p.Default
	}
	return m
}

// ValidateParams ensures all incoming param values are within the schema's bounds.
func ValidateParams(incoming map[string]any, schema []ParameterSchema) map[string]any {
	validated := make(map[string]any, len(schema))
	schemaMap := make(map[string]ParameterSchema, len(schema))
	for _, s := range schema {
		schemaMap[s.Name] = s
	}

	for k, v := range incoming {
		if s, ok := schemaMap[k]; ok {
			if fv, ok := toFloat(v); ok {
				if fv < s.Min {
					fv = s.Min
				} else if fv > s.Max {
					fv = s.Max
				}
				validated[k] = fv
			}
		}
	}

	// Fill in defaults for missing params
	for _, s := range schema {
		if _, ok := validated[s.Name]; !ok {
			validated[s.Name] = s.Default
		}
	}

	return validated
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// DescribeParams returns a human-readable description of simulation parameters
// for inclusion in the prompt.
func DescribeParams(schema []ParameterSchema) string {
	if len(schema) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, p := range schema {
		sb.WriteString(fmt.Sprintf("  - %s (%s): %.4g to %.4g, default %.4g %s\n",
			p.Name, p.Label, p.Min, p.Max, p.Default, p.Unit))
	}
	return sb.String()
}
