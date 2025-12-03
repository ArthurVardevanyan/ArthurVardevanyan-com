package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
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
	input = strings.ReplaceAll(input, "\r", " ")
	input = strings.ReplaceAll(input, "\n", " ")
	re := regexp.MustCompile(`[^\x20-\x7E]`)
	input = re.ReplaceAllString(input, "")
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
	to := os.Getenv("SMTP_TO")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	if from == "" || password == "" || to == "" || smtpHost == "" || smtpPort == "" {
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

	// Construct the email message with proper headers
	headers := make(map[string]string)
	// IMPORTANT: Never use user-supplied data in headers. Only use config/env values here.
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = "Contact Form Submission"
	headers["Content-Type"] = "text/plain; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n"
	// Do not include raw user-supplied email address in body to prevent email content injection.
	message += fmt.Sprintf("Submitted Name: %s\nMessage:\n%s",
		safeName,
		safeMessage)

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Message) > 10000 {
		http.Error(w, "Message too long", http.StatusBadRequest)
		return
	}

	if err := verifyRecaptcha(req.RecaptchaResponse); err != nil {
		log.Printf("Recaptcha verification failed: %v", err)
		http.Error(w, "Recaptcha verification failed", http.StatusUnauthorized)
		return
	}

	if err := sendEmail(req); err != nil {
		log.Printf("Failed to send email: %v", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
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
