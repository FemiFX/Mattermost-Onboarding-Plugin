package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

// handleSignatureDialog opens an interactive dialog for EOTO signature generation
func (p *Plugin) handleSignatureDialog(w http.ResponseWriter, r *http.Request, req *model.PostActionIntegrationRequest) {
	// Get user info to pre-fill form
	user, appErr := p.API.GetUser(req.UserId)
	if appErr != nil {
		p.API.LogError("failed to get user", "err", appErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	callbackURL, err := p.pluginURL()
	if err != nil {
		p.API.LogError("pluginURL error", "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dialog := model.OpenDialogRequest{
		TriggerId: req.TriggerId,
		URL:       callbackURL + "/submit-signature",
		Dialog: model.Dialog{
			Title:            "Generate EOTO Email Signature",
			IntroductionText: "Fill in your details to generate your EOTO email signature:",
			Elements: []model.DialogElement{
				{
					DisplayName: "Full Name",
					Name:        "full_name",
					Type:        "text",
					Placeholder: "Max Mustermann",
					Default:     user.GetFullName(),
					HelpText:    "Your complete name as it should appear in the signature",
				},
				{
					DisplayName: "Position",
					Name:        "position",
					Type:        "text",
					Placeholder: "Projektkoordinator*in",
					HelpText:    "Your job title or role at EOTO",
				},
				{
					DisplayName: "Pronouns",
					Name:        "pronouns",
					Type:        "text",
					Placeholder: "er/ihm / he/him",
					Optional:    true,
					HelpText:    "Format: 'er/ihm / he/him' or 'sie/ihr / she/her' or 'Keine Pronomen / No Pronouns'",
				},
				{
					DisplayName: "Email",
					Name:        "email",
					Type:        "text",
					SubType:     "email",
					Default:     user.Email,
					HelpText:    "Your EOTO email address",
				},
				{
					DisplayName: "Project",
					Name:        "project",
					Type:        "select",
					HelpText:    "Select the EOTO project you work for",
					Options: []*model.PostActionOptions{
						{Text: "Each One", Value: "each-one"},
						{Text: "CommUnity", Value: "community"},
						{Text: "CommUnity Zentrum (CUZ)", Value: "cuz"},
						{Text: "Jugendangebote", Value: "jugend"},
						{Text: "Netzwerk-Antirassismus (NAR)", Value: "nar"},
						{Text: "Afrolution", Value: "afrolution"},
					},
					Default: "each-one",
				},
				{
					DisplayName: "Work Number",
					Name:        "work_number",
					Type:        "text",
					SubType:     "tel",
					Optional:    true,
					Placeholder: "Tel.: 030 12345678",
					HelpText:    "Your work phone number (optional, include 'Tel.:' prefix)",
				},
			},
			SubmitLabel:    "Generate Signature",
			NotifyOnCancel: false,
		},
	}

	if appErr := p.API.OpenInteractiveDialog(dialog); appErr != nil {
		p.API.LogError("failed to open dialog", "err", appErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return success response
	resp := &model.PostActionIntegrationResponse{
		EphemeralText: "Opening EOTO signature generator...",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSignatureSubmission processes the dialog submission and generates the EOTO signature
func (p *Plugin) handleSignatureSubmission(w http.ResponseWriter, r *http.Request) {
	var submission model.SubmitDialogRequest
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		p.API.LogError("failed to decode dialog submission", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := submission.UserId
	channelID := submission.ChannelId

	// Extract form data
	fullName, _ := submission.Submission["full_name"].(string)
	position, _ := submission.Submission["position"].(string)
	pronouns, _ := submission.Submission["pronouns"].(string)
	email, _ := submission.Submission["email"].(string)
	project, _ := submission.Submission["project"].(string)
	workNumber, _ := submission.Submission["work_number"].(string)

	// Validate required fields
	errors := make(map[string]string)
	if fullName == "" {
		errors["full_name"] = "Full name is required"
	}
	if position == "" {
		errors["position"] = "Position is required"
	}
	if email == "" {
		errors["email"] = "Email is required"
	}
	if project == "" {
		errors["project"] = "Project is required"
	}

	if len(errors) > 0 {
		resp := &model.SubmitDialogResponse{
			Errors: errors,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Generate signature using EOTO templates
	signatureData := SignatureData{
		FullName:   fullName,
		Position:   position,
		Pronouns:   pronouns,
		Email:      email,
		Project:    project,
		WorkNumber: workNumber,
	}

	signatureHTML, err := GenerateSignature(signatureData)
	if err != nil {
		p.API.LogError("failed to generate signature", "err", err.Error())

		resp := &model.SubmitDialogResponse{
			Error: "Failed to generate signature. Please try again.",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Upload HTML file to Mattermost
	fileID, err := p.uploadSignatureFile(userID, channelID, signatureHTML, fullName, project)
	if err != nil {
		p.API.LogError("failed to upload signature file", "err", err.Error())

		resp := &model.SubmitDialogResponse{
			Error: "Failed to upload signature file. Please try again.",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Get DM channel with bot
	dmChannel, appErr := p.API.GetDirectChannel(p.botUserID, userID)
	if appErr != nil {
		p.API.LogError("failed to get DM channel", "err", appErr.Error())
		dmChannel = &model.Channel{Id: channelID} // Fallback to current channel
	}

	// Post message with download link
	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: dmChannel.Id,
		Message: fmt.Sprintf(
			"✅ **EOTO Email Signature Generated Successfully!**\n\n"+
				"Hi %s, your email signature for **%s** is ready to use.\n\n"+
				"**To use this signature:**\n"+
				"1. Download the HTML file below\n"+
				"2. Open it in a web browser\n"+
				"3. Select all content (Ctrl+A / Cmd+A)\n"+
				"4. Copy (Ctrl+C / Cmd+C)\n"+
				"5. Paste into your email client's signature settings\n\n"+
				"**For Outlook:**\n"+
				"- Open Outlook → File → Options → Mail → Signatures\n"+
				"- Create new signature, paste the copied content\n\n"+
				"**For Thunderbird:**\n"+
				"- Tools → Account Settings → Select your email → Attach signature from file\n"+
				"- Choose the downloaded HTML file\n\n"+
				"_Project: %s_",
			fullName,
			formatProjectName(project),
			formatProjectName(project),
		),
		FileIds: []string{fileID},
	}

	if _, appErr := p.API.CreatePost(post); appErr != nil {
		p.API.LogError("failed to create post", "err", appErr.Error())
	}

	// Return success (dialog will close)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&model.SubmitDialogResponse{})
}

// uploadSignatureFile uploads the generated signature HTML to Mattermost
func (p *Plugin) uploadSignatureFile(userID, channelID string, htmlContent, fullName, project string) (string, error) {
	// Create filename matching Python app format: {name}_{project}_Signatur.html
	sanitizedName := strings.ReplaceAll(fullName, " ", "_")
	filename := fmt.Sprintf("%s_%s_Signatur.html", sanitizedName, project)

	// Upload file
	fileInfo, appErr := p.API.UploadFile(
		[]byte(htmlContent),
		channelID,
		filename,
	)
	if appErr != nil {
		return "", appErr
	}

	return fileInfo.Id, nil
}

// formatProjectName makes the project name more readable
func formatProjectName(project string) string {
	switch project {
	case "each-one":
		return "Each One"
	case "community":
		return "CommUnity"
	case "cuz":
		return "CommUnity Zentrum (CUZ)"
	case "jugend":
		return "Jugendangebote"
	case "nar":
		return "Netzwerk-Antirassismus (NAR)"
	case "afrolution":
		return "Afrolution"
	default:
		return "Each One"
	}
}
