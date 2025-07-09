package main

import (
    "fmt"
    "log"
    "net/http"
    "net/smtp"
    "os"
)

func main() {
    http.HandleFunc("/submit-contact", handleSubmit)
    http.HandleFunc("/", serveIndex) // Serve index.html at root

    port := os.Getenv("PORT")
    if port == "" {
        port = "8088"
    }

    log.Printf("Server starting on port %s...", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    http.ServeFile(w, r, "index.html")
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form data", http.StatusBadRequest)
        return
    }

    userEmail := r.FormValue("email")
    if userEmail == "" {
        http.Error(w, "Email is required", http.StatusBadRequest)
        return
    }

    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort := os.Getenv("SMTP_PORT")
    smtpUser := os.Getenv("SMTP_USER")
    smtpPass := os.Getenv("SMTP_PASS")
    toEmail := os.Getenv("TO_EMAIL")

    if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || toEmail == "" {
        log.Println("SMTP environment variables are not fully set.")
        http.Error(w, "Server configuration error: SMTP details missing.", http.StatusInternalServerError)
        return
    }

    subject := "New contact form submission from Amsoft website"
    body := fmt.Sprintf("Someone with the email address %s has contacted you from the Amsoft website.", userEmail)
    msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", toEmail, subject, body))

    auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
    addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

    err := smtp.SendMail(addr, auth, smtpUser, []string{toEmail}, msg)
    if err != nil {
        log.Printf("Error sending email: %v", err)
        http.Error(w, "Failed to send email", http.StatusInternalServerError)
        return
    }

    log.Printf("Email successfully sent to %s from %s", toEmail, userEmail)
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Email sent successfully!")
}