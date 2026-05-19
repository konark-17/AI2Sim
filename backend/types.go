package main

// GenerateRequest is the JSON body sent by the frontend.
type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

// ParameterDef describes a single adjustable slider for the simulation.
type ParameterDef struct {
	Name  string  `json:"name"`
	Label string  `json:"label"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Step  float64 `json:"step"`
}

// GenerateResponse is the JSON payload returned to the frontend.
type GenerateResponse struct {
	Code          string             `json:"code"`
	ParameterDefs []ParameterDef     `json:"parameterDefs"`
	Defaults      map[string]float64 `json:"defaults"`
}

// ErrorResponse is returned on failures.
type ErrorResponse struct {
	Error string `json:"error"`
}
