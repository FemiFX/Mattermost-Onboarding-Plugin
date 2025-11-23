package main

// translationsDE contains all German translations
var translationsDE = Translations{
	// Welcome message
	WelcomeGreeting: "👋 Hallo %s, willkommen bei %s!",
	WelcomeIntro:    "Ich bin dein Onboarding-Assistent. Ich führe dich durch ein paar schnelle Schritte, um dich einzurichten.",
	WelcomeClosing:  "_Du kannst jederzeit zu dieser DM zurückkehren, um deinen Fortschritt zu sehen._",

	// Step titles
	Step1Title: "Schritt 1: Konten & Zugang",
	Step2Title: "Schritt 2: Vervollständige dein Profil",
	Step3Title: "Schritt 3: Kommunikationskanäle",
	Step4Title: "Schritt 4: Tools & Ausrüstung",
	Step5Title: "Schritt 5: Arbeitsweisen & Richtlinien",
	Step6Title: "Schritt 6: Menschen & Check-ins",

	// Step descriptions
	Step1Description: "Stelle sicher, dass du dich überall anmelden kannst, wo du es benötigst:\n" +
		"- Google Workspace (EOTO E-Mail-Adresse ausgegeben & getestet)\n" +
		"- Nextcloud (Dateien & gemeinsame Team-Ordner)\n" +
		"- Timebutler (Zeiterfassung / Anwesenheit)\n" +
		"- Mattermost (du bist hier 🎉)\n" +
		"- Alle rollenspezifischen Tools (z.B. CRM, Finanztools)\n\n" +
		"Mehr Details: ",

	Step2Description: "Hilf Kollegen, dich leicht zu erkennen und zu erreichen:\n" +
		"- Lade ein klares Profilfoto hoch\n" +
		"- Füge deinen vollständigen Namen und Pronomen hinzu (falls gewünscht)\n" +
		"- Lege deinen Jobtitel & deine Abteilung fest\n" +
		"- Stelle deine Zeitzone und Arbeitszeiten ein\n" +
		"- Generiere deine E-Mail-Signatur ✉️\n\n" +
		"Schnellreferenz: ",

	Step3Description: "Tritt den Räumen bei, in denen Informationen fließen:\n" +
		"- `#announcements` — organisationsweite Updates\n" +
		"- `#helpdesk` — IT-Support & schnelle Fragen\n" +
		"- `#introductions` — sag allen Hallo\n" +
		"- Deine Team- / Projektkanäle (frage deinen Manager)\n\n" +
		"Richtlinien: ",

	Step4Description: "Bestätige, dass deine Hardware und Kerntools bereit sind:\n" +
		"- Laptop erhalten, startet korrekt und du kannst dich anmelden\n" +
		"- WLAN-Zugang an deinem üblichen Arbeitsort(en)\n" +
		"- Nextcloud-Client installiert (falls erforderlich)\n" +
		"- E-Mail & Kalender funktionieren auf deinem Hauptgerät\n" +
		"- Erforderliches VPN oder Fernzugriff konfiguriert\n\n" +
		"Siehe: ",

	Step5Description: "Mache einen ersten Durchgang durch die Arbeitsweise bei EOTO:\n" +
		"- Arbeitszeiten, Gleitzeit und Urlaubsprozess\n" +
		"- Datenschutz & Datenschutz-Grundlagen (DSGVO-Bewusstsein)\n" +
		"- Kommunikationserwartungen (Antwortzeiten, DM vs. Kanäle)\n" +
		"- Wie wir Dateien speichern und teilen (Nextcloud-Struktur)\n\n" +
		"Beginne hier: ",

	Step6Description: "Stelle sicher, dass du mit den richtigen Menschen verbunden bist:\n" +
		"- Kurzer Vorstellungsbeitrag in `#introductions`\n" +
		"- 1:1-Vorstellung mit deinem Manager (geplant)\n" +
		"- Check-in mit deinem Onboarding-Buddy (falls zugewiesen)\n" +
		"- Füge wichtige Personen zu deinen Favoriten in Mattermost hinzu\n\n" +
		"Tipps: ",

	// Step links
	Step1Link: "[Konten & Zugang Leitfaden](https://outline.akinlosotu.tech)",
	Step2Link: "[Mattermost Profil & Benachrichtigungen](https://outline.akinlosotu.tech)",
	Step3Link: "[Kommunikation & Kanäle](https://outline.akinlosotu.tech)",
	Step4Link: "[Geräte & IT-Einrichtung](https://outline.akinlosotu.tech)",
	Step5Link: "[EOTO Handbuch](https://outline.akinlosotu.tech)",
	Step6Link: "[Onboarding & Zusammenarbeit bei EOTO](https://outline.akinlosotu.tech)",

	// Button labels
	ButtonMarkAccountsReady:    "Konten bereit markieren",
	ButtonMarkProfileComplete:  "Profil vollständig markieren",
	ButtonMarkChannelsJoined:   "Kanäle beigetreten markieren",
	ButtonMarkToolsReady:       "Tools bereit markieren",
	ButtonMarkPoliciesReviewed: "Richtlinien überprüft markieren",
	ButtonMarkIntrosDone:       "Vorstellungen erledigt markieren",
	ButtonGenerateSignature:    "✉️ E-Mail-Signatur generieren",

	// Signature dialog
	DialogSignatureTitle:        "EOTO E-Mail-Signatur generieren",
	DialogSignatureIntro:        "Fülle deine Details aus, um deine EOTO E-Mail-Signatur zu generieren:",
	DialogFullName:              "Vollständiger Name",
	DialogFullNameHelp:          "Dein vollständiger Name, wie er in der Signatur erscheinen soll",
	DialogPosition:              "Position",
	DialogPositionHelp:          "Dein Jobtitel oder deine Rolle bei EOTO",
	DialogPronouns:              "Pronomen",
	DialogPronounsPlaceholder:   "er/ihm / he/him",
	DialogPronounsHelp:          "Format: 'er/ihm / he/him' oder 'sie/ihr / she/her' oder 'Keine Pronomen / No Pronouns'",
	DialogEmail:                 "E-Mail",
	DialogEmailHelp:             "Deine EOTO E-Mail-Adresse",
	DialogProject:               "Projekt",
	DialogProjectHelp:           "Wähle das EOTO-Projekt aus, für das du arbeitest",
	DialogWorkNumber:            "Arbeitsnummer",
	DialogWorkNumberPlaceholder: "Tel.: 030 12345678",
	DialogWorkNumberHelp:        "Deine Arbeitstelefonnummer (optional, füge 'Tel.:' Präfix hinzu)",
	DialogSubmitButton:          "Signatur generieren",

	// Project names
	ProjectEachOne:    "Each One",
	ProjectCommunity:  "CommUnity",
	ProjectCUZ:        "CommUnity Zentrum (CUZ)",
	ProjectJugend:     "Jugendangebote",
	ProjectNAR:        "Netzwerk-Antirassismus (NAR)",
	ProjectAfrolution: "Afrolution",

	// Success messages
	SignatureGeneratedTitle: "✅ **EOTO E-Mail-Signatur erfolgreich generiert!**",
	SignatureGeneratedMessage: "Hallo %s, deine E-Mail-Signatur für **%s** ist einsatzbereit.\n\n" +
		"**So verwendest du diese Signatur:**\n" +
		"1. Lade die HTML-Datei unten herunter\n" +
		"2. Öffne sie in einem Webbrowser\n" +
		"3. Wähle den gesamten Inhalt aus (Strg+A / Cmd+A)\n" +
		"4. Kopieren (Strg+C / Cmd+C)\n" +
		"5. Füge in die Signatureinstellungen deines E-Mail-Clients ein\n\n",

	SignatureInstructionsTitle: "**Für Outlook:**\n",
	SignatureInstructionsOutlook: "- Öffne Outlook → Datei → Optionen → E-Mail → Signaturen\n" +
		"- Erstelle eine neue Signatur, füge den kopierten Inhalt ein\n\n",

	SignatureInstructionsThunderbird: "**Für Thunderbird:**\n" +
		"- Extras → Konten-Einstellungen → Wähle deine E-Mail → Signatur aus Datei anhängen\n" +
		"- Wähle die heruntergeladene HTML-Datei\n\n",

	StepMarkedComplete: "Schritt '%s' als erledigt markiert ✔️",
	DialogOpening:      "EOTO Signaturgenerator wird geöffnet...",

	// Error messages
	ErrorGeneral: "Ein Fehler ist aufgetreten. Bitte versuche es erneut.",
}
