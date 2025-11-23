package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

func (p *Plugin) startOnboardingForUser(user *model.User) error {
	// Idempotent: if we already have state, don’t re-start
	existing, err := p.loadState(user.Id)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	state := &OnboardingState{
		UserID:         user.Id,
		CompletedSteps: map[string]bool{},
		StartedAt:      time.Now().UTC(),
	}
	if err := p.saveState(state); err != nil {
		return err
	}

	// Open DM channel between bot and user
	channel, appErr := p.API.GetDirectChannel(p.botUserID, user.Id)
	if appErr != nil {
		return appErr
	}

	displayName := user.GetFullName()
	if displayName == "" {
		displayName = user.Username
	}

	teamName := p.lookupPrimaryTeamName(user)

	// Get translations
	tr := p.getTranslations()

	welcomeMsg := fmt.Sprintf(tr.WelcomeGreeting, displayName, teamName) + "\n\n" +
		tr.WelcomeIntro + "\n\n" +
		tr.WelcomeClosing

	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channel.Id,
		Message:   welcomeMsg,
		Props: map[string]interface{}{
			"attachments": p.buildChecklistAttachments(state),
		},
	}

	if _, appErr := p.API.CreatePost(post); appErr != nil {
		return appErr
	}

	return nil
}

func (p *Plugin) buildChecklistAttachments(state *OnboardingState) []*model.SlackAttachment {
	pluginURL, err := p.pluginURL()
	if err != nil {
		p.API.LogError("pluginURL not configured", "err", err.Error())
		return []*model.SlackAttachment{}
	}
	callbackURL := pluginURL + "/complete-step"

	isDone := func(step string) bool {
		return state.CompletedSteps[step]
	}

	// Get translations
	tr := p.getTranslations()

	return []*model.SlackAttachment{
		{
			Title: tr.Step1Title,
			Text:  checkbox(isDone("accounts")) + " " + tr.Step1Description + tr.Step1Link,
			Actions: []*model.PostAction{
				{
					Name: tr.ButtonMarkAccountsReady,
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: callbackURL,
						Context: map[string]interface{}{
							"step": "accounts",
						},
					},
				},
			},
		},
		{
			Title: tr.Step2Title,
			Text:  checkbox(isDone("profile")) + " " + tr.Step2Description + tr.Step2Link,
			Actions: []*model.PostAction{
				{
					Name: tr.ButtonGenerateSignature,
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: callbackURL,
						Context: map[string]interface{}{
							"action": "open_signature_dialog",
						},
					},
				},
				{
					Name: tr.ButtonMarkProfileComplete,
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: callbackURL,
						Context: map[string]interface{}{
							"step": "profile",
						},
					},
				},
			},
		},
		{
			Title: tr.Step3Title,
			Text:  checkbox(isDone("channels")) + " " + tr.Step3Description + tr.Step3Link,
			Actions: []*model.PostAction{
				{
					Name: tr.ButtonMarkChannelsJoined,
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: callbackURL,
						Context: map[string]interface{}{
							"step": "channels",
						},
					},
				},
			},
		},
		{
			Title: tr.Step4Title,
			Text:  checkbox(isDone("tools")) + " " + tr.Step4Description + tr.Step4Link,
			Actions: []*model.PostAction{
				{
					Name: tr.ButtonMarkToolsReady,
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: callbackURL,
						Context: map[string]interface{}{
							"step": "tools",
						},
					},
				},
			},
		},
		{
			Title: tr.Step5Title,
			Text:  checkbox(isDone("policies")) + " " + tr.Step5Description + tr.Step5Link,
			Actions: []*model.PostAction{
				{
					Name: tr.ButtonMarkPoliciesReviewed,
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: callbackURL,
						Context: map[string]interface{}{
							"step": "policies",
						},
					},
				},
			},
		},
		{
			Title: tr.Step6Title,
			Text:  checkbox(isDone("intro")) + " " + tr.Step6Description + tr.Step6Link,
			Actions: []*model.PostAction{
				{
					Name: tr.ButtonMarkIntrosDone,
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: callbackURL,
						Context: map[string]interface{}{
							"step": "intro",
						},
					},
				},
			},
		},
	}
}

func checkbox(done bool) string {
	if done {
		return "✅ "
	}
	return "☐ "
}

// Handle integration callback when user clicks a button
func (p *Plugin) handleCompleteStep(w http.ResponseWriter, r *http.Request) {
	var req model.PostActionIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		p.API.LogError("failed to decode integration request", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := req.UserId

	// Check if this is a signature generator action
	if actionRaw, ok := req.Context["action"]; ok {
		action, _ := actionRaw.(string)
		if action == "open_signature_dialog" {
			p.handleSignatureDialog(w, r, &req)
			return
		}
	}

	stepRaw, ok := req.Context["step"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	step, _ := stepRaw.(string)

	if !isAllowedStep(step) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	state, err := p.loadState(userID)
	if err != nil {
		p.API.LogError("failed to load onboarding state", "user_id", userID, "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if state == nil {
		state = &OnboardingState{
			UserID:         userID,
			CompletedSteps: map[string]bool{},
		}
	}

	state.CompletedSteps[step] = true
	if err := p.saveState(state); err != nil {
		p.API.LogError("failed to save onboarding state", "user_id", userID, "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Rebuild attachments to reflect updated checkboxes
	attachments := p.buildChecklistAttachments(state)

	// Get user info to rebuild welcome message
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		p.API.LogError("failed to get user", "err", appErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	displayName := user.GetFullName()
	if displayName == "" {
		displayName = user.Username
	}

	teamName := p.lookupPrimaryTeamName(user)

	// Get translations
	tr := p.getTranslations()

	// Rebuild welcome message
	welcomeMsg := fmt.Sprintf(tr.WelcomeGreeting, displayName, teamName) + "\n\n" +
		tr.WelcomeIntro + "\n\n" +
		tr.WelcomeClosing

	// Respond with updated message that includes welcome text
	resp := &model.PostActionIntegrationResponse{
		Update: &model.Post{
			Id:        req.PostId,
			ChannelId: req.ChannelId,
			UserId:    req.UserId,
			Message:   welcomeMsg,
			Props: map[string]interface{}{
				"attachments": attachments,
			},
		},
		EphemeralText: fmt.Sprintf(tr.StepMarkedComplete, step),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		p.API.LogError("failed to encode integration response", "err", err.Error())
	}
}

func isAllowedStep(step string) bool {
	_, ok := onboardingStepSet[step]
	return ok
}

func (p *Plugin) lookupPrimaryTeamName(user *model.User) string {
	memberships, appErr := p.API.GetTeamsForUser(user.Id)
	if appErr != nil || len(memberships) == 0 {
		return "Mattermost"
	}
	// Use the first team; in most orgs there is one primary team.
	return memberships[0].DisplayName
}
