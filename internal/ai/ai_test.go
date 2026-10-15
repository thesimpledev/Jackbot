package ai

import "testing"

func TestNewDefaultsEmptyModelToGPT6Luna(t *testing.T) {
	s := New("test-key", "", "", nil)
	if s.model != "gpt-6-luna" {
		t.Errorf("model = %q, want %q", s.model, "gpt-6-luna")
	}
}

func TestNewKeepsConfiguredModel(t *testing.T) {
	s := New("test-key", "some-model", "", nil)
	if s.model != "some-model" {
		t.Errorf("model = %q, want %q", s.model, "some-model")
	}
}
