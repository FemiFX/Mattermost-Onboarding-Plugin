package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// SignatureData holds all the information needed to generate an EOTO signature
type SignatureData struct {
	FullName   string
	Position   string
	Pronouns   string
	Email      string
	Project    string
	WorkNumber string
}

// EOTO Project-specific signature templates
const (
	eachOneTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<table border="0">
  <tbody>
    <tr>
      <td bgcolor="#FFFFFF">
        <a href="https://eoto-archiv.de" target="blank">
          <img src="https://mailserver.eoto-archiv.de/EOTO_Logo_1.png" alt="EOTO-Logo-ohne-e-V" height="50px" vspace="0" hspace="10" />
        </a>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
          <strong>{{.FullName}}</strong><br>
          {{.Pronouns}}<br>
          {{.Position}} - Each One<br>
          Each One Teach One (EOTO) e.V.<br><br>
          Kamerunerstraße 16 | 13351 Berlin<br>
          {{.WorkNumber}}
          Email: <a href="mailto:{{.Email}}" style="color: #05576d; text-decoration:none;">{{.Email}}</a><br>
          <a href="http://each-one.de" style="color: #05576d; text-decoration:none;">Web: www.each-one.de</a>
          <a href="http://eoto-archiv.de" style="color: #05576d; text-decoration:none;">Web: www.eoto-archiv.de</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #E08800; font-weight: regular; font-size: 12px; line-height: 18px; padding: 0px 0px 0px 10px;">
            <a href="https://www.facebook.com/EOTO.eV" style="color: #E08800; text-decoration:none;">Facebook</a> |
            <a href="https://www.instagram.com/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Instagram</a> |
            <a href="https://www.pinterest.de/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Pinterest</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
            Amtsgericht Charlottenburg<br>
            Registernummer: VR 31576B<br>
            Vorstand: Daniel Gyamerah, Susanna Steinbach
          </p>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px; padding: 5px 0px 0px 10px;">
          <strong>Unterstützen Sie unsere Arbeit mit Ihrer Spende.</strong><br><br>
          Spendenkonto:<br>
          Each One Teach One (EOTO) e.V. | GLS Bank<br>
          IBAN: DE24 4306 0967 1153 5692 00<br>
          BIC: GENODEM1GLS<br><br>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <img src="https://mailserver.eoto-archiv.de/IDPAD-Logo-E.png" height="50px" alt="International Decade for People of African Descent" vspace="0" hspace="10" />
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 8px; line-height: auto; padding: 0px 0px 0px 10px;">
          Each One Teach One (EOTO) e.V. supports the International Decade for People of African Descent.<br>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px"></p>
        </p>
      </td>
    </tr>
    <tr>
      <td height="30"></td>
    </tr>
  </tbody>
</table>`

	communityTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<table border="0">
  <tbody>
    <tr>
      <td bgcolor="#FFFFFF">
        <a href="https://eoto-archiv.de" target="blank">
          <img src="https://mailserver.eoto-archiv.de/EOTO_Logo_1.png" alt="EOTO-Logo-ohne-e-V" height="50px" vspace="0" hspace="10" />
        </a>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
          <strong>{{.FullName}}</strong><br>
          {{.Pronouns}}<br>
          {{.Position}} - CommUnity<br>
          Each One Teach One (EOTO) e.V.<br><br>
          Kamerunerstraße 16 | 13351 Berlin<br>
          {{.WorkNumber}}
          Email: <a href="mailto:{{.Email}}" style="color: #05576d; text-decoration:none;">{{.Email}}</a><br>
          <a href="http://eoto-archiv.de" style="color: #05576d; text-decoration:none;">Web: www.eoto-archiv.de</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #E08800; font-weight: regular; font-size: 12px; line-height: 18px; padding: 0px 0px 0px 10px;">
            <a href="https://www.facebook.com/EOTO.eV" style="color: #E08800; text-decoration:none;">Facebook</a> |
            <a href="https://www.instagram.com/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Instagram</a> |
            <a href="https://www.pinterest.de/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Pinterest</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
            Amtsgericht Charlottenburg<br>
            Registernummer: VR 31576B<br>
            Vorstand: Daniel Gyamerah, Susanna Steinbach
          </p>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px; padding: 5px 0px 0px 10px;">
          <strong>Unterstützen Sie unsere Arbeit mit Ihrer Spende.</strong><br><br>
          Spendenkonto:<br>
          Each One Teach One (EOTO) e.V. | GLS Bank<br>
          IBAN: DE24 4306 0967 1153 5692 00<br>
          BIC: GENODEM1GLS<br><br>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <img src="https://mailserver.eoto-archiv.de/IDPAD-Logo-E.png" height="50px" alt="International Decade for People of African Descent" vspace="0" hspace="10" />
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 8px; line-height: auto; padding: 0px 0px 0px 10px;">
          Each One Teach One (EOTO) e.V. supports the International Decade for People of African Descent.<br>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px"></p>
        </p>
      </td>
    </tr>
    <tr>
      <td height="30"></td>
    </tr>
  </tbody>
</table>`

	cuzTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<table border="0">
  <tbody>
    <tr>
      <td bgcolor="#FFFFFF">
        <a href="https://eoto-archiv.de" target="blank">
          <img src="https://mailserver.eoto-archiv.de/EOTO_Logo_1.png" alt="EOTO-Logo-ohne-e-V" height="50px" vspace="0" hspace="10" />
        </a>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
          <strong>{{.FullName}}</strong><br>
          {{.Pronouns}}<br>
          {{.Position}} - CommUnity Zentrum (CUZ)<br>
          Each One Teach One (EOTO) e.V.<br><br>
          Kamerunerstraße 16 | 13351 Berlin<br>
          {{.WorkNumber}}
          Email: <a href="mailto:{{.Email}}" style="color: #05576d; text-decoration:none;">{{.Email}}</a><br>
          <a href="http://eoto-archiv.de" style="color: #05576d; text-decoration:none;">Web: www.eoto-archiv.de</a>
          <a href="http://cuz.berlin" style="color: #05576d; text-decoration:none;">Web: www.cuz.berlin</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #E08800; font-weight: regular; font-size: 12px; line-height: 18px; padding: 0px 0px 0px 10px;">
            <a href="https://www.facebook.com/EOTO.eV" style="color: #E08800; text-decoration:none;">Facebook</a> |
            <a href="https://www.instagram.com/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Instagram</a> |
            <a href="https://www.pinterest.de/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Pinterest</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
            Amtsgericht Charlottenburg<br>
            Registernummer: VR 31576B<br>
            Vorstand: Daniel Gyamerah, Susanna Steinbach
          </p>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px; padding: 5px 0px 0px 10px;">
          <strong>Unterstützen Sie unsere Arbeit mit Ihrer Spende.</strong><br><br>
          Spendenkonto:<br>
          Each One Teach One (EOTO) e.V. | GLS Bank<br>
          IBAN: DE24 4306 0967 1153 5692 00<br>
          BIC: GENODEM1GLS<br><br>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <img src="https://mailserver.eoto-archiv.de/IDPAD-Logo-E.png" height="50px" alt="International Decade for People of African Descent" vspace="0" hspace="10" />
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 8px; line-height: auto; padding: 0px 0px 0px 10px;">
          Each One Teach One (EOTO) e.V. supports the International Decade for People of African Descent.<br>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px"></p>
        </p>
      </td>
    </tr>
    <tr>
      <td height="30"></td>
    </tr>
  </tbody>
</table>`

	jugendTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<table border="0">
  <tbody>
    <tr>
      <td bgcolor="#FFFFFF">
        <a href="https://eoto-archiv.de" target="blank">
          <img src="https://mailserver.eoto-archiv.de/EOTO_Logo_1.png" alt="EOTO-Logo-ohne-e-V" height="50px" vspace="0" hspace="10" />
        </a>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
          <strong>{{.FullName}}</strong><br>
          {{.Pronouns}}<br>
          {{.Position}} - Jugendangebote<br>
          Each One Teach One (EOTO) e.V.<br><br>
          Kamerunerstraße 16 | 13351 Berlin<br>
          {{.WorkNumber}}
          Email: <a href="mailto:{{.Email}}" style="color: #05576d; text-decoration:none;">{{.Email}}</a><br>
          <a href="http://eoto-archiv.de" style="color: #05576d; text-decoration:none;">Web: www.eoto-archiv.de</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #E08800; font-weight: regular; font-size: 12px; line-height: 18px; padding: 0px 0px 0px 10px;">
            <a href="https://www.facebook.com/EOTO.eV" style="color: #E08800; text-decoration:none;">Facebook</a> |
            <a href="https://www.instagram.com/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Instagram</a> |
            <a href="https://www.pinterest.de/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Pinterest</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
            Amtsgericht Charlottenburg<br>
            Registernummer: VR 31576B<br>
            Vorstand: Daniel Gyamerah, Susanna Steinbach
          </p>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px; padding: 5px 0px 0px 10px;">
          <strong>Unterstützen Sie unsere Arbeit mit Ihrer Spende.</strong><br><br>
          Spendenkonto:<br>
          Each One Teach One (EOTO) e.V. | GLS Bank<br>
          IBAN: DE24 4306 0967 1153 5692 00<br>
          BIC: GENODEM1GLS<br><br>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <img src="https://mailserver.eoto-archiv.de/IDPAD-Logo-E.png" height="50px" alt="International Decade for People of African Descent" vspace="0" hspace="10" />
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 8px; line-height: auto; padding: 0px 0px 0px 10px;">
          Each One Teach One (EOTO) e.V. supports the International Decade for People of African Descent.<br>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px"></p>
        </p>
      </td>
    </tr>
    <tr>
      <td height="30"></td>
    </tr>
  </tbody>
</table>`

	narTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<table border="0">
  <tbody>
    <tr>
      <td bgcolor="#FFFFFF">
        <a href="https://eoto-archiv.de" target="blank">
          <img src="https://mailserver.eoto-archiv.de/EOTO_Logo_1.png" alt="EOTO-Logo-ohne-e-V" height="50px" vspace="0" hspace="10" />
        </a>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
          <strong>{{.FullName}}</strong><br>
          {{.Pronouns}}<br>
          {{.Position}} - Netzwerk-Antirassismus (NAR)<br>
          Each One Teach One (EOTO) e.V.<br><br>
          Kamerunerstraße 16 | 13351 Berlin<br>
          {{.WorkNumber}}
          Email: <a href="mailto:{{.Email}}" style="color: #05576d; text-decoration:none;">{{.Email}}</a><br>
          <a href="http://eoto-archiv.de" style="color: #05576d; text-decoration:none;">Web: www.eoto-archiv.de</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #E08800; font-weight: regular; font-size: 12px; line-height: 18px; padding: 0px 0px 0px 10px;">
            <a href="https://www.facebook.com/EOTO.eV" style="color: #E08800; text-decoration:none;">Facebook</a> |
            <a href="https://www.instagram.com/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Instagram</a> |
            <a href="https://www.pinterest.de/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Pinterest</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
            Amtsgericht Charlottenburg<br>
            Registernummer: VR 31576B<br>
            Vorstand: Daniel Gyamerah, Susanna Steinbach
          </p>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px; padding: 5px 0px 0px 10px;">
          <strong>Unterstützen Sie unsere Arbeit mit Ihrer Spende.</strong><br><br>
          Spendenkonto:<br>
          Each One Teach One (EOTO) e.V. | GLS Bank<br>
          IBAN: DE24 4306 0967 1153 5692 00<br>
          BIC: GENODEM1GLS<br><br>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <img src="https://mailserver.eoto-archiv.de/IDPAD-Logo-E.png" height="50px" alt="International Decade for People of African Descent" vspace="0" hspace="10" />
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 8px; line-height: auto; padding: 0px 0px 0px 10px;">
          Each One Teach One (EOTO) e.V. supports the International Decade for People of African Descent.<br>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px"></p>
        </p>
      </td>
    </tr>
    <tr>
      <td height="30"></td>
    </tr>
  </tbody>
</table>`

	afrolutionTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<table border="0">
  <tbody>
    <tr>
      <td bgcolor="#FFFFFF">
        <a href="https://eoto-archiv.de" target="blank">
          <img src="https://mailserver.eoto-archiv.de/EOTO_Logo_1.png" alt="EOTO-Logo-ohne-e-V" height="50px" vspace="0" hspace="10" />
        </a>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
          <strong>{{.FullName}}</strong><br>
          {{.Pronouns}}<br>
          {{.Position}} - Afrolution<br>
          Each One Teach One (EOTO) e.V.<br><br>
          Kamerunerstraße 16 | 13351 Berlin<br>
          {{.WorkNumber}}
          Email: <a href="mailto:{{.Email}}" style="color: #05576d; text-decoration:none;">{{.Email}}</a><br>
          <a href="http://eoto-archiv.de" style="color: #05576d; text-decoration:none;">Web: www.eoto-archiv.de</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #E08800; font-weight: regular; font-size: 12px; line-height: 18px; padding: 0px 0px 0px 10px;">
            <a href="https://www.facebook.com/EOTO.eV" style="color: #E08800; text-decoration:none;">Facebook</a> |
            <a href="https://www.instagram.com/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Instagram</a> |
            <a href="https://www.pinterest.de/eachoneteachone_official/" style="color: #E08800; text-decoration:none;">Pinterest</a>
          <p style="font-family:Open Sans, Helvetica, Arial; color:#05576d; font-size: 12px; font-weight: regular; line-height: 18px; padding: 0px 0px 0px 10px;">
            Amtsgericht Charlottenburg<br>
            Registernummer: VR 31576B<br>
            Vorstand: Daniel Gyamerah, Susanna Steinbach
          </p>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px; padding: 5px 0px 0px 10px;">
          <strong>Unterstützen Sie unsere Arbeit mit Ihrer Spende.</strong><br><br>
          Spendenkonto:<br>
          Each One Teach One (EOTO) e.V. | GLS Bank<br>
          IBAN: DE24 4306 0967 1153 5692 00<br>
          BIC: GENODEM1GLS<br><br>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <img src="https://mailserver.eoto-archiv.de/IDPAD-Logo-E.png" height="50px" alt="International Decade for People of African Descent" vspace="0" hspace="10" />
        <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 8px; line-height: auto; padding: 0px 0px 0px 10px;">
          Each One Teach One (EOTO) e.V. supports the International Decade for People of African Descent.<br>
          <p style="font-family:Open Sans, Helvetica, Arial; color: #05576d; font-size: 12px; line-height: 17px"></p>
        </p>
      </td>
    </tr>
    <tr>
      <td height="30"></td>
    </tr>
  </tbody>
</table>`
)

