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

	// Get translations
	tr := p.getTranslations()

	dialog := model.OpenDialogRequest{
		TriggerId: req.TriggerId,
		URL:       callbackURL + "/submit-signature",
		Dialog: model.Dialog{
			Title:            tr.DialogSignatureTitle,
			IntroductionText: tr.DialogSignatureIntro,
			Elements: []model.DialogElement{
				{
					DisplayName: tr.DialogFullName,
					Name:        "full_name",
					Type:        "text",
					Placeholder: "Max Mustermann",
					Default:     user.GetFullName(),
					HelpText:    tr.DialogFullNameHelp,
				},
				{
					DisplayName: tr.DialogPosition,
					Name:        "position",
					Type:        "text",
					Placeholder: "Projektkoordinator*in",
					HelpText:    tr.DialogPositionHelp,
				},
				{
					DisplayName: tr.DialogPronouns,
					Name:        "pronouns",
					Type:        "text",
					Placeholder: tr.DialogPronounsPlaceholder,
					Optional:    true,
					HelpText:    tr.DialogPronounsHelp,
				},
				{
					DisplayName: tr.DialogEmail,
					Name:        "email",
					Type:        "text",
					SubType:     "email",
					Default:     user.Email,
					HelpText:    tr.DialogEmailHelp,
				},
				{
					DisplayName: tr.DialogProject,
					Name:        "project",
					Type:        "select",
					HelpText:    tr.DialogProjectHelp,
					Options: []*model.PostActionOptions{
						{Text: tr.ProjectEachOne, Value: "each-one"},
						{Text: tr.ProjectCommunity, Value: "community"},
						{Text: tr.ProjectCUZ, Value: "cuz"},
						{Text: tr.ProjectJugend, Value: "jugend"},
						{Text: tr.ProjectNAR, Value: "nar"},
						{Text: tr.ProjectAfrolution, Value: "afrolution"},
					},
					Default: "each-one",
				},
				{
					DisplayName: tr.DialogWorkNumber,
					Name:        "work_number",
					Type:        "text",
					SubType:     "tel",
					Optional:    true,
					Placeholder: tr.DialogWorkNumberPlaceholder,
					HelpText:    tr.DialogWorkNumberHelp,
				},
			},
			SubmitLabel:    tr.DialogSubmitButton,
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
		EphemeralText: tr.DialogOpening,
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

	// Get translations
	tr := p.getTranslations()

	// Post message with download link
	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: dmChannel.Id,
		Message: tr.SignatureGeneratedTitle + "\n\n" +
			fmt.Sprintf(tr.SignatureGeneratedMessage, fullName, formatProjectName(project, &tr)) +
			tr.SignatureInstructionsTitle +
			tr.SignatureInstructionsOutlook +
			tr.SignatureInstructionsThunderbird +
			fmt.Sprintf("_Project: %s_", formatProjectName(project, &tr)),
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

// formatProjectName makes the project name more readable using translations
func formatProjectName(project string, tr *Translations) string {
	switch project {
	case "each-one":
		return tr.ProjectEachOne
	case "community":
		return tr.ProjectCommunity
	case "cuz":
		return tr.ProjectCUZ
	case "jugend":
		return tr.ProjectJugend
	case "nar":
		return tr.ProjectNAR
	case "afrolution":
		return tr.ProjectAfrolution
	default:
		return tr.ProjectEachOne
	}
}
