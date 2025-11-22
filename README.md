# Mattermost Onboarding Plugin

A Mattermost plugin that provides automated, interactive onboarding for new team members. When a user is created, the plugin sends them a personalized welcome message via a bot and guides them through a checklist of onboarding tasks with interactive buttons.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [How It Works](#how-it-works)
- [Customization Guide](#customization-guide)
- [Building & Deployment](#building--deployment)
- [Configuration](#configuration)
- [Development Tips](#development-tips)

---

## Overview

This plugin automates the onboarding process for new users joining your Mattermost workspace. It:

1. **Detects new users** automatically when they're created
2. **Creates a dedicated onboarding bot** that DMs the new user
3. **Sends an interactive checklist** with steps covering accounts, profile setup, channels, tools, policies, and introductions
4. **Tracks progress** persistently using Mattermost's key-value store
5. **Updates in real-time** as users click buttons to mark steps complete

---

## Features

- ✅ **Automated trigger**: Runs when a new user is created (via `UserHasBeenCreated` hook)
- ✅ **Interactive UI**: Uses Slack-style attachments with buttons for each step
- ✅ **Persistent state**: Stores onboarding progress in Mattermost's KV store
- ✅ **Bot management**: Creates and manages a custom bot user with profile icon
- ✅ **Idempotent**: Won't restart onboarding if a user already has state
- ✅ **Configurable**: Admin can set a welcome channel via plugin settings

---

## Architecture

### Tech Stack

- **Language**: Go 1.22+
- **Framework**: Mattermost Plugin SDK (`github.com/mattermost/mattermost/server/public/plugin`)
- **State Management**: Mattermost KV Store (JSON-serialized state)
- **HTTP Integration**: Interactive message buttons with webhook callbacks

### Core Components

| Component | File | Purpose |
|-----------|------|---------|
| **Plugin Entry Point** | [`server/main.go`](server/main.go) | Initializes the plugin and registers it with Mattermost |
| **Plugin Core** | [`server/plugin.go`](server/plugin.go) | Main plugin struct, hooks (OnActivate, UserHasBeenCreated, ServeHTTP), bot management, KV helpers |
| **Onboarding Logic** | [`server/onboarding.go`](server/onboarding.go) | Starts onboarding, builds interactive checklist, handles button clicks |
| **Data Models** | [`server/model.go`](server/model.go) | Defines `OnboardingState` struct and step constants |
| **Manifest** | [`plugin.json`](plugin.json) | Plugin metadata, executable path, settings schema |
| **Build System** | [`Makefile`](Makefile) | Compiles Go binary and packages plugin for deployment |
| **Assets** | [`assets/icon.png`](assets/icon.png) | Bot profile picture |

---

## Project Structure

```
mm-onboarding-plugin/
├── plugin.json              # Plugin manifest (ID, name, version, settings)
├── Makefile                 # Build & package automation
├── go.mod                   # Go module dependencies
├── go.sum                   # Dependency checksums
├── assets/
│   └── icon.png             # Bot profile icon
├── server/
│   ├── main.go              # Plugin entry point
│   ├── plugin.go            # Core plugin logic & hooks
│   ├── onboarding.go        # Onboarding workflow & checklist
│   ├── model.go             # Data structures
│   └── dist/                # Compiled binary (generated)
└── dist/                    # Packaged plugin .tar.gz (generated)
```

---

## How It Works

### 1. Plugin Activation ([`plugin.go:29`](server/plugin.go#L29))

When the plugin is enabled:

- **`OnActivate()`** is called
- Ensures the bot user exists (creates or fetches it)
- Sets bot profile (display name, description, icon)
- Stores bot user ID in KV store for reuse

**Bot Configuration** ([`plugin.go:22-26`](server/plugin.go#L22-L26)):
```go
botUsername = "onboarding-assistant"
botDisplayName = "EOTO Onboarding Helper"
botDescription = "Guides new teammates through onboarding."
botIconPath = "assets/icon.png"
```

### 2. New User Detection ([`plugin.go:97`](server/plugin.go#L97))

When a user is created in Mattermost:

- **`UserHasBeenCreated()`** hook fires
- Ignores bot users
- Calls `startOnboardingForUser(user)`

### 3. Starting Onboarding ([`onboarding.go:12`](server/onboarding.go#L12))

**`startOnboardingForUser()`**:

1. **Checks if onboarding already started** (idempotent)
2. **Creates initial state**:
   ```go
   OnboardingState{
     UserID: user.Id,
     CompletedSteps: map[string]bool{},
     StartedAt: time.Now().UTC(),
   }
   ```
3. **Opens a DM channel** between bot and new user
4. **Sends welcome message** with interactive checklist

### 4. Interactive Checklist ([`onboarding.go:68`](server/onboarding.go#L68))

**`buildChecklistAttachments()`** creates 6 Slack-style attachments:

| Step | ID | Description |
|------|----|----|
| **Accounts & Access** | `accounts` | Google Workspace, Nextcloud, Timebutler, Mattermost, role-specific tools |
| **Complete Profile** | `profile` | Photo, name, pronouns, job title, timezone |
| **Communication Channels** | `channels` | Join #announcements, #helpdesk, #introductions, team channels |
| **Tools & Equipment** | `tools` | Laptop setup, Wi-Fi, Nextcloud client, email, VPN |
| **Policies & Practices** | `policies` | Working hours, GDPR, communication norms, file storage |
| **People & Check-ins** | `intro` | Introduction post, 1:1 with manager, onboarding buddy |

Each step includes:
- ✅/☐ checkbox (based on completion state)
- Descriptive text with sub-tasks
- Link to documentation (currently `https://outline.akinlosotu.tech`)
- **Button** that triggers HTTP callback to `/complete-step`

### 5. Button Click Handling ([`onboarding.go:220`](server/onboarding.go#L220))

When a user clicks a button:

1. **HTTP POST** sent to plugin endpoint `/complete-step`
2. **`handleCompleteStep()`** decodes the request
3. Validates step ID against allowed steps ([`model.go:16-22`](server/model.go#L16-L22))
4. **Loads user's state** from KV store
5. **Marks step complete**: `state.CompletedSteps[step] = true`
6. **Saves updated state**
7. **Rebuilds checklist** with updated checkboxes
8. **Responds with updated post** (checkboxes update in real-time)
9. Shows ephemeral confirmation: _"Marked step 'profile' complete ✔️"_

### 6. State Persistence ([`plugin.go:146-179`](server/plugin.go#L146-L179))

**KV Store Functions**:

- **`loadState(userID)`**: Fetches `OnboardingState` from KV (key: `onboarding:user:<userID>`)
- **`saveState(state)`**: Serializes state to JSON and stores it

**Data Structure** ([`model.go:5-10`](server/model.go#L5-L10)):
```go
type OnboardingState struct {
  UserID         string          `json:"user_id"`
  CompletedSteps map[string]bool `json:"completed_steps"`
  StartedAt      time.Time       `json:"started_at"`
  LastUpdated    time.Time       `json:"last_updated"`
}
```

---

## Customization Guide

### Adding/Removing Onboarding Steps

**1. Update Step Constants** ([`model.go:16-22`](server/model.go#L16-L22)):

```go
var onboardingSteps = []string{
  "profile",
  "channels",
  "handbook",
  "mfa",        // Add new step here
  "intro",
}
```

**2. Add Attachment in Checklist** ([`onboarding.go:68`](server/onboarding.go#L68)):

```go
{
  Title: "Step X: Enable MFA",
  Text: checkbox(isDone("mfa")) + " Secure your account with multi-factor authentication:\n" +
    "- Install authenticator app (Google Authenticator, Authy)\n" +
    "- Navigate to Account Settings → Security\n" +
    "- Scan QR code and verify\n\n" +
    "Guide: [MFA Setup](https://outline.akinlosotu.tech/mfa)",
  Actions: []*model.PostAction{
    {
      Name: "Mark MFA Enabled",
      Type: model.PostActionTypeButton,
      Integration: &model.PostActionIntegration{
        URL: callbackURL,
        Context: map[string]interface{}{
          "step": "mfa",
        },
      },
    },
  },
}
```

### Customizing Welcome Message

Edit [`onboarding.go:44-50`](server/onboarding.go#L44-L50):

```go
welcomeMsg := fmt.Sprintf(
  "👋 Hi %s, welcome to %s!\n\n"+
    "Your custom welcome message here...\n\n"+
    "_You can come back to this DM anytime to see your progress._",
  displayName,
  teamName,
)
```

### Changing Bot Appearance

**Bot Metadata** ([`plugin.go:22-26`](server/plugin.go#L22-L26)):
```go
const botUsername = "onboarding-assistant"
const botDisplayName = "EOTO Onboarding Helper"  // Change this
const botDescription = "Guides new teammates through onboarding."  // Change this
const botIconPath = "assets/icon.png"  // Replace icon.png
```

**Icon**: Replace [`assets/icon.png`](assets/icon.png) with your custom bot avatar (recommended size: 128x128 or 256x256).

### Customizing Documentation Links

All step attachments link to `https://outline.akinlosotu.tech`. Replace these in [`onboarding.go:68-209`](server/onboarding.go#L68-L209):

```go
"More details: [Accounts & Access Guide](https://your-docs-site.com/accounts)"
```

### Adjusting Plugin ID and Metadata

Edit [`plugin.json`](plugin.json):

```json
{
  "id": "com.akinlosotutech.onboarding",  // Change to your org domain
  "name": "Onboarding Assistant",
  "description": "Your custom description",
  "version": "0.1.0"
}
```

**⚠️ Important**: If you change the plugin ID, also update:
- [`plugin.go:136`](server/plugin.go#L136): `url.JoinPath(parsed.String(), "plugins", "com.permafx.onboarding")` (currently has a typo - should match plugin.json ID)
- Rebuild and reinstall the plugin (Mattermost caches plugins by ID)

### Adding Configuration Settings

Edit [`plugin.json:10-20`](plugin.json#L10-L20) to add admin-configurable options:

```json
"settings_schema": {
  "settings": [
    {
      "key": "WelcomeChannel",
      "display_name": "Welcome Channel",
      "type": "text",
      "help_text": "Optional: public channel to post a welcome message (e.g. town-square).",
      "default": "town-square"
    },
    {
      "key": "EnableSlackIntegration",
      "display_name": "Enable Slack Notifications",
      "type": "bool",
      "help_text": "Send onboarding status to Slack",
      "default": "false"
    }
  ]
}
```

Access settings in code:
```go
config := p.API.GetConfig()
// Settings are stored in PluginSettings.Plugins["com.akinlosotutech.onboarding"]
```

### Adding Completion Rewards/Actions

When all steps are complete, trigger an action in [`onboarding.go:220`](server/onboarding.go#L220):

```go
func (p *Plugin) handleCompleteStep(w http.ResponseWriter, r *http.Request) {
  // ... existing code ...

  state.CompletedSteps[step] = true

  // Check if all steps complete
  allDone := true
  for _, s := range onboardingSteps {
    if !state.CompletedSteps[s] {
      allDone = false
      break
    }
  }

  if allDone {
    // Send congratulations message
    p.API.SendEphemeralPost(userID, &model.Post{
      UserId:    p.botUserID,
      ChannelId: req.ChannelId,
      Message:   "🎉 Congratulations! You've completed all onboarding steps!",
    })

    // Post to public channel
    if channel, err := p.API.GetChannelByName("town-square", teamID, false); err == nil {
      p.API.CreatePost(&model.Post{
        UserId:    p.botUserID,
        ChannelId: channel.Id,
        Message:   fmt.Sprintf("Welcome @%s to the team! 🎉", username),
      })
    }
  }

  // ... rest of existing code ...
}
```

### Localizing Messages (i18n)

For multi-language support:

1. Add dependency:
   ```bash
   go get github.com/mattermost/mattermost-server/v6/model
   ```

2. Create localization files:
   ```
   server/i18n/
   ├── en.json
   └── de.json
   ```

3. Use localized strings in code:
   ```go
   welcomeMsg := p.API.T("onboarding.welcome", map[string]interface{}{
     "DisplayName": displayName,
     "TeamName": teamName,
   })
   ```

---

## Building & Deployment

### Prerequisites

- **Go 1.22+** (check with `go version`)
- **jq** (for extracting version from plugin.json)
  ```bash
  sudo apt install jq  # Ubuntu/Debian
  brew install jq      # macOS
  ```
- Mattermost server 9.0.0+ (specified in [`plugin.json:6`](plugin.json#L6))

### Build Process

The [`Makefile`](Makefile) automates compilation and packaging:

```bash
# Clean previous builds
make clean

# Build Go binary (creates server/dist/plugin-linux-amd64)
make build

# Build + package plugin (creates dist/com.akinlosotutech.onboarding-0.1.0.tar.gz)
make package

# Both clean + package
make clean package
```

**What `make package` does**:

1. Compiles Go code for Linux AMD64
2. Creates directory structure in `dist/`
3. Copies `plugin.json`, `server/dist/` binary, and `assets/`
4. Creates `.tar.gz` archive

**Output**:
```
dist/
└── com.akinlosotutech.onboarding-0.1.0.tar.gz
```

### Installing the Plugin

1. **Build the package**:
   ```bash
   make clean package
   ```

2. **Upload to Mattermost**:
   - Go to **System Console** → **Plugins** → **Management**
   - Click **Upload Plugin**
   - Select `dist/com.akinlosotutech.onboarding-0.1.0.tar.gz`
   - Click **Upload**

3. **Enable the plugin**:
   - Find "Onboarding Assistant" in the plugin list
   - Click **Enable**

4. **Configure (optional)**:
   - Click **Settings** to configure welcome channel

### Testing

**Create a test user** to trigger onboarding:

```bash
# Via mmctl (Mattermost CLI)
mmctl user create --email test@example.com --username testuser --password Password123!

# Or via System Console → Users → Create User
```

The bot should immediately DM the new user with the onboarding checklist.

---

## Configuration

### Plugin Settings

Configured in **System Console** → **Plugins** → **Onboarding Assistant**:

| Setting | Key | Description | Default |
|---------|-----|-------------|---------|
| **Welcome Channel** | `WelcomeChannel` | Public channel to post welcome message | `town-square` |

Currently, the welcome channel setting is **defined but not implemented**. To use it, add code in [`onboarding.go:12`](server/onboarding.go#L12):

```go
func (p *Plugin) startOnboardingForUser(user *model.User) error {
  // ... existing code ...

  // Post to public channel if configured
  channelName := p.API.GetPluginConfig().Settings["WelcomeChannel"]
  if channelName != "" {
    if channel, err := p.API.GetChannelByName(channelName, teamID, false); err == nil {
      p.API.CreatePost(&model.Post{
        UserId:    p.botUserID,
        ChannelId: channel.Id,
        Message:   fmt.Sprintf("Everyone welcome %s to the team! 👋", displayName),
      })
    }
  }

  // ... rest of code ...
}
```

### Environment Variables

The Makefile uses:

- **`GOCACHE`**: Go build cache (defaults to `.gocache/` in project)
- **`GOMODCACHE`**: Go module cache (defaults to `.gomodcache/` in project)

These keep caches local to avoid polluting system directories.

---

## Development Tips

### Local Development Setup

1. **Clone this repository**
2. **Install dependencies**:
   ```bash
   go mod download
   ```
3. **Run Go tests** (if you add them):
   ```bash
   go test ./...
   ```

### Debugging

**Add logging** throughout your code:

```go
p.API.LogInfo("User onboarding started", "user_id", user.Id)
p.API.LogError("Failed to create post", "err", err.Error())
p.API.LogDebug("Step completed", "step", step, "user_id", userID)
```

**View logs** in Mattermost:

- **System Console** → **Logs** (real-time)
- Or server logs file (configured in `config.json`)

### Common Issues

**Issue**: `pluginURL not configured` error

**Fix**: Ensure `SiteURL` is set in **System Console** → **Environment** → **Web Server** → **Site URL**

---

**Issue**: Buttons don't update when clicked

**Check**:
- Plugin ID in [`plugin.go:136`](server/plugin.go#L136) matches [`plugin.json:2`](plugin.json#L2)
- Site URL is accessible from Mattermost server (for webhook callbacks)
- Check server logs for HTTP errors

---

**Issue**: Bot user creation fails

**Cause**: Username collision (another user/bot has `onboarding-assistant`)

**Behavior**: Plugin automatically retries with timestamped username like `onboarding-assistant-1700000000`

---

### Testing KV Store Directly

Use Mattermost API or `mmctl`:

```bash
# View onboarding state for a user
mmctl plugin kv get com.akinlosotutech.onboarding "onboarding:user:USER_ID_HERE"

# Clear state (to re-trigger onboarding)
mmctl plugin kv delete com.akinlosotutech.onboarding "onboarding:user:USER_ID_HERE"
```

### Extending HTTP Endpoints

Add new routes in [`plugin.go:109-122`](server/plugin.go#L109-L122):

```go
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
  switch r.URL.Path {
  case "/complete-step":
    p.handleCompleteStep(w, r)
  case "/reset-onboarding":
    p.handleResetOnboarding(w, r)  // New endpoint
  case "/onboarding-stats":
    p.handleStats(w, r)  // New endpoint
  default:
    w.WriteHeader(http.StatusNotFound)
  }
}
```

Access at: `https://your-mattermost.com/plugins/com.akinlosotutech.onboarding/reset-onboarding`

### Adding Slash Commands

Register commands in `OnActivate()`:

```go
func (p *Plugin) OnActivate() error {
  // ... existing code ...

  if err := p.API.RegisterCommand(&model.Command{
    Trigger:          "onboarding-reset",
    AutoComplete:     true,
    AutoCompleteDesc: "Reset your onboarding checklist",
    AutoCompleteHint: "",
  }); err != nil {
    return err
  }

  return nil
}
```

Handle in `ExecuteCommand()`:

```go
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
  if args.Command == "/onboarding-reset" {
    // Delete user's state
    key := onboardingKVPrefix + args.UserId
    p.API.KVDelete(key)
    return &model.CommandResponse{
      ResponseType: model.CommandResponseTypeEphemeral,
      Text:         "Your onboarding has been reset!",
    }, nil
  }
  return &model.CommandResponse{}, nil
}
```

### Adding Database Migrations

For complex state, use Postgres instead of KV store:

1. Add database dependency:
   ```bash
   go get github.com/lib/pq
   ```

2. Access Mattermost DB in plugin (requires elevated permissions in plugin manifest)

3. Run migrations on `OnActivate()`

---

## Summary: What to Tweak for Your Org

| Customization | File(s) | Lines |
|---------------|---------|-------|
| **Plugin ID & name** | [`plugin.json`](plugin.json) | 2-4, [`plugin.go:136`](server/plugin.go#L136) |
| **Onboarding steps** | [`model.go`](server/model.go), [`onboarding.go`](server/onboarding.go) | 16-22, 68-209 |
| **Welcome message** | [`onboarding.go`](server/onboarding.go) | 44-50 |
| **Documentation links** | [`onboarding.go`](server/onboarding.go) | Throughout attachments |
| **Bot name & icon** | [`plugin.go`](server/plugin.go), [`assets/icon.png`](assets/icon.png) | 22-26 |
| **Settings schema** | [`plugin.json`](plugin.json) | 10-20 |
| **Completion actions** | [`onboarding.go`](server/onboarding.go) | 220-282 (in handler) |

---

## License

This plugin is provided as-is for EOTO organization. Modify and distribute as needed.

---

## Support

For issues, check:

1. **Mattermost logs**: System Console → Logs
2. **Plugin status**: System Console → Plugins → Management
3. **KV store**: Use `mmctl plugin kv` commands
4. **Server health**: Ensure Site URL is configured correctly

---

**Happy onboarding! 🚀**