// GenerateSignature creates an HTML email signature based on the provided data and project
func GenerateSignature(data SignatureData) (string, error) {
	// Format pronouns like Python app does
	formattedPronouns := formatPronouns(data.Pronouns)

	// Select template based on project
	var templateStr string
	switch data.Project {
	case "each-one":
		templateStr = eachOneTemplate
	case "community":
		templateStr = communityTemplate
	case "cuz":
		templateStr = cuzTemplate
	case "jugend":
		templateStr = jugendTemplate
	case "nar":
		templateStr = narTemplate
	case "afrolution":
		templateStr = afrolutionTemplate
	default:
		templateStr = eachOneTemplate // Default to Each One
	}

	// Prepare data for template
	templateData := struct {
		FullName   string
		Position   string
		Pronouns   string
		Email      string
		WorkNumber string
	}{
		FullName:   data.FullName,
		Position:   data.Position,
		Pronouns:   formattedPronouns,
		Email:      data.Email,
		WorkNumber: data.WorkNumber,
	}

	// Parse and execute template
	tmpl, err := template.New("signature").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", err
	}

	signature := buf.String()

	// Remove work_number line if empty (like Python app does)
	if data.WorkNumber == "" {
		lines := strings.Split(signature, "\n")
		var filteredLines []string
		for _, line := range lines {
			if !strings.Contains(line, "{{.WorkNumber}}") && strings.TrimSpace(line) != "" {
				filteredLines = append(filteredLines, line)
			}
		}
		signature = strings.Join(filteredLines, "\n")
	} else {
		// Add line break after work number if it exists
		signature = strings.ReplaceAll(signature, data.WorkNumber, data.WorkNumber+"<br>")
	}

	return signature, nil
}

// formatPronouns formats pronouns like the Python app:
// "er/ihm / he/him" -> "Pronomen er/ihm - pronouns he/him"
// "Keine Pronomen / No Pronouns" -> "Keine Pronomen / No Pronouns"
func formatPronouns(pronouns string) string {
	if pronouns == "" {
		return ""
	}

	// Check if it contains the split pattern
	if strings.Contains(pronouns, " / ") {
		parts := strings.Split(pronouns, " / ")
		if len(parts) == 2 {
			germanPart := strings.TrimSpace(parts[0])
			englishPart := strings.TrimSpace(parts[1])

			// Check if it's the "No Pronouns" case
			if strings.ToLower(germanPart) == "keine pronomen" || strings.ToLower(englishPart) == "no pronouns" {
				return pronouns
			}

			// Format as: Pronomen {german} - pronouns {english}
			return fmt.Sprintf("Pronomen %s - pronouns %s", germanPart, englishPart)
		}
	}

	return pronouns
}
