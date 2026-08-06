package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestIsValidServiceName(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple lowercase", "payments", true},
		{"kebab-case", "order-events-consumer", true},
		{"alphanumeric with hyphen", "service-2", true},
		{"empty string rejected", "", false},
		{"uppercase rejected", "Payments-Api", false},
		{"spaces rejected", "payments api", false},
		{"underscore rejected", "payments_api", false},
		{"special characters rejected", "payments@api!", false},
		{"leading hyphen rejected", "-payments", false},
		{"trailing hyphen rejected", "payments-", false},
		{"double hyphen rejected", "payments--api", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := isValidServiceName(c.input)
			if got != c.want {
				t.Errorf("isValidServiceName(%q) = %v, want %v", c.input, got, c.want)
			}
		})
	}
}

func TestPromptServiceNameRePromptsOnInvalidInput(t *testing.T) {
	input := "Payments Api\npayments-api\n"
	r := bufio.NewReader(strings.NewReader(input))
	var out bytes.Buffer

	got, err := promptServiceName(r, &out)
	if err != nil {
		t.Fatalf("promptServiceName returned error: %v", err)
	}
	if got != "payments-api" {
		t.Errorf("promptServiceName() = %q, want %q", got, "payments-api")
	}
	if !strings.Contains(out.String(), "Invalid service name") {
		t.Errorf("expected re-prompt message for invalid input, got output: %q", out.String())
	}
}
