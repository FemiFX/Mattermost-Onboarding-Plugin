package main

// translationsEN contains all English translations
var translationsEN = Translations{
	// Welcome message
	WelcomeGreeting: "👋 Hi %s, welcome to %s!",
	WelcomeIntro:    "I'm your onboarding assistant. I'll guide you through a few quick steps to get set up.",
	WelcomeClosing:  "_You can come back to this DM anytime to see your progress._",

	// Step titles
	Step1Title: "Step 1: Accounts & Access",
	Step2Title: "Step 2: Complete Your Profile",
	Step3Title: "Step 3: Communication Channels",
	Step4Title: "Step 4: Tools & Equipment",
	Step5Title: "Step 5: Work Practices & Policies",
	Step6Title: "Step 6: People & Check-ins",

	// Step descriptions
	Step1Description: "Make sure you can log in everywhere you need to:\n" +
		"- Google Workspace (EOTO email address issued & tested)\n" +
		"- Nextcloud (files & shared team folders)\n" +
		"- Timebutler (time tracking / attendance)\n" +
		"- Mattermost (you're here 🎉)\n" +
		"- Any role-specific tools (e.g. CRM, finance tools)\n\n" +
		"More details: ",

	Step2Description: "Help colleagues recognize and reach you easily:\n" +
		"- Upload a clear profile photo\n" +
		"- Add your full name and pronouns (if desired)\n" +
		"- Set your job title & department\n" +
		"- Configure your timezone and working hours\n" +
		"- Generate your email signature ✉️\n\n" +
		"Quick reference: ",

	Step3Description: "Join the spaces where information flows:\n" +
		"- `#announcements` — organization-wide updates\n" +
		"- `#helpdesk` — IT support & quick questions\n" +
		"- `#introductions` — say hello to everyone\n" +
		"- Your team / project channels (ask your manager)\n\n" +
		"Guidelines: ",

	Step4Description: "Confirm your hardware and core tools are ready:\n" +
		"- Laptop received, boots correctly, and you can log in\n" +
		"- Wi-Fi access at your usual work location(s)\n" +
		"- Nextcloud client installed (if required)\n" +
		"- Email & calendar working on your primary device\n" +
		"- Required VPN or remote access configured\n\n" +
		"See: ",

	Step5Description: "Take an initial pass through how we work at EOTO:\n" +
		"- Working hours, flextime, and vacation process\n" +
		"- Privacy & data protection basics (GDPR awareness)\n" +
		"- Communication expectations (response times, DM vs. channels)\n" +
		"- How we store and share files (Nextcloud structure)\n\n" +
		"Start here: ",

	Step6Description: "Make sure you're connected with the right people:\n" +
		"- Brief introduction post in `#introductions`\n" +
		"- 1:1 intro meeting with your manager (scheduled)\n" +
		"- Check-in with your onboarding buddy (if assigned)\n" +
		"- Add key people to your favorites in Mattermost\n\n" +
		"Tips: ",

	// Step links
	Step1Link: "[Accounts & Access Guide](https://outline.akinlosotu.tech)",
	Step2Link: "[Mattermost Profile & Notifications](https://outline.akinlosotu.tech)",
	Step3Link: "[Communication & Channels](https://outline.akinlosotu.tech)",
	Step4Link: "[Devices & IT Setup](https://outline.akinlosotu.tech)",
	Step5Link: "[EOTO Handbook](https://outline.akinlosotu.tech)",
	Step6Link: "[Onboarding & Collaboration at EOTO](https://outline.akinlosotu.tech)",

	// Button labels
	ButtonMarkAccountsReady:    "Mark Accounts Ready",
	ButtonMarkProfileComplete:  "Mark Profile Complete",
	ButtonMarkChannelsJoined:   "Mark Channels Joined",
	ButtonMarkToolsReady:       "Mark Tools Ready",
	ButtonMarkPoliciesReviewed: "Mark Policies Reviewed",
	ButtonMarkIntrosDone:       "Mark Intros Done",
	ButtonGenerateSignature:    "✉️ Generate Email Signature",

	// Signature dialog
	DialogSignatureTitle:        "Generate EOTO Email Signature",
	DialogSignatureIntro:        "Fill in your details to generate your EOTO email signature:",
	DialogFullName:              "Full Name",
	DialogFullNameHelp:          "Your full name as it should appear in the signature",
	DialogPosition:              "Position",
	DialogPositionHelp:          "Your job title or role at EOTO",
	DialogPronouns:              "Pronouns",
	DialogPronounsPlaceholder:   "she/her / sie/ihr",
	DialogPronounsHelp:          "Format: 'he/him / er/ihm' or 'she/her / sie/ihr' or 'No Pronouns / Keine Pronomen'",
	DialogEmail:                 "Email",
	DialogEmailHelp:             "Your EOTO email address",
	DialogProject:               "Project",
	DialogProjectHelp:           "Select the EOTO project you're working for",
	DialogWorkNumber:            "Work Number",
	DialogWorkNumberPlaceholder: "Tel.: 030 12345678",
	DialogWorkNumberHelp:        "Your work phone number (optional, include 'Tel.:' prefix)",
	DialogSubmitButton:          "Generate Signature",

	// Project names
	ProjectEachOne:    "Each One",
	ProjectCommunity:  "CommUnity",
	ProjectCUZ:        "CommUnity Zentrum (CUZ)",
	ProjectJugend:     "Youth Programs",
	ProjectNAR:        "Network-Antiracism (NAR)",
	ProjectAfrolution: "Afrolution",

	// Success messages
	SignatureGeneratedTitle: "✅ **EOTO Email Signature Successfully Generated!**",
	SignatureGeneratedMessage: "Hi %s, your email signature for **%s** is ready to use.\n\n" +
		"**How to use this signature:**\n" +
		"1. Download the HTML file below\n" +
		"2. Open it in a web browser\n" +
		"3. Select all content (Ctrl+A / Cmd+A)\n" +
		"4. Copy (Ctrl+C / Cmd+C)\n" +
		"5. Paste into your email client's signature settings\n\n",

	SignatureInstructionsTitle: "**For Outlook:**\n",
	SignatureInstructionsOutlook: "- Open Outlook → File → Options → Mail → Signatures\n" +
		"- Create a new signature, paste the copied content\n\n",

	SignatureInstructionsThunderbird: "**For Thunderbird:**\n" +
		"- Tools → Account Settings → Select your email → Attach signature from file\n" +
		"- Select the downloaded HTML file\n\n",

	StepMarkedComplete: "Marked step '%s' complete ✔️",
	DialogOpening:      "Opening EOTO signature generator...",

	// Error messages
	ErrorGeneral: "An error occurred. Please try again.",
}
