# EOTO Mattermost Onboarding Plugin

A comprehensive, multilingual Mattermost plugin that provides automated, interactive onboarding for new EOTO team members. Features include an interactive checklist, integrated email signature generator with authentic EOTO project templates, and full German/English language support.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [How It Works](#how-it-works)
- [Internationalization (i18n)](#internationalization-i18n)
- [Email Signature Generator](#email-signature-generator)
- [Customization Guide](#customization-guide)
- [Building & Deployment](#building--deployment)
- [Configuration](#configuration)
- [Development Tips](#development-tips)
- [Troubleshooting](#troubleshooting)

---

## Overview

This plugin automates the onboarding process for new users joining your Mattermost workspace. It:

1. **Detects new users** automatically when they're created
2. **Creates a dedicated onboarding bot** that DMs the new user
3. **Sends an interactive checklist** with steps covering accounts, profile setup, channels, tools, policies, and introductions
4. **Tracks progress** persistently using Mattermost's key-value store
5. **Updates in real-time** as users click buttons to mark steps complete
6. **Generates EOTO email signatures** with project-specific templates (Each One, CommUnity, CUZ, Jugend, NAR, Afrolution)
7. **Supports multiple languages** (German default, English available) configured via admin settings

---

## Features

### Core Functionality
- ✅ **Automated trigger**: Runs when a new user is created (via `UserHasBeenCreated` hook)
- ✅ **Interactive UI**: Uses Slack-style attachments with buttons for each step
- ✅ **Persistent state**: Stores onboarding progress in Mattermost's KV store
- ✅ **Bot management**: Creates and manages a custom bot user with profile icon
- ✅ **Idempotent**: Won't restart onboarding if a user already has state
- ✅ **Welcome message preservation**: Messages remain intact when marking steps complete

### Advanced Features
- 🌍 **Multilingual (i18n)**: Full German and English translations
- ✉️ **Email signature generator**: Interactive dialog with 6 authentic EOTO project templates
- 🎨 **Professional formatting**: Table-based HTML signatures compatible with all email clients
- 👤 **Pronoun support**: Format pronouns in both German and English (e.g., "Pronomen er/ihm - pronouns he/him")
- 📊 **Project templates**: Each One, CommUnity, CUZ, Jugend, NAR, Afrolution
- 🔧 **Admin configurable**: Language setting in System Console

---

## Architecture

### Tech Stack

- **Language**: Go 1.25
- **Framework**: Mattermost Plugin SDK v0.1.21 (`github.com/mattermost/mattermost/server/public`)
- **State Management**: Mattermost KV Store (JSON-serialized state)
- **HTTP Integration**: Interactive message buttons with webhook callbacks
- **Templating**: Go `html/template` for email signatures
- **Build System**: Makefile with explicit Go 1.25 binary path (`/usr/local/go/bin/go`)

### Core Components

| Component | File | Purpose |
|-----------|------|---------|
| **Plugin Entry Point** | [`server/main.go`](server/main.go) | Initializes the plugin and registers it with Mattermost |
| **Plugin Core** | [`server/plugin.go`](server/plugin.go) | Main plugin struct, hooks (OnActivate, UserHasBeenCreated, ServeHTTP), bot management, HTTP routing |
| **Onboarding Logic** | [`server/onboarding.go`](server/onboarding.go) | Starts onboarding, builds interactive checklist, handles button clicks |
| **Signature Generator** | [`server/signature.go`](server/signature.go) | Interactive dialog for signature generation, form validation, file upload |
| **Signature Templates** | [`server/signature_templates.go`](server/signature_templates.go) | 6 authentic EOTO project-specific HTML email templates (500+ lines) |
| **Internationalization** | [`server/i18n.go`](server/i18n.go) | Translation infrastructure, language detection, helper functions |
| **German Translations** | [`server/i18n_de.go`](server/i18n_de.go) | Complete German translations (default language) |
| **English Translations** | [`server/i18n_en.go`](server/i18n_en.go) | Complete English translations |
| **Data Models** | [`server/model.go`](server/model.go) | Defines `OnboardingState` struct and step constants |
| **Manifest** | [`plugin.json`](plugin.json) | Plugin metadata, executable path, language settings schema |
| **Build System** | [`Makefile`](Makefile) | Compiles Go binary and packages plugin for deployment |
| **Assets** | [`assets/icon.png`](assets/icon.png) | Bot profile picture |

---

## Project Structure

```
mm-onboarding-plugin/
├── plugin.json                      # Plugin manifest (ID, name, version, language settings)
├── Makefile                         # Build & package automation (uses Go 1.25)
├── go.mod                           # Go module dependencies (v1.25)
├── go.sum                           # Dependency checksums
├── README.md                        # This file
├── assets/
│   └── icon.png                     # Bot profile icon (128x128 or 256x256 recommended)
├── server/
│   ├── main.go                      # Plugin entry point
│   ├── plugin.go                    # Core plugin logic, hooks, HTTP routing
│   ├── onboarding.go                # Onboarding workflow, checklist builder, step completion
│   ├── signature.go                 # Signature dialog handler, form submission, file upload
│   ├── signature_templates.go       # 6 EOTO project email signature templates
│   ├── i18n.go                      # Translation infrastructure and helpers
│   ├── i18n_de.go                   # German translations (default)
│   ├── i18n_en.go                   # English translations
│   ├── model.go                     # Data structures (OnboardingState, SignatureData)
│   └── dist/                        # Compiled binary (generated during build)
└── dist/                            # Packaged plugin .tar.gz (generated by make package)
```

---

## How It Works

### 1. Plugin Activation ([`plugin.go:29`](server/plugin.go))

When the plugin is enabled in System Console:

- **`OnActivate()`** is called
- Ensures the bot user exists (creates or fetches it)
- Sets bot profile (display name, description, icon from `assets/icon.png`)
- Stores bot user ID in KV store for reuse
- Registers HTTP routes for button callbacks

**Bot Configuration** ([`plugin.go:22-26`](server/plugin.go)):
```go
botUsername = "onboarding-assistant"
botDisplayName = "EOTO Onboarding Helper"
botDescription = "Guides new teammates through onboarding."
botIconPath = "assets/icon.png"
```

### 2. New User Detection ([`plugin.go:97`](server/plugin.go))

When a user is created in Mattermost:

- **`UserHasBeenCreated()`** hook fires
- Ignores bot users (checks `user.IsBot`)
- Calls `startOnboardingForUser(user)` to begin the workflow

### 3. Starting Onboarding ([`onboarding.go:12`](server/onboarding.go))

**`startOnboardingForUser()`**:

1. **Checks if onboarding already started** (idempotent via `loadState()`)
2. **Creates initial state**:
   ```go
   OnboardingState{
     UserID: user.Id,
     CompletedSteps: map[string]bool{},
     StartedAt: time.Now().UTC(),
   }
   ```
3. **Opens a DM channel** between bot and new user
4. **Sends welcome message** with user's full name or username
5. **Attaches interactive checklist** with buttons

### 4. Interactive Checklist ([`onboarding.go:68`](server/onboarding.go))

**`buildChecklistAttachments()`** creates 6 Slack-style attachments, each representing an onboarding step:

| Step | ID | Key Components |
|------|----|----|
| **Step 1: Accounts & Access** | `accounts` | Google Workspace, Nextcloud, Timebutler, Mattermost, role-specific tools |
| **Step 2: Complete Your Profile** | `profile` | Photo, name, pronouns, job title, timezone, **email signature generator button** |
| **Step 3: Communication Channels** | `channels` | #announcements, #helpdesk, #introductions, team channels |
| **Step 4: Tools & Equipment** | `tools` | Laptop setup, Wi-Fi, Nextcloud client, email, VPN |
| **Step 5: Work Practices & Policies** | `policies` | Working hours, flextime, GDPR, communication norms, file storage |
| **Step 6: People & Check-ins** | `intro` | Introduction post, 1:1 with manager, onboarding buddy |

Each step includes:
- ✅/☐ checkbox (dynamically updates based on completion state)
- Descriptive text with sub-tasks in the selected language
- Link to documentation (`https://outline.akinlosotu.tech`)
- **Button(s)** that trigger HTTP callback to `/complete-step`

**Step 2 is special**: It has TWO buttons:
1. "Generate Email Signature" → Opens interactive dialog
2. "Mark Profile Complete" → Marks step as done

### 5. Button Click Handling ([`onboarding.go:231`](server/onboarding.go))

When a user clicks a button:

1. **HTTP POST** sent to plugin endpoint `/complete-step`
2. **`handleCompleteStep()`** decodes the `PostActionIntegrationRequest`
3. **Routes based on action**:
   - If `context["action"] == "open_signature_dialog"` → Call `handleSignatureDialog()`
   - Otherwise, continue with step completion logic
4. Validates step ID against allowed steps ([`model.go:16-23`](server/model.go))
5. **Loads user's state** from KV store
6. **Marks step complete**: `state.CompletedSteps[step] = true`
7. **Saves updated state** to KV store
8. **Rebuilds checklist** with updated checkboxes
9. **Reconstructs welcome message** to prevent it from disappearing
10. **Responds with updated post** (checkboxes update in real-time)
11. Shows ephemeral confirmation using translated message

### 6. Welcome Message Preservation

**Problem solved**: When Mattermost updates a post via `PostActionIntegrationResponse`, if you only update `Props` (attachments), the original message text disappears and shows "Edited" tag.

**Solution** ([`onboarding.go:273-290`](server/onboarding.go)):
```go
// Get user info to rebuild welcome message
user, appErr := p.API.GetUser(userID)
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

// Include in response
resp := &model.PostActionIntegrationResponse{
    Update: &model.Post{
        Message:   welcomeMsg,  // KEY: This preserves the message
        Props: map[string]interface{}{
            "attachments": attachments,
        },
    },
    EphemeralText: fmt.Sprintf(tr.StepMarkedComplete, step),
}
```

### 7. State Persistence ([`plugin.go:146-179`](server/plugin.go))

**KV Store Functions**:

- **`loadState(userID)`**: Fetches `OnboardingState` from KV (key: `onboarding:user:<userID>`)
- **`saveState(state)`**: Serializes state to JSON and stores it in KV
- **`ensureBotUser()`**: Ensures bot exists, creates if needed, caches bot ID in KV

**Data Structure** ([`model.go:5-11`](server/model.go)):
```go
type OnboardingState struct {
    UserID         string          `json:"user_id"`
    CompletedSteps map[string]bool `json:"completed_steps"`
    StartedAt      time.Time       `json:"started_at"`
    LastUpdated    time.Time       `json:"last_updated,omitempty"`
}
```

**Allowed Steps** ([`model.go:16-23`](server/model.go)):
```go
var onboardingSteps = []string{
    "accounts",
    "profile",
    "channels",
    "tools",
    "policies",
    "intro",
}
```

---

## Internationalization (i18n)

The plugin supports **full multilingual operation** with German (default) and English.

### How It Works

1. **Admin Configuration** ([`plugin.json`](plugin.json)):
   - System Console → Plugins → Onboarding Assistant
   - Setting: "Bot Language / Bot-Sprache"
   - Options: Deutsch (de), English (en)
   - Default: German (`de`)

2. **Translation Infrastructure** ([`server/i18n.go`](server/i18n.go)):
   ```go
   type Translations struct {
       WelcomeGreeting string
       WelcomeIntro    string
       WelcomeClosing  string
       Step1Title      string
       Step1Description string
       // ... 60+ translation keys
   }

   func (p *Plugin) getTranslations() Translations {
       language := p.getPluginSetting("Language", "de")
       switch language {
       case "en":
           return translationsEN
       case "de":
           return translationsDE
       default:
           return translationsDE
       }
   }
   ```

3. **Usage in Code**:
   ```go
   tr := p.getTranslations()
   welcomeMsg := fmt.Sprintf(tr.WelcomeGreeting, displayName, teamName)
   ```

### Translation Files

- **[`server/i18n_de.go`](server/i18n_de.go)**: German translations (Willkommen, Schritt 1, etc.)
- **[`server/i18n_en.go`](server/i18n_en.go)**: English translations (Welcome, Step 1, etc.)

### What's Translated

- ✅ Welcome messages
- ✅ All 6 step titles and descriptions
- ✅ Button labels
- ✅ Signature dialog fields and help text
- ✅ Project names in dropdown
- ✅ Success and error messages
- ✅ Step completion confirmations

### Adding a New Language

1. **Create new translation file** (e.g., `server/i18n_fr.go`):
   ```go
   package main

   var translationsFR = Translations{
       WelcomeGreeting: "👋 Bonjour %s, bienvenue à %s!",
       WelcomeIntro:    "Je suis votre assistant d'intégration...",
       // ... translate all 60+ keys
   }
   ```

2. **Update language switch** in [`server/i18n.go`](server/i18n.go):
   ```go
   func (p *Plugin) getTranslations() Translations {
       language := p.getPluginSetting("Language", "de")
       switch language {
       case "en":
           return translationsEN
       case "de":
           return translationsDE
       case "fr":
           return translationsFR
       default:
           return translationsDE
       }
   }
   ```

3. **Add to plugin settings** in [`plugin.json`](plugin.json):
   ```json
   {
     "key": "Language",
     "type": "dropdown",
     "options": [
       {"display_name": "Deutsch", "value": "de"},
       {"display_name": "English", "value": "en"},
       {"display_name": "Français", "value": "fr"}
     ]
   }
   ```

---

## Email Signature Generator

The plugin includes an **integrated email signature generator** with authentic EOTO project templates.

### Features

- 🎨 **6 project-specific templates**: Each One, CommUnity, CUZ, Jugend, NAR, Afrolution
- 👤 **Pronoun formatting**: Supports bilingual pronouns (e.g., "er/ihm / he/him" → "Pronomen er/ihm - pronouns he/him")
- 📞 **Optional work number**: Omits field entirely if empty (matches Python app behavior)
- 📧 **Email-compatible HTML**: Table-based layouts work in Outlook, Thunderbird, Gmail
- 💾 **File download**: Generates `.html` file user can download and install
- 🌍 **Multilingual dialog**: Form labels translated based on language setting

### How It Works

1. **User clicks "Generate Email Signature"** button in Step 2 of checklist

2. **`handleSignatureDialog()`** opens interactive dialog ([`server/signature.go:13`](server/signature.go)):
   - Pre-fills full name and email from user profile
   - Shows dropdown with 6 EOTO projects
   - Fields: Full Name, Position, Pronouns (optional), Email, Project, Work Number (optional)

3. **User submits form**

4. **`handleSignatureSubmission()`** processes submission ([`server/signature.go:112`](server/signature.go)):
   - Validates required fields
   - Calls `GenerateSignature()` with form data

5. **`GenerateSignature()`** creates HTML ([`server/signature_templates.go:26`](server/signature_templates.go)):
   ```go
   func GenerateSignature(data SignatureData) (string, error) {
       // Format pronouns: "er/ihm / he/him" → "Pronomen er/ihm - pronouns he/him"
       formattedPronouns := formatPronouns(data.Pronouns)

       // Select template based on project
       var templateStr string
       switch data.Project {
       case "each-one":
           templateStr = eachOneTemplate
       case "community":
           templateStr = communityTemplate
       // ... etc for all 6 projects
       }

       // Parse and execute template
       tmpl, err := template.New("signature").Parse(templateStr)
       tmpl.Execute(&buf, data)

       // Remove work number row if empty
       if data.WorkNumber == "" {
           html = removeWorkNumberFromHTML(html)
       }

       return html, nil
   }
   ```

6. **Upload signature file** to Mattermost ([`server/signature.go:234`](server/signature.go)):
   - Filename format: `{Name}_{project}_Signatur.html` (e.g., `Max_Mustermann_each-one_Signatur.html`)
   - Posted to DM with bot

7. **Bot sends instructions** for installing signature in email clients

### Signature Templates

Each template ([`server/signature_templates.go`](server/signature_templates.go)) includes:
- Project-specific logo
- EOTO branding and colors
- Contact information (email, phone, pronouns)
- Social media links
- Table-based HTML for email client compatibility

**Projects**:
1. **Each One** (`each-one`)
2. **CommUnity** (`community`)
3. **CommUnity Zentrum** (`cuz`)
4. **Jugendangebote** (`jugend`)
5. **Netzwerk-Antirassismus** (`nar`)
6. **Afrolution** (`afrolution`)

### Pronoun Formatting

The plugin formats pronouns to match the Python app's behavior ([`signature_templates.go:96`](server/signature_templates.go)):

```go
func formatPronouns(pronouns string) string {
    if pronouns == "" {
        return ""
    }
    // Input: "er/ihm / he/him"
    // Output: "Pronomen er/ihm - pronouns he/him"
    if strings.Contains(pronouns, " / ") {
        parts := strings.Split(pronouns, " / ")
        if len(parts) == 2 {
            germanPart := strings.TrimSpace(parts[0])
            englishPart := strings.TrimSpace(parts[1])
            return fmt.Sprintf("Pronomen %s - pronouns %s", germanPart, englishPart)
        }
    }
    return pronouns
}
```

### Customizing Signature Templates

To update a signature template:

1. **Edit template string** in [`server/signature_templates.go`](server/signature_templates.go)
2. Look for the project's template variable (e.g., `eachOneTemplate`, `communityTemplate`)
3. Modify HTML (remember: use tables, not divs, for email compatibility)
4. Test in multiple email clients (Outlook, Thunderbird, Gmail)
5. Rebuild: `make clean package`

**Example**: Change Each One logo URL:
```go
const eachOneTemplate = `
<!DOCTYPE html>
<html>
<body>
<table>
    <tr>
        <td><img src="https://new-url.com/logo.png" width="150" /></td>
        ...
```

---

## Customization Guide

### Adding/Removing Onboarding Steps

**1. Update Step Constants** ([`model.go:16-23`](server/model.go)):

```go
var onboardingSteps = []string{
    "accounts",
    "profile",
    "channels",
    "tools",
    "policies",
    "mfa",        // NEW: Add multi-factor authentication step
    "intro",
}
```

**2. Add Translations** for new step in [`i18n_de.go`](server/i18n_de.go) and [`i18n_en.go`](server/i18n_en.go):

```go
// In i18n_de.go
Step7Title: "Schritt 7: Zwei-Faktor-Authentifizierung",
Step7Description: "Sichere dein Konto mit MFA:\n" +
    "- Installiere Authenticator-App (Google Authenticator, Authy)\n" +
    "- Gehe zu Kontoeinstellungen → Sicherheit\n" +
    "- Scanne QR-Code und verifiziere\n\n" +
    "Anleitung: ",
Step7Link: "[MFA-Einrichtung](https://outline.akinlosotu.tech/s/mfa)",
ButtonMarkMFAEnabled: "MFA aktiviert markieren",

// In i18n_en.go
Step7Title: "Step 7: Multi-Factor Authentication",
Step7Description: "Secure your account with MFA:\n" +
    "- Install authenticator app (Google Authenticator, Authy)\n" +
    "- Navigate to Account Settings → Security\n" +
    "- Scan QR code and verify\n\n" +
    "Guide: ",
Step7Link: "[MFA Setup](https://outline.akinlosotu.tech/s/mfa)",
ButtonMarkMFAEnabled: "Mark MFA Enabled",
```

**3. Update Translations struct** in [`i18n.go`](server/i18n.go):

```go
type Translations struct {
    // ... existing fields ...
    Step7Title       string
    Step7Description string
    Step7Link        string
    ButtonMarkMFAEnabled string
}
```

**4. Add Attachment in Checklist** ([`onboarding.go:68`](server/onboarding.go)):

```go
{
    Title: tr.Step7Title,
    Text:  checkbox(isDone("mfa")) + " " + tr.Step7Description + tr.Step7Link,
    Actions: []*model.PostAction{
        {
            Name: tr.ButtonMarkMFAEnabled,
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
```

### Customizing Welcome Message

Edit translation files ([`i18n_de.go`](server/i18n_de.go), [`i18n_en.go`](server/i18n_en.go)):

```go
// German
WelcomeGreeting: "👋 Willkommen %s bei %s!",
WelcomeIntro:    "Deine eigene Begrüßungsnachricht hier...",
WelcomeClosing:  "_Du kannst jederzeit zu dieser Nachricht zurückkehren._",

// English
WelcomeGreeting: "👋 Welcome %s to %s!",
WelcomeIntro:    "Your custom welcome message here...",
WelcomeClosing:  "_You can return to this message anytime._",
```

### Changing Bot Appearance

**Bot Metadata** ([`plugin.go:22-26`](server/plugin.go)):
```go
const botUsername = "onboarding-assistant"        // Internal username (lowercase, no spaces)
const botDisplayName = "Your Custom Bot Name"    // Displayed name in UI
const botDescription = "Your custom description" // Bot profile description
const botIconPath = "assets/icon.png"            // Path to bot avatar
```

**Icon**: Replace [`assets/icon.png`](assets/icon.png):
- Recommended size: 128x128 or 256x256 pixels
- Format: PNG with transparency
- Square aspect ratio

### Customizing Documentation Links

Update translation files to change all documentation URLs:

```go
// In i18n_de.go and i18n_en.go
Step1Link: "[Konten & Zugriff](https://your-docs.example.com/accounts)",
Step2Link: "[Profil einrichten](https://your-docs.example.com/profile)",
// ... etc
```

### Adjusting Plugin ID and Metadata

**⚠️ CRITICAL**: Plugin ID must be consistent across files!

1. **Edit [`plugin.json`](plugin.json)**:
   ```json
   {
     "id": "com.yourorg.onboarding",
     "name": "Your Org Onboarding",
     "description": "Your custom description",
     "version": "0.2.0"
   }
   ```

2. **Update plugin ID reference** in [`plugin.go:136`](server/plugin.go):
   ```go
   resolved, err := url.JoinPath(parsed.String(), "plugins", "com.yourorg.onboarding")
   ```

3. **Update Makefile** ([`Makefile:3`](Makefile)):
   ```makefile
   PLUGIN_ID := com.yourorg.onboarding
   ```

4. **Rebuild completely**:
   ```bash
   make clean
   make package
   ```

### Adding Configuration Settings

Edit [`plugin.json`](plugin.json) settings schema:

```json
"settings_schema": {
    "settings": [
        {
            "key": "Language",
            "display_name": "Bot Language / Bot-Sprache",
            "type": "dropdown",
            "default": "de",
            "options": [
                {"display_name": "Deutsch", "value": "de"},
                {"display_name": "English", "value": "en"}
            ]
        },
        {
            "key": "WelcomeChannel",
            "display_name": "Welcome Channel",
            "type": "text",
            "help_text": "Channel to post public welcome (e.g., town-square)",
            "default": "town-square"
        },
        {
            "key": "EnableSignatureGenerator",
            "display_name": "Enable Signature Generator",
            "type": "bool",
            "help_text": "Show email signature button in onboarding",
            "default": "true"
        }
    ]
}
```

Access settings in code:
```go
func (p *Plugin) getPluginSetting(key, defaultValue string) string {
    config := p.API.GetConfig()
    if config == nil || config.PluginSettings.Plugins == nil {
        return defaultValue
    }
    pluginConfig, ok := config.PluginSettings.Plugins["com.akinlosotutech.onboardinghelper"]
    if !ok {
        return defaultValue
    }
    value, ok := pluginConfig[key].(string)
    if !ok {
        return defaultValue
    }
    return value
}
```

### Adding Completion Rewards/Actions

Track when all steps are complete and trigger actions:

```go
func (p *Plugin) handleCompleteStep(w http.ResponseWriter, r *http.Request) {
    // ... existing code ...

    state.CompletedSteps[step] = true

    // Check if all steps complete
    allComplete := len(state.CompletedSteps) == len(onboardingSteps)
    for _, s := range onboardingSteps {
        if !state.CompletedSteps[s] {
            allComplete = false
            break
        }
    }

    if allComplete {
        tr := p.getTranslations()

        // Send congratulations DM
        p.API.CreatePost(&model.Post{
            UserId:    p.botUserID,
            ChannelId: req.ChannelId,
            Message:   "🎉 " + tr.CompletionMessage,
        })

        // Post to public channel
        if channel, err := p.API.GetChannelByName("town-square", teamID, false); err == nil {
            user, _ := p.API.GetUser(userID)
            p.API.CreatePost(&model.Post{
                UserId:    p.botUserID,
                ChannelId: channel.Id,
                Message:   fmt.Sprintf("Willkommen @%s im Team! 🎉", user.Username),
            })
        }
    }

    // ... rest of existing code ...
}
```

### Customizing Signature Templates for Your Organization

If you're not EOTO, you'll want to replace the signature templates:

1. **Identify your projects/departments** (replace Each One, CommUnity, etc.)

2. **Update project list** in [`signature.go:72-79`](server/signature.go):
   ```go
   Options: []*model.PostActionOptions{
       {Text: "Marketing", Value: "marketing"},
       {Text: "Engineering", Value: "engineering"},
       {Text: "Sales", Value: "sales"},
   },
   ```

3. **Create new templates** in [`signature_templates.go`](server/signature_templates.go):
   ```go
   const marketingTemplate = `
   <!DOCTYPE html>
   <html>
   <body>
   <table style="font-family: Arial, sans-serif; font-size: 12pt;">
       <tr>
           <td>
               <strong>{{.FullName}}</strong><br>
               {{.Position}}
               {{if .Pronouns}}<br><span style="font-size: 10pt;">{{.Pronouns}}</span>{{end}}
           </td>
       </tr>
       <!-- Your org's branding here -->
   </table>
   </body>
   </html>
   `
   ```

4. **Update `GenerateSignature()` switch** ([`signature_templates.go:38`](server/signature_templates.go)):
   ```go
   switch data.Project {
   case "marketing":
       templateStr = marketingTemplate
   case "engineering":
       templateStr = engineeringTemplate
   // ...
   }
   ```

5. **Update translations** for project names in [`i18n_de.go`](server/i18n_de.go) and [`i18n_en.go`](server/i18n_en.go)

---

## Building & Deployment

### Prerequisites

- **Go 1.25** installed at `/usr/local/go/bin/go` (check with `/usr/local/go/bin/go version`)
- **jq** (for extracting version from plugin.json)
  ```bash
  sudo apt install jq  # Ubuntu/Debian
  brew install jq      # macOS
  ```
- **Mattermost Server 9.0.0+** (specified in [`plugin.json:6`](plugin.json))

### Build Process

The [`Makefile`](Makefile) automates compilation and packaging:

```bash
# Clean previous builds
make clean

# Build Go binary only (creates server/dist/plugin-linux-amd64)
make build

# Build + package plugin (creates dist/com.akinlosotutech.onboardinghelper-0.1.0.tar.gz)
make package

# Both clean + package in one command
make clean package
```

**What `make package` does**:

1. Compiles Go code for Linux AMD64 using `/usr/local/go/bin/go`
2. Creates directory structure in `dist/com.akinlosotutech.onboardinghelper/`
3. Copies:
   - `plugin.json` (manifest)
   - `server/dist/plugin-linux-amd64` (binary)
   - `assets/icon.png` (bot avatar)
4. Creates `.tar.gz` archive

**Output**:
```
dist/
├── com.akinlosotutech.onboardinghelper/
│   ├── plugin.json
│   ├── assets/
│   │   └── icon.png
│   └── server/
│       └── dist/
│           └── plugin-linux-amd64
└── com.akinlosotutech.onboardinghelper-0.1.0.tar.gz  ← Upload this
```

### Go Version Management

This project requires Go 1.25. The Makefile explicitly uses `/usr/local/go/bin/go` to avoid version conflicts.

**If you see "package maps is not in GOROOT" error**:
```bash
# Verify Go 1.25 is installed
/usr/local/go/bin/go version

# Should show: go version go1.25.x linux/amd64

# If not, install Go 1.25+:
wget https://go.dev/dl/go1.25.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.4.linux-amd64.tar.gz
```

### Installing the Plugin

1. **Build the package**:
   ```bash
   make clean package
   ```

2. **Upload to Mattermost**:
   - Log in as System Admin
   - Go to **System Console** → **Plugins** → **Plugin Management**
   - Click **Upload Plugin**
   - Select `dist/com.akinlosotutech.onboardinghelper-0.1.0.tar.gz`
   - Click **Upload**

3. **Enable the plugin**:
   - Find "Onboarding Assistant" in the plugin list
   - Click **Enable**

4. **Configure language** (optional):
   - Click **Settings**
   - Choose "Bot Language / Bot-Sprache"
   - Select Deutsch or English
   - Click **Save**

### Updating the Plugin

**When you make code changes**:

1. **Increment version** in [`plugin.json`](plugin.json):
   ```json
   "version": "0.2.0"
   ```

2. **Rebuild**:
   ```bash
   make clean package
   ```

3. **Remove old plugin** in System Console → Plugins → Management

4. **Upload new version**

5. **Enable** the new version

**Alternative**: Some Mattermost versions support "force replace" which overwrites the existing plugin without removing it first.

### Testing

**Create a test user** to trigger onboarding:

```bash
# Via mmctl (Mattermost CLI)
mmctl user create --email test@example.com --username testuser --password Password123!

# Or via System Console → User Management → Create User
```

The bot should immediately DM the new user with the onboarding checklist in the configured language.

**Test signature generator**:
1. Log in as test user
2. Find DM from "EOTO Onboarding Helper" bot
3. Click "✉️ E-Mail-Signatur generieren" (German) or "✉️ Generate Email Signature" (English)
4. Fill form and submit
5. Verify signature HTML file is posted

---

## Configuration

### Plugin Settings

Configured in **System Console** → **Plugins** → **Onboarding Assistant**:

| Setting | Key | Type | Description | Default |
|---------|-----|------|-------------|---------|
| **Bot Language** | `Language` | Dropdown | Language for all bot messages and UI | `de` (German) |

### Environment Variables (Build-time)

The Makefile uses:

- **`GO`**: Path to Go 1.25 binary (hardcoded to `/usr/local/go/bin/go`)
- **`GOCACHE`**: Go build cache (set to `.gocache/` in project)
- **`GOMODCACHE`**: Go module cache (set to `.gomodcache/` in project)

These keep caches local to avoid polluting system directories.

---

## Development Tips

### Local Development Setup

1. **Clone this repository**
2. **Install dependencies**:
   ```bash
   /usr/local/go/bin/go mod download
   /usr/local/go/bin/go mod tidy
   ```
3. **Run Go tests** (when you add them):
   ```bash
   /usr/local/go/bin/go test ./...
   ```

### Debugging

**Add logging** throughout your code:

```go
p.API.LogInfo("User onboarding started", "user_id", user.Id, "language", language)
p.API.LogError("Failed to create post", "err", err.Error())
p.API.LogDebug("Step completed", "step", step, "user_id", userID)
```

**View logs** in Mattermost:

- **Real-time**: System Console → Logs
- **File**: Server logs file (path configured in `config.json`)

**Search logs**:
```bash
# On server
grep "onboarding" /var/log/mattermost/mattermost.log
```

### Testing KV Store Directly

Use `mmctl` to inspect/modify onboarding state:

```bash
# View onboarding state for a user
mmctl plugin kv get com.akinlosotutech.onboardinghelper "onboarding:user:<USER_ID>"

# Clear state (to re-trigger onboarding for testing)
mmctl plugin kv delete com.akinlosotutech.onboardinghelper "onboarding:user:<USER_ID>"

# List all onboarding states
mmctl plugin kv list com.akinlosotutech.onboardinghelper --filter "onboarding:user:"
```

### Adding HTTP Endpoints

Add new routes in [`plugin.go:116-126`](server/plugin.go):

```go
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/complete-step":
        p.handleCompleteStep(w, r)
    case "/submit-signature":
        p.handleSignatureSubmission(w, r)
    case "/reset-onboarding":
        p.handleResetOnboarding(w, r)  // NEW
    case "/onboarding-stats":
        p.handleStats(w, r)  // NEW
    default:
        w.WriteHeader(http.StatusNotFound)
    }
}
```

Access at: `https://your-mattermost.com/plugins/com.akinlosotutech.onboardinghelper/reset-onboarding`

### Adding Slash Commands

Register commands in `OnActivate()`:

```go
func (p *Plugin) OnActivate() error {
    // ... existing code ...

    if err := p.API.RegisterCommand(&model.Command{
        Trigger:          "onboarding-reset",
        AutoComplete:     true,
        AutoCompleteDesc: "Reset your onboarding checklist",
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
        key := fmt.Sprintf("onboarding:user:%s", args.UserId)
        p.API.KVDelete(key)

        tr := p.getTranslations()
        return &model.CommandResponse{
            ResponseType: model.CommandResponseTypeEphemeral,
            Text:         tr.OnboardingResetMessage,
        }, nil
    }
    return &model.CommandResponse{}, nil
}
```

### Hot Reload During Development

Mattermost doesn't support hot reload for Go plugins, so you must:

1. Make code changes
2. `make clean package`
3. Upload new plugin via System Console
4. Disable old version
5. Enable new version

**Tip**: Bump the version number in `plugin.json` each time to avoid confusion.

---

## Troubleshooting

### Build Issues

**Q: `go.mod:5: unknown directive: toolchain`**

**A**: Remove the `toolchain` line from `go.mod`:
```bash
sed -i '/toolchain/d' go.mod
/usr/local/go/bin/go mod tidy
make clean package
```

---

**Q: `package maps is not in GOROOT`**

**A**: You're using Go 1.18 instead of Go 1.25+:
```bash
# Verify version
/usr/local/go/bin/go version

# Should show go1.25.x

# If not, install Go 1.25:
wget https://go.dev/dl/go1.25.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.4.linux-amd64.tar.gz
```

---

**Q: Build succeeds but plugin won't enable**

**A**: Check Mattermost logs for errors:
```bash
# On server
tail -f /var/log/mattermost/mattermost.log | grep onboarding
```

Common causes:
- Plugin ID mismatch between `plugin.json` and `plugin.go:136`
- Missing dependencies (run `go mod tidy`)
- Incompatible Mattermost version (need 9.0.0+)

---

### Deployment Issues

**Q: Plugin uploads but doesn't appear in list**

**A**:
- Check System Console → Plugins → Management → "Upload Plugin" → Look for error messages
- Verify `.tar.gz` structure: `tar -tzf yourplugin.tar.gz | head -20`
- Should show: `com.akinlosotutech.onboardinghelper/plugin.json` at root

---

**Q: Plugin enables but bot doesn't send messages**

**A**:
1. Check `SiteURL` is set: System Console → Environment → Web Server → Site URL
2. Verify bot was created: System Console → User Management → Search for "onboarding-assistant"
3. Check logs for errors during `OnActivate()`

---

**Q: Buttons don't work when clicked**

**A**:
- **Most common**: Plugin ID mismatch. Verify:
  ```bash
  grep '"id"' plugin.json
  grep 'plugins/' server/plugin.go
  ```
  Must match exactly!
- Check Mattermost can reach its own `SiteURL` (webhook callbacks need this)
- Look for HTTP errors in logs when button is clicked

---

**Q: Signature dialog doesn't open**

**A**:
- Check logs for "failed to open dialog"
- Verify user has permission to use dialogs
- Try updating Mattermost to latest version (dialog support improved in recent releases)

---

**Q: Images in signatures not displaying**

**A**:
- Image URLs in templates must be publicly accessible (HTTPS)
- Some email clients block external images by default
- Test in multiple clients: Outlook, Thunderbird, Gmail

---

### Runtime Issues

**Q: Welcome message disappears when marking steps complete**

**A**: This was fixed in the current version. If you still see it:
- Verify you're running the latest plugin version
- Check [`onboarding.go:273-290`](server/onboarding.go) includes welcome message reconstruction

---

**Q: Language setting doesn't change bot language**

**A**:
1. Verify setting is saved: System Console → Plugins → Onboarding Assistant → Language
2. Disable and re-enable plugin
3. Check `p.getPluginSetting("Language", "de")` returns correct value (add log)

---

**Q: Onboarding checklist sent multiple times to same user**

**A**:
- Check KV store for duplicate keys
- Verify `loadState()` is working correctly
- The plugin should be idempotent (won't re-send if state exists)

---

**Q: Pronoun formatting incorrect in signature**

**A**:
- Input must match format: `"er/ihm / he/him"` (with space-slash-space between German and English)
- Check `formatPronouns()` in [`signature_templates.go:96`](server/signature_templates.go)
- Test with various inputs to ensure robustness

---

## Summary: Quick Customization Checklist

Use this checklist when adapting the plugin for your organization:

- [ ] **Change plugin ID** in [`plugin.json`](plugin.json) and [`plugin.go:136`](server/plugin.go)
- [ ] **Update bot name and icon** in [`plugin.go:22-26`](server/plugin.go) and [`assets/icon.png`](assets/icon.png)
- [ ] **Customize onboarding steps** in [`model.go`](server/model.go) and [`onboarding.go`](server/onboarding.go)
- [ ] **Replace documentation links** in translation files ([`i18n_de.go`](server/i18n_de.go), [`i18n_en.go`](server/i18n_en.go))
- [ ] **Update signature templates** in [`signature_templates.go`](server/signature_templates.go) with your org's branding
- [ ] **Change project list** in [`signature.go`](server/signature.go) to match your departments
- [ ] **Translate all messages** to your organization's languages
- [ ] **Test thoroughly** by creating test users
- [ ] **Update version number** in [`plugin.json`](plugin.json) before each release

---

## License

This plugin is built for EOTO organization. Modify and distribute as needed for your organization.

---

## Support & Contributing

### Getting Help

1. **Check logs**: System Console → Logs (filter for "onboarding")
2. **Verify KV store**: Use `mmctl plugin kv` commands
3. **Test configuration**: Ensure `SiteURL` is set correctly
4. **Review this README**: Most issues are covered in Troubleshooting section

### Contributing

When making improvements:

1. **Test thoroughly** with both German and English language settings
2. **Verify signature templates** in multiple email clients
3. **Update translations** for both languages
4. **Document changes** in commit messages
5. **Increment version number** in `plugin.json`

---

**Happy onboarding! 🚀**

Built with ❤️ for EOTO using [Mattermost Plugin Framework](https://developers.mattermost.com/integrate/plugins/)
