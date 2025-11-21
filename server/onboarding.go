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

	welcomeMsg := fmt.Sprintf(
		"👋 Hi %s, welcome to %s!\n\n"+
			"I’m your onboarding assistant. I’ll guide you through a few quick steps to get set up.\n\n"+
			"_You can come back to this DM anytime to see your progress._",
		displayName,
		teamName,
	)

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

	return []*model.SlackAttachment{
		{
			Title: "Step 1: Update your profile",
			Text:  checkbox(isDone("profile")) + " Add your photo, job title, and timezone.",
			Actions: []*model.PostAction{
				{
					Name: "Open Profile",
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						// For now, just mark complete. You can later actually link to profile URL.
						URL: callbackURL,
						Context: map[string]interface{}{
							"step": "profile",
						},
					},
				},
			},
		},
		{
			Title: "Step 2: Join key channels",
			Text:  checkbox(isDone("channels")) + " Make sure you’re in #announcements and #helpdesk.",
			Actions: []*model.PostAction{
				{
					Name: "Mark Step Complete",
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
			Title: "Step 3: Read the handbook",
			Text:  checkbox(isDone("handbook")) + " Skim the internal handbook to understand how we work.",
			Actions: []*model.PostAction{
				{
					Name: "Open Handbook",
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: callbackURL,
						Context: map[string]interface{}{
							"step": "handbook",
						},
					},
				},
			},
		},
		{
			Title: "Step 4: Enable MFA",
			Text:  checkbox(isDone("mfa")) + " Turn on multi-factor authentication for your account.",
			Actions: []*model.PostAction{
				{
					Name: "Mark MFA Complete",
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: callbackURL,
						Context: map[string]interface{}{
							"step": "mfa",
						},
					},
				},
			},
		},
		{
			Title: "Step 5: Meet your buddy",
			Text:  checkbox(isDone("intro")) + " Schedule a quick intro with your onboarding buddy or manager.",
			Actions: []*model.PostAction{
				{
					Name: "Mark Meeting Scheduled",
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

	// Respond with an ephemeral or updated message
	resp := &model.PostActionIntegrationResponse{
		Update: &model.Post{
			Id:        req.PostId,
			ChannelId: req.ChannelId,
			UserId:    req.UserId,
			Props: map[string]interface{}{
				"attachments": attachments,
			},
		},
		EphemeralText: fmt.Sprintf("Marked step '%s' complete ✔️", step),
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
