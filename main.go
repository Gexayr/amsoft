package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv" // Добавляем импорт для преобразования строки в число
)

func main() {
	http.HandleFunc("/submit-contact", handleSubmit)
// 	http.HandleFunc("/", serveIndex) // Serve index.html at root

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

/* func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "index.html")
} */

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
	smtpPortStr := os.Getenv("SMTP_PORT") // Изменяем имя переменной, чтобы избежать конфликта
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	toEmail := os.Getenv("TO_EMAIL")

		log.Println(smtpHost, smtpPortStr, smtpUser, smtpPass, toEmail)


	if smtpHost == "" || smtpPortStr == "" || smtpUser == "" || smtpPass == "" || toEmail == "" {
		log.Println("SMTP environment variables are not fully set.")
		http.Error(w, "Server configuration error: SMTP details missing.", http.StatusInternalServerError)
		return
	}

	smtpPort, err := strconv.Atoi(smtpPortStr) // Преобразуем порт в целое число
	if err != nil {
		log.Printf("Invalid SMTP_PORT: %v", err)
		http.Error(w, "Server configuration error: Invalid SMTP port.", http.StatusInternalServerError)
		return
	}

	subject := "New contact form submission from Amsoft website"
	body := fmt.Sprintf("Someone with the email address %s has contacted you from the Amsoft website.", userEmail)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", toEmail, subject, body))

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort) // Используем преобразованный порт

	var client *smtp.Client
	var tlsConfig *tls.Config = &tls.Config{
		ServerName: smtpHost,
		// InsecureSkipVerify: true, // Раскомментируйте ТОЛЬКО ДЛЯ ОТЛАДКИ, если подозреваете проблемы с проверкой сертификата
	}

	// Выбираем способ подключения в зависимости от порта
	if smtpPort == 465 {
		// Для порта 465 (Implicit TLS/SSL) используем tls.Dial
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			log.Printf("Error establishing implicit TLS connection to SMTP server on port 465: %v", err)
			http.Error(w, "Failed to connect securely to SMTP server", http.StatusInternalServerError)
			return
		}
		client, err = smtp.NewClient(conn, smtpHost)
		if err != nil {
			log.Printf("Error creating SMTP client from implicit TLS connection: %v", err)
			http.Error(w, "Failed to initialize SMTP client", http.StatusInternalServerError)
			return
		}
	} else {
		// Для порта 587 (STARTTLS) или других портов используем smtp.Dial и StartTLS
		c, err := smtp.Dial(addr)
		if err != nil {
			log.Printf("Error connecting to SMTP server on port %d: %v", smtpPort, err)
			http.Error(w, "Failed to connect to SMTP server", http.StatusInternalServerError)
			return
		}
		client = c

		if err = client.StartTLS(tlsConfig); err != nil {
			log.Printf("Error starting TLS on port %d: %v", smtpPort, err)
			http.Error(w, "Failed to establish secure connection with SMTP server", http.StatusInternalServerError)
			return
		}
	}

	defer client.Close() // Закрываем клиент в конце функции

	if err = client.Auth(auth); err != nil {
		log.Printf("Error authenticating to SMTP server: %v", err)
		http.Error(w, "Failed to authenticate with SMTP server", http.StatusInternalServerError)
		return
	}

	if err = client.Mail(smtpUser); err != nil {
		log.Printf("Error setting sender: %v", err)
		http.Error(w, "Failed to set sender email", http.StatusInternalServerError)
		return
	}
	if err = client.Rcpt(toEmail); err != nil {
		log.Printf("Error setting recipient: %v", err)
		http.Error(w, "Failed to set recipient email", http.StatusInternalServerError)
		return
	}

	w_mail, err := client.Data()
	if err != nil {
		log.Printf("Error getting data writer: %v", err)
		http.Error(w, "Failed to prepare email data", http.StatusInternalServerError)
		return
	}
	_, err = w_mail.Write(msg)
	if err != nil {
		log.Printf("Error writing email data: %v", err)
		http.Error(w, "Failed to write email data", http.StatusInternalServerError)
		return
	}
	err = w_mail.Close()
	if err != nil {
		log.Printf("Error closing mail writer: %v", err)
		http.Error(w, "Failed to finalize email sending", http.StatusInternalServerError)
		return
	}

	log.Printf("Email successfully sent to %s from %s", toEmail, userEmail)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Email sent successfully!")
}
