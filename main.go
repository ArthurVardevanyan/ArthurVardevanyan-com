package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type EmailRequest struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Message           string `json:"message"`
	RecaptchaResponse string `json:"g-recaptcha-response"`
}

type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

func verifyRecaptcha(responseToken string) error {
	secret := os.Getenv("RECAPTCHA_SECRET_KEY")
	if secret == "" {
		return fmt.Errorf("RECAPTCHA_SECRET_KEY not set")
	}

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify",
		map[string][]string{
			"secret":   {secret},
			"response": {responseToken},
		})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result RecaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("recaptcha verification failed: %v", result.ErrorCodes)
	}

	return nil
}

func sanitizeHeader(input string) string {
	// Only allow ASCII letters, digits, '@', '.', '+', and '-'. Remove everything else.
	re := regexp.MustCompile(`[^a-zA-Z0-9@.+\-]`)
	safe := re.ReplaceAllString(input, "")
	// Remove any remaining CR/LF, just in case
	safe = strings.ReplaceAll(safe, "\r", "")
	safe = strings.ReplaceAll(safe, "\n", "")
	return safe
}

func sanitizeBody(input string) string {
	// Normalize newlines to \n
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")

	// Remove URLs in the message to prevent phishing/content injection.
	urlRe := regexp.MustCompile(`https?://[^\s]+`)
	input = urlRe.ReplaceAllString(input, "[URL Removed]")

	// Remove all non-printable ASCII chars except newline and tab.
	re := regexp.MustCompile(`[^\x20-\x7E\n\t]`)
	input = re.ReplaceAllString(input, "")

	// Escape potential HTML, though body is text.
	return html.EscapeString(input)
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func sanitizeName(input string) string {
	// Allow only alphabets, literal spaces, hyphens, apostrophes.
	// This strictly excludes newlines, carriage returns, tabs, and other control characters.
	re := regexp.MustCompile(`[^a-zA-Z \-\']`)
	input = re.ReplaceAllString(input, "")

	// Collapse multiple spaces and trim
	spaceRe := regexp.MustCompile(` +`)
	return strings.TrimSpace(spaceRe.ReplaceAllString(input, " "))
}

// singleLine removes any newlines or carriage returns and collapses to a single line.
func singleLine(input string) string {
	// Remove newlines, carriage returns, tabs
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")
	input = strings.ReplaceAll(input, "\t", " ")
	// Collapse multiple spaces, trim
	spaceRe := regexp.MustCompile(` +`)
	return strings.TrimSpace(spaceRe.ReplaceAllString(input, " "))
}

func sendEmail(req EmailRequest) error {
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	if from == "" || password == "" || smtpHost == "" {
		return fmt.Errorf("SMTP configuration missing")
	}

	safeName := singleLine(sanitizeName(req.Name))
	safeEmail := sanitizeHeader(req.Email)
	safeMessage := sanitizeBody(req.Message)

	if len(safeName) == 0 || len(safeName) > 100 {
		return fmt.Errorf("Invalid name format")
	}

	if !isValidEmail(safeEmail) || len(safeEmail) > 254 {
		return fmt.Errorf("Invalid email address")
	}

	// Send the email to the configured sender (the site owner)
	to := from

	// Construct the email message with proper headers
	headers := make(map[string]string)
	// IMPORTANT: Never use user-supplied data in headers. Only use config/env values here.
	headers["From"] = from
	headers["To"] = to
	headers["Reply-To"] = safeEmail
	headers["Subject"] = "Contact Form Submission"
	headers["Content-Type"] = "text/plain; charset=UTF-8"

	var msgBuffer bytes.Buffer
	for k, v := range headers {
		msgBuffer.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msgBuffer.WriteString("\r\n")

	// Use text/template to safely construct the body
	t := template.Must(template.New("emailBody").Parse("Submitted Name: {{.Name}}\nSubmitted Email: {{.Email}}\nMessage:\n{{.Message}}"))
	if err := t.Execute(&msgBuffer, map[string]string{
		"Name":    safeName,
		"Email":   safeEmail,
		"Message": safeMessage,
	}); err != nil {
		return err
	}

	// Extract host from smtpHost (which is expected to be host:port) for authentication
	host, _, err := net.SplitHostPort(smtpHost)
	if err != nil {
		return fmt.Errorf("invalid SMTP_HOST format: %v", err)
	}

	auth := smtp.PlainAuth("", from, password, host)
	err = smtp.SendMail(smtpHost, auth, from, []string{to}, msgBuffer.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func emailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EmailRequest
	isForm := false

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	} else {
		// Assume form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}
		req.Name = r.FormValue("name")
		req.Email = r.FormValue("email")
		req.Message = r.FormValue("message")
		req.RecaptchaResponse = r.FormValue("recaptcha_response")
		isForm = true
	}

	if len(req.Message) > 10000 {
		if isForm {
			http.Redirect(w, r, "/#emailFailed", http.StatusSeeOther)
			return
		}
		http.Error(w, "Message too long", http.StatusBadRequest)
		return
	}

	if err := verifyRecaptcha(req.RecaptchaResponse); err != nil {
		log.Printf("Recaptcha verification failed: %v", err)
		if isForm {
			http.Redirect(w, r, "/#emailFailed", http.StatusSeeOther)
			return
		}
		http.Error(w, "Recaptcha verification failed", http.StatusUnauthorized)
		return
	}

	if err := sendEmail(req); err != nil {
		log.Printf("Failed to send email: %v", err)
		if isForm {
			http.Redirect(w, r, "/#emailFailed", http.StatusSeeOther)
			return
		}
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	if isForm {
		http.Redirect(w, r, "/#emailSent", http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully"))
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func startupzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	staticPath := os.Getenv("KO_DATA_PATH")
	if staticPath == "" {
		staticPath = "./kodata"
	}

	fs := http.FileServer(http.Dir(staticPath))
	http.Handle("/", fs)

	http.HandleFunc("/email", emailHandler)
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/startupz", startupzHandler)

	log.Printf("Server starting on port %s serving %s", port, staticPath)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
