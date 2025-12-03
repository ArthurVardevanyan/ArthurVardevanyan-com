package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/wneessen/go-mail"
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
	safe := html.EscapeString(input)

	// Convert newlines to \r\n for SMTP compliance
	return strings.ReplaceAll(safe, "\n", "\r\n")
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
	// Remove any control characters, line/paragraph separators, etc.
	// This protects against email content and header injection.
	// Unicode controls: [0x00-0x1F], [0x7F], [U+2028], [U+2029]
	re := regexp.MustCompile(`[\x00-\x1F\x7F\u2028\u2029]`)
	input = re.ReplaceAllString(input, " ")
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

	// Apply robust sanitization to all fields before use in body
	safeName := singleLine(sanitizeName(req.Name))
	safeMessage := singleLine(sanitizeBody(req.Message))

	if len(safeName) == 0 || len(safeName) > 100 {
		return fmt.Errorf("Invalid name format")
	}

	// Create a new message
	m := mail.NewMsg()
	if err := m.From(from); err != nil {
		return fmt.Errorf("failed to set From address: %w", err)
	}
	if err := m.To(from); err != nil {
		return fmt.Errorf("failed to set To address: %w", err)
	}

	// The library validates this email address automatically
	// We set Reply-To so you can reply to the user
	if err := m.ReplyTo(req.Email); err != nil {
		log.Printf("Invalid Reply-To email provided: %v", err)
		// We continue without Reply-To if invalid
	}

	m.Subject("Contact Form Submission")

	// Construct body
	body := fmt.Sprintf("Submitted Name: %s\nSubmitted Email: %s\nMessage:\n%s",
		safeName, req.Email, safeMessage)

	m.SetBodyString(mail.TypeTextPlain, body)

	// Setup the client
	// It automatically handles port splitting and authentication
	c, err := mail.NewClient(smtpHost,
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(from),
		mail.WithPassword(password),
		// mail.WithTLSPolicy(mail.TLSMandatory), // Recommended for security
	)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	if err := c.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
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
