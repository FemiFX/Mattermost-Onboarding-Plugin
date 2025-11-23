package main

// Translations contains all user-facing text for the onboarding plugin
type Translations struct {
	// Welcome message
	WelcomeGreeting string
	WelcomeIntro    string
	WelcomeClosing  string

	// Step titles
	Step1Title string
	Step2Title string
	Step3Title string
	Step4Title string
	Step5Title string
	Step6Title string

	// Step descriptions
	Step1Description string
	Step2Description string
	Step3Description string
	Step4Description string
	Step5Description string
	Step6Description string

	// Step links
	Step1Link string
	Step2Link string
	Step3Link string
	Step4Link string
	Step5Link string
	Step6Link string

	// Button labels
	ButtonMarkAccountsReady  string
	ButtonMarkProfileComplete string
	ButtonMarkChannelsJoined string
	ButtonMarkToolsReady     string
	ButtonMarkPoliciesReviewed string
	ButtonMarkIntrosDone     string
	ButtonGenerateSignature  string

	// Signature dialog
	DialogSignatureTitle      string
	DialogSignatureIntro      string
	DialogFullName            string
	DialogFullNameHelp        string
	DialogPosition            string
	DialogPositionHelp        string
	DialogPronouns            string
	DialogPronounsPlaceholder string
	DialogPronounsHelp        string
	DialogEmail               string
	DialogEmailHelp           string
	DialogProject             string
	DialogProjectHelp         string
	DialogWorkNumber          string
	DialogWorkNumberPlaceholder string
	DialogWorkNumberHelp      string
	DialogSubmitButton        string

	// Project names in dialog
	ProjectEachOne   string
	ProjectCommunity string
	ProjectCUZ       string
	ProjectJugend    string
	ProjectNAR       string
	ProjectAfrolution string

	// Success messages
	SignatureGeneratedTitle   string
	SignatureGeneratedMessage string
	SignatureInstructionsTitle string
	SignatureInstructionsOutlook string
	SignatureInstructionsThunderbird string
	StepMarkedComplete        string
	DialogOpening             string

	// Error messages
	ErrorGeneral string
}

// getTranslations returns the appropriate translation set based on plugin config
func (p *Plugin) getTranslations() Translations {
	// Get language from plugin settings (default to German)
	language := p.getPluginSetting("Language", "de")

	switch language {
	case "en":
		return translationsEN
	case "de":
		return translationsDE
	default:
		return translationsDE // Default to German
	}
}

// getPluginSetting retrieves a plugin configuration setting
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
