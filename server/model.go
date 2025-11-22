package main

import "time"

type OnboardingState struct {
	UserID         string          `json:"user_id"`
	CompletedSteps map[string]bool `json:"completed_steps"`
	StartedAt      time.Time       `json:"started_at"`
	LastUpdated    time.Time       `json:"last_updated"`
}

const (
	onboardingKVPrefix = "onboarding:user:"
)

var onboardingSteps = []string{
	"accounts",
	"profile",
	"channels",
	"tools",
	"policies",
	"intro",
}

var onboardingStepSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(onboardingSteps))
	for _, step := range onboardingSteps {
		m[step] = struct{}{}
	}
	return m
}()
