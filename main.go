package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/submit-contact", handleSubmit)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatID == "" {
		log.Println("TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID is not set.")
		http.Error(w, "Server configuration error: Telegram details missing.", http.StatusInternalServerError)
		return
	}

	text := fmt.Sprintf("New contact form submission from Amsoft website\n\nEmail: %s", userEmail)

	payload, err := json.Marshal(map[string]string{
		"chat_id": chatID,
		"text":    text,
	})
	if err != nil {
		log.Printf("Error marshaling Telegram payload: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("Error sending Telegram message: %v", err)
		http.Error(w, "Failed to send Telegram message", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Telegram API returned non-200 status: %d", resp.StatusCode)
		http.Error(w, "Failed to deliver Telegram message", http.StatusInternalServerError)
		return
	}

	log.Printf("Telegram message sent for contact from %s", userEmail)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Message sent successfully!")
}
