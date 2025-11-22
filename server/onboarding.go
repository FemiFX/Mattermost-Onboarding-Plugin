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
			Title: "Step 1: Accounts & Access",
			Text: checkbox(isDone("accounts")) + " Make sure you can sign in everywhere you need:\n" +
				"- Google Workspace (EOTO email address issued & tested)\n" +
				"- Nextcloud (files & shared team folders)\n" +
				"- Timebutler (time tracking / attendance)\n" +
				"- Mattermost (you’re here 🎉)\n" +
				"- Any role-specific tools (e.g. CRM, finance tools)\n\n" +
				"More details: [Accounts & Access Guide](https://outline.akinlosotu.tech)",
			Actions: []*model.PostAction{
				{
					Name: "Mark Accounts Ready",
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
			Title: "Step 2: Complete Your Profile",
			Text: checkbox(isDone("profile")) + " Help colleagues recognise and reach you easily:\n" +
				"- Upload a clear profile photo\n" +
				"- Add your full name and pronouns (if you wish)\n" +
				"- Set your job title & department\n" +
				"- Set your timezone and working hours\n\n" +
				"Quick reference: [Mattermost Profile & Notifications](https://outline.akinlosotu.tech)",
			Actions: []*model.PostAction{
				{
					Name: "Mark Profile Complete",
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
			Title: "Step 3: Communication Channels",
			Text: checkbox(isDone("channels")) + " Join the spaces where information flows:\n" +
				"- `#announcements` — organisation-wide updates\n" +
				"- `#helpdesk` — IT support & quick questions\n" +
				"- `#introductions` — say hello to everyone\n" +
				"- Your team / project channels (ask your manager)\n\n" +
				"Guidelines: [Communication & Channels](https://outline.akinlosotu.tech)",
			Actions: []*model.PostAction{
				{
					Name: "Mark Channels Joined",
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
			Title: "Step 4: Tools & Equipment",
			Text: checkbox(isDone("tools")) + " Confirm your hardware and core tools are ready:\n" +
				"- Laptop received, boots correctly, and you can log in\n" +
				"- Wi-Fi access at your usual working location(s)\n" +
				"- Nextcloud client installed (if needed)\n" +
				"- Email & calendar working on your primary device\n" +
				"- Any required VPN or remote access configured\n\n" +
				"See: [Devices & IT Setup](https://outline.akinlosotu.tech)",
			Actions: []*model.PostAction{
				{
					Name: "Mark Tools Ready",
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
			Title: "Step 5: Working Practices & Policies",
			Text: checkbox(isDone("policies")) + " Take a first pass through how we work at EOTO:\n" +
				"- Working hours, flex-time, and time-off process\n" +
				"- Data protection & privacy basics (GDPR awareness)\n" +
				"- Communication expectations (response times, DM vs channels)\n" +
				"- How we store and share files (Nextcloud structure)\n\n" +
				"Start here: [EOTO Handbook](https://outline.akinlosotu.tech)",
			Actions: []*model.PostAction{
				{
					Name: "Mark Policies Reviewed",
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
			Title: "Step 6: People & Check-ins",
			Text: checkbox(isDone("intro")) + " Make sure you’re connected to the right humans:\n" +
				"- Short introduction post in `#introductions`\n" +
				"- 1:1 intro with your manager (scheduled)\n" +
				"- Check-in with your onboarding buddy (if assigned)\n" +
				"- Add key people to your favorites in Mattermost\n\n" +
				"Tips: [Onboarding & Collaboration at EOTO](https://outline.akinlosotu.tech)",
			Actions: []*model.PostAction{
				{
					Name: "Mark Intros Done",
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
