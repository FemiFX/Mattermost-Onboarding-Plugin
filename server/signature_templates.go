package main

import (
	"bytes"
	"html/template"
)

// SignatureData holds all the information needed to generate a signature
type SignatureData struct {
	FullName    string
	JobTitle    string
	Department  string
	Email       string
	Phone       string
	LinkedIn    string
	IncludeLogo bool
	Style       string
	CompanyName string
	CompanyURL  string
	LogoURL     string
}

// Template styles
const (
	professionalTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; }
        .signature { max-width: 500px; }
        .name { font-size: 18px; font-weight: bold; color: #1e3a8a; margin: 0; }
        .title { font-size: 14px; color: #64748b; margin: 5px 0; }
        .contact { font-size: 13px; color: #475569; margin: 3px 0; }
        .contact a { color: #1e3a8a; text-decoration: none; }
        .contact a:hover { text-decoration: underline; }
        .divider { border-top: 2px solid #1e3a8a; margin: 10px 0; }
        .company { font-size: 14px; font-weight: bold; color: #1e3a8a; margin-top: 10px; }
        .logo { margin-top: 10px; }
        .logo img { max-width: 150px; height: auto; }
    </style>
</head>
<body>
    <div class="signature">
        <p class="name">{{.FullName}}</p>
        <p class="title">{{.JobTitle}}{{if .Department}} | {{.Department}}{{end}}</p>
        <div class="divider"></div>
        <p class="contact">📧 <a href="mailto:{{.Email}}">{{.Email}}</a></p>
        {{if .Phone}}<p class="contact">📱 {{.Phone}}</p>{{end}}
        {{if .LinkedIn}}<p class="contact">🔗 <a href="{{.LinkedIn}}" target="_blank">LinkedIn</a></p>{{end}}
        {{if .IncludeLogo}}
        <div class="logo">
            <img src="{{.LogoURL}}" alt="{{.CompanyName}}" />
        </div>
        {{end}}
        <p class="company"><a href="{{.CompanyURL}}" style="color: #1e3a8a; text-decoration: none;">{{.CompanyName}}</a></p>
    </div>
</body>
</html>
`

	modernTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; }
        .signature { max-width: 500px; background: linear-gradient(to right, #f0fdf4, #ffffff); padding: 20px; border-left: 4px solid #10b981; }
        .name { font-size: 20px; font-weight: bold; color: #065f46; margin: 0; }
        .title { font-size: 14px; color: #059669; margin: 5px 0; font-weight: 500; }
        .contact { font-size: 13px; color: #374151; margin: 5px 0; }
        .contact a { color: #10b981; text-decoration: none; font-weight: 500; }
        .contact a:hover { text-decoration: underline; }
        .company { font-size: 14px; color: #065f46; margin-top: 15px; font-weight: bold; }
        .logo { margin-top: 15px; }
        .logo img { max-width: 150px; height: auto; }
        .icon { display: inline-block; width: 20px; text-align: center; }
    </style>
</head>
<body>
    <div class="signature">
        <p class="name">{{.FullName}}</p>
        <p class="title">{{.JobTitle}}{{if .Department}} · {{.Department}}{{end}}</p>
        <div style="margin: 10px 0;"></div>
        <p class="contact"><span class="icon">✉️</span> <a href="mailto:{{.Email}}">{{.Email}}</a></p>
        {{if .Phone}}<p class="contact"><span class="icon">📞</span> {{.Phone}}</p>{{end}}
        {{if .LinkedIn}}<p class="contact"><span class="icon">💼</span> <a href="{{.LinkedIn}}" target="_blank">Connect on LinkedIn</a></p>{{end}}
        {{if .IncludeLogo}}
        <div class="logo">
            <img src="{{.LogoURL}}" alt="{{.CompanyName}}" />
        </div>
        {{end}}
        <p class="company"><a href="{{.CompanyURL}}" style="color: #065f46; text-decoration: none;">{{.CompanyName}}</a></p>
    </div>
</body>
</html>
`

	minimalistTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; }
        .signature { max-width: 500px; border-top: 1px solid #000000; padding-top: 10px; }
        .name { font-size: 16px; font-weight: 600; color: #000000; margin: 5px 0; }
        .title { font-size: 13px; color: #666666; margin: 2px 0; }
        .contact { font-size: 12px; color: #333333; margin: 2px 0; }
        .contact a { color: #000000; text-decoration: none; }
        .contact a:hover { text-decoration: underline; }
        .separator { color: #cccccc; margin: 0 5px; }
        .company { font-size: 13px; color: #000000; margin-top: 10px; }
        .logo { margin-top: 10px; }
        .logo img { max-width: 120px; height: auto; filter: grayscale(100%); }
    </style>
</head>
<body>
    <div class="signature">
        <p class="name">{{.FullName}}</p>
        <p class="title">{{.JobTitle}}{{if .Department}} / {{.Department}}{{end}}</p>
        <p class="contact">
            <a href="mailto:{{.Email}}">{{.Email}}</a>
            {{if .Phone}}<span class="separator">|</span>{{.Phone}}{{end}}
            {{if .LinkedIn}}<span class="separator">|</span><a href="{{.LinkedIn}}" target="_blank">LinkedIn</a>{{end}}
        </p>
        {{if .IncludeLogo}}
        <div class="logo">
            <img src="{{.LogoURL}}" alt="{{.CompanyName}}" />
        </div>
        {{end}}
        <p class="company"><a href="{{.CompanyURL}}" style="color: #000000; text-decoration: none;">{{.CompanyName}}</a></p>
    </div>
</body>
</html>
`
)

// GenerateSignature creates an HTML email signature based on the provided data and style
func GenerateSignature(data SignatureData) (string, error) {
	// Set default company info if not provided
	if data.CompanyName == "" {
		data.CompanyName = "EOTO"
	}
	if data.CompanyURL == "" {
		data.CompanyURL = "https://akinlosotu.tech"
	}
	if data.LogoURL == "" {
		data.LogoURL = "https://akinlosotu.tech/logo.png"
	}

	// Select template based on style
	var templateStr string
	switch data.Style {
	case "modern":
		templateStr = modernTemplate
	case "minimalist":
		templateStr = minimalistTemplate
	default:
		templateStr = professionalTemplate
	}

	// Parse and execute template
	tmpl, err := template.New("signature").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
