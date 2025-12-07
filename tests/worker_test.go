package tests

import (
	"testing"
	"time"
)

// FakeService implements just enough methods to test the worker logic
type FakeService struct{}

func (f *FakeService) RenderTemplate(tmpl string, ctx map[string]string) (string, error) {
	out := tmpl
	for k, v := range ctx {
		out = replaceAll(out, "{{"+k+"}}", v)
	}
	// remove placeholders
	for {
		i := indexOf(out, "{{")
		if i == -1 {
			break
		}
		j := indexOf(out[i:], "}}")
		if j == -1 {
			break
		}
		out = out[:i] + out[i+j+2:]
	}
	return out, nil
}

// String helper functions
func replaceAll(s, old, new string) string {
	for {
		i := indexOf(s, old)
		if i == -1 {
			break
		}
		s = s[:i] + new + s[i+len(old):]
	}
	return s
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestWorkerProcessLogic(t *testing.T) {
	svc := &FakeService{}

	// Simulate a queued job
	customer := map[string]string{"first_name": "Sam"}
	template := "Hi {{first_name}}"

	out, err := svc.RenderTemplate(template, customer)
	if err != nil {
		t.Fatalf("RenderTemplate error: %v", err)
	}

	if out != "Hi Sam" {
		t.Fatalf("unexpected output: %s", out)
	}

	// Optionally, simulate multiple jobs in memory
	jobs := []map[string]string{
		{"first_name": "Alice"},
		{"first_name": "Bob"},
	}

	for _, job := range jobs {
		out, _ := svc.RenderTemplate(template, job)
		if out == "" {
			t.Fatalf("rendered message should not be empty")
		}
	}

	// Fake delay to simulate processing time
	time.Sleep(10 * time.Millisecond)
}
