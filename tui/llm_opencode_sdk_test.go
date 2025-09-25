//go:build !integration

package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestLLMInterface(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		setupLLM func() LLM
		model    string
		prompt   Prompt
		wantResp string
		wantErr  error
		tuiSim   bool // For integration sim
	}{
		{
			name:     "Happy path: NewSession and single-turn chat",
			setupLLM: func() LLM { return &MockLLM{Responses: []string{"Mock response: Hello!"}} },
			model:    "gpt-3.5",
			prompt:   Prompt{System: "You are helpful.", User: "Hi"},
			wantResp: "Mock response: Hello!",
			wantErr:  nil,
		},
		{
			name:     "Error: Invalid model",
			setupLLM: func() LLM { return &MockLLM{Responses: []string{""}} },
			model:    "invalid",
			prompt:   Prompt{User: "Test"},
			wantResp: "",
			wantErr:  ErrInvalidModel,
		},
		{
			name:     "Multi-turn chat",
			setupLLM: func() LLM { return &MockLLM{Responses: []string{"First", "Second"}} },
			model:    "llama2",
			prompt:   Prompt{User: "First query"},
			wantResp: "First",
			wantErr:  nil,
		},
		{
			name:     "Timeout error",
			setupLLM: func() LLM { return &MockLLM{Errors: []error{ErrTimeout}} },
			model:    "gpt-4",
			prompt:   Prompt{User: "Slow query"},
			wantResp: "",
			wantErr:  ErrTimeout,
		},
		{
			name:     "Empty prompt",
			setupLLM: func() LLM { return &MockLLM{Responses: []string{"No input"}} },
			model:    "default",
			prompt:   Prompt{},
			wantResp: "",
			wantErr:  ErrNoResponse,
		},
		// Integration: TUI F2 switching simulation
		{
			name:     "E2E Sim: Chat with F2 toggle (LLM on)",
			setupLLM: func() LLM { return &MockLLM{Responses: []string{"TUI response"}} },
			model:    "enabled",
			prompt:   Prompt{User: "Chat msg"},
			wantResp: "TUI response",
			wantErr:  nil,
			tuiSim:   true,
		},
		{
			name:     "E2E Sim: F2 toggle disables LLM (fallback to echo)",
			setupLLM: func() LLM { return &MockLLM{Responses: []string{"Disabled"}} },
			model:    "disabled", // Simulates F2 off
			prompt:   Prompt{User: "Fallback"},
			wantResp: "Echo: Fallback", // Mock fallback
			wantErr:  nil,
			tuiSim:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := tt.setupLLM()
			session, err := llm.NewSession(ctx, tt.model)
			if tt.name == "Error: Invalid model" {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
				}
				if err != nil {
					return
				}
			} else {
				if err != nil {
					t.Fatalf("NewSession() unexpected error = %v", err)
				}
			}

			// Append initial message
			if err := session.AppendMessage("user", tt.prompt.User); err != nil {
				t.Fatal(err)
			}

			resp, err := llm.Generate(ctx, session, tt.prompt)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(resp, tt.wantResp) {
				t.Errorf("Generate() = %q, want %q", resp, tt.wantResp)
			}

			// TUI Sim: Mock F2 switch (integration)
			if tt.tuiSim {
				t.Run("F2SwitchSim", func(t *testing.T) {
					// Simulate F2 keypress: Toggle model setting
					toggledModel := toggleF2Setting(tt.model) // Mock func: e.g., if "enabled" -> "disabled"
					newSession, _ := llm.NewSession(ctx, toggledModel)
					fallbackResp, _ := llm.Generate(ctx, newSession, tt.prompt)
					if toggledModel == "disabled" && !strings.Contains(fallbackResp, "Echo:") {
						t.Error("F2 toggle should fallback to echo mode")
					}
				})
			}
		})
	}
}

// Mock implementations
type MockLLM struct {
	Responses []string
	Errors    []error
	idx       int
}

func (m *MockLLM) NewSession(ctx context.Context, model string) (Session, error) {
	if model == "invalid" {
		return nil, ErrInvalidModel
	}
	return &MockSession{model: model}, nil
}

func (m *MockLLM) Generate(ctx context.Context, session Session, prompt Prompt) (string, error) {
	if s, ok := session.(*MockSession); ok && s.model == "disabled" {
		return "Echo: " + prompt.User, nil
	}
	if m.idx < len(m.Errors) {
		defer func() { m.idx++ }()
		return "", m.Errors[m.idx]
	}
	if prompt.User == "" {
		return "", ErrNoResponse
	}
	defer func() { m.idx++ }()
	if m.idx < len(m.Responses) {
		return m.Responses[m.idx], nil
	}
	return "Default mock", nil
}

type MockSession struct {
	history []LLMMessage
	model   string
}

func (s *MockSession) AppendMessage(role, content string) error {
	s.history = append(s.history, LLMMessage{Role: role, Content: content})
	return nil
}

func (s *MockSession) GetHistory() []LLMMessage {
	return s.history
}

// Mock TUI helper (simulates F2)
func toggleF2Setting(model string) string {
	if model == "enabled" {
		return "disabled"
	}
	return "enabled" // Or echo fallback
}

func TestMockSession(t *testing.T) {
	s := &MockSession{}
	s.AppendMessage("user", "test")
	hist := s.GetHistory()
	if len(hist) != 1 || hist[0].Content != "test" {
		t.Error("MockSession append failed")
	}
}
