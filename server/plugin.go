package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin

	botUserID string
}

const botUserKVKey = "onboarding:bot_user_id"
const botUsername = "eoto-onboarding-bot"
const botDisplayName = "EOTO Onboarding Helper"
const botDescription = "Guides new teammates through onboarding."
const botIconPath = "assets/icon.png"

// OnActivate runs when the plugin is enabled.
func (p *Plugin) OnActivate() error {
	if err := p.ensureBotUser(); err != nil {
		return err
	}

	p.API.LogInfo("Onboarding plugin activated", "bot_user_id", p.botUserID)
	return nil
}

func (p *Plugin) ensureBotUser() error {
	if p.botUserID != "" {
		p.ensureBotProfile()
		p.ensureBotIcon()
		return nil
	}

	if storedID, err := p.loadBotUserID(); err == nil && storedID != "" {
		p.botUserID = storedID
		p.ensureBotProfile()
		p.ensureBotIcon()
		return nil
	} else if err != nil {
		return err
	}

	bot := &model.Bot{
		Username:    botUsername,
		DisplayName: botDisplayName,
		Description: botDescription,
	}

	botUser, appErr := p.API.GetBot(bot.Username, true)
	if appErr == nil {
		if botUser == nil {
			return fmt.Errorf("GetBot returned no bot and no error")
		}
		p.botUserID = botUser.UserId
		if err := p.saveBotUserID(p.botUserID); err != nil {
			return err
		}
		p.ensureBotProfile()
		p.ensureBotIcon()
		return nil
	}
	if appErr.StatusCode != http.StatusNotFound {
		return appErr
	}

	createdBot, appErr := p.API.CreateBot(bot)
	if appErr != nil {
		// Username may already be taken by a regular user; retry with a unique suffix.
		bot.Username = fmt.Sprintf("%s-%d", bot.Username, time.Now().Unix())
		createdBot, appErr = p.API.CreateBot(bot)
		if appErr != nil {
			return appErr
		}
	}

	p.botUserID = createdBot.UserId
	if err := p.saveBotUserID(p.botUserID); err != nil {
		return err
	}
	p.ensureBotProfile()
	p.ensureBotIcon()
	return nil
}

// UserHasBeenCreated is called when a new user is created.
func (p *Plugin) UserHasBeenCreated(c *plugin.Context, user *model.User) {
	// Ignore bots
	if user.IsBot {
		return
	}

	if err := p.startOnboardingForUser(user); err != nil {
		p.API.LogError("Failed to start onboarding", "user_id", user.Id, "err", err.Error())
	}
}

// ServeHTTP handles interactive button callbacks from posts.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	// Only handle POST integrations from interactive message actions
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {
	case "/complete-step":
		p.handleCompleteStep(w, r)
	case "/submit-signature":
		p.handleSignatureSubmission(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// Helper to build plugin URL for integration callbacks.
func (p *Plugin) pluginURL() (string, error) {
	cfg := p.API.GetConfig()
	if cfg.ServiceSettings.SiteURL == nil || *cfg.ServiceSettings.SiteURL == "" {
		return "", fmt.Errorf("siteURL is not configured")
	}

	parsed, err := url.Parse(*cfg.ServiceSettings.SiteURL)
	if err != nil {
		return "", fmt.Errorf("parse siteURL: %w", err)
	}

	resolved, err := url.JoinPath(parsed.String(), "plugins", "com.akinlosotutech.onboardinghelper")
	if err != nil {
		return "", fmt.Errorf("join plugin path: %w", err)
	}

	return resolved, nil
}

// KV helpers

func (p *Plugin) loadState(userID string) (*OnboardingState, error) {
	key := onboardingKVPrefix + userID
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return nil, fmt.Errorf("KVGet: %w", appErr)
	}
	if data == nil {
		return nil, nil
	}

	var state OnboardingState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (p *Plugin) saveState(state *OnboardingState) error {
	state.LastUpdated = time.Now().UTC()
	if state.StartedAt.IsZero() {
		state.StartedAt = state.LastUpdated
	}

	b, err := json.Marshal(state)
	if err != nil {
		return err
	}

	key := onboardingKVPrefix + state.UserID
	if appErr := p.API.KVSet(key, b); appErr != nil {
		return appErr
	}
	return nil
}

func (p *Plugin) loadBotUserID() (string, error) {
	data, appErr := p.API.KVGet(botUserKVKey)
	if appErr != nil {
		return "", appErr
	}
	if data == nil {
		return "", nil
	}
	return string(data), nil
}

func (p *Plugin) saveBotUserID(id string) error {
	return p.API.KVSet(botUserKVKey, []byte(id))
}

func (p *Plugin) ensureBotProfile() {
	if p.botUserID == "" {
		return
	}
	display := botDisplayName
	desc := botDescription
	patch := &model.BotPatch{
		DisplayName: &display,
		Description: &desc,
	}
	if _, appErr := p.API.PatchBot(p.botUserID, patch); appErr != nil {
		p.API.LogWarn("failed to patch bot profile", "err", appErr.Error())
	}
}

func (p *Plugin) ensureBotIcon() {
	if p.botUserID == "" {
		return
	}

	bundlePath, appErr := p.API.GetBundlePath()
	if appErr != nil {
		p.API.LogWarn("failed to get bundle path for bot icon", "err", appErr.Error())
		return
	}

	iconPath := filepath.Join(bundlePath, botIconPath)
	data, err := os.ReadFile(iconPath)
	if err != nil {
		p.API.LogDebug("bot icon not set; file not found", "path", iconPath, "err", err.Error())
		return
	}

	if appErr := p.API.SetProfileImage(p.botUserID, data); appErr != nil {
		p.API.LogWarn("failed to set bot icon", "err", appErr.Error())
	}
}
