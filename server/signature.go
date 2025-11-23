package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

// handleSignatureDialog opens an interactive dialog for signature generation
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
			Title:            "Generate Email Signature",
			IntroductionText: "Fill in your details to generate your professional email signature:",
			Elements: []model.DialogElement{
				{
					DisplayName: "Full Name",
					Name:        "full_name",
					Type:        "text",
					Placeholder: "John Doe",
					Default:     user.GetFullName(),
				},
				{
					DisplayName: "Job Title",
					Name:        "job_title",
					Type:        "text",
					Placeholder: "Senior Software Engineer",
					HelpText:    "Your role at EOTO",
				},
				{
					DisplayName: "Department",
					Name:        "department",
					Type:        "select",
					Options: []*model.PostActionOptions{
						{Text: "Engineering", Value: "Engineering"},
						{Text: "Product", Value: "Product"},
						{Text: "Design", Value: "Design"},
						{Text: "Marketing", Value: "Marketing"},
						{Text: "Sales", Value: "Sales"},
						{Text: "Operations", Value: "Operations"},
						{Text: "HR", Value: "HR"},
						{Text: "Finance", Value: "Finance"},
						{Text: "Executive", Value: "Executive"},
					},
				},
				{
					DisplayName: "Email",
					Name:        "email",
					Type:        "text",
					SubType:     "email",
					Default:     user.Email,
				},
				{
					DisplayName: "Phone Number",
					Name:        "phone",
					Type:        "text",
					SubType:     "tel",
					Optional:    true,
					Placeholder: "+234 123 456 7890",
				},
				{
					DisplayName: "LinkedIn Profile",
					Name:        "linkedin",
					Type:        "text",
					SubType:     "url",
					Optional:    true,
					Placeholder: "https://linkedin.com/in/yourprofile",
				},
				{
					DisplayName: "Include Company Logo",
					Name:        "include_logo",
					Type:        "bool",
					Default:     "true",
					HelpText:    "Add EOTO logo to signature",
				},
				{
					DisplayName: "Signature Style",
					Name:        "style",
					Type:        "radio",
					Options: []*model.PostActionOptions{
						{Text: "Professional (Blue)", Value: "professional"},
						{Text: "Modern (Green)", Value: "modern"},
						{Text: "Minimalist (Black & White)", Value: "minimalist"},
					},
					Default: "professional",
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
		EphemeralText: "Opening signature generator...",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSignatureSubmission processes the dialog submission and generates the signature
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
	jobTitle, _ := submission.Submission["job_title"].(string)
	department, _ := submission.Submission["department"].(string)
	email, _ := submission.Submission["email"].(string)
	phone, _ := submission.Submission["phone"].(string)
	linkedin, _ := submission.Submission["linkedin"].(string)
	includeLogoStr, _ := submission.Submission["include_logo"].(string)
	style, _ := submission.Submission["style"].(string)

	// Validate required fields
	if fullName == "" || jobTitle == "" || email == "" {
		resp := &model.SubmitDialogResponse{
			Errors: map[string]string{
				"full_name": "Full name is required",
				"job_title": "Job title is required",
				"email":     "Email is required",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	includeLogo := includeLogoStr == "true"

	// Generate signature
	signatureData := SignatureData{
		FullName:    fullName,
		JobTitle:    jobTitle,
		Department:  department,
		Email:       email,
		Phone:       phone,
		LinkedIn:    linkedin,
		IncludeLogo: includeLogo,
		Style:       style,
		CompanyName: "EOTO",
		CompanyURL:  "https://akinlosotu.tech",
		LogoURL:     "https://akinlosotu.tech/logo.png",
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
	fileID, err := p.uploadSignatureFile(userID, channelID, signatureHTML, fullName)
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
			"✅ **Email Signature Generated Successfully!**\n\n"+
				"Hi %s, your email signature is ready to use.\n\n"+
				"**To use this signature:**\n"+
				"1. Download the HTML file below\n"+
				"2. Open it in a web browser\n"+
				"3. Select all content (Ctrl+A / Cmd+A)\n"+
				"4. Copy (Ctrl+C / Cmd+C)\n"+
				"5. Paste into your email client's signature settings\n\n"+
				"_Style: %s_",
			fullName,
			formatStyleName(style),
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
func (p *Plugin) uploadSignatureFile(userID, channelID string, htmlContent, fullName string) (string, error) {
	// Create filename
	sanitizedName := strings.ReplaceAll(strings.ToLower(fullName), " ", "_")
	filename := fmt.Sprintf("%s_email_signature.html", sanitizedName)

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

// formatStyleName makes the style name more readable
func formatStyleName(style string) string {
	switch style {
	case "modern":
		return "Modern (Green)"
	case "minimalist":
		return "Minimalist (Black & White)"
	default:
		return "Professional (Blue)"
	}
}
