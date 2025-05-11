package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	imapServer := os.Getenv("OUTLOOK_IMAP_SERVER")
	if imapServer == "" {
		log.Fatal("OUTLOOK_IMAP_SERVER not set in .env file")
	}

	// Parse flags
	username := flag.String("username", "", "Outlook email address")
	password := flag.String("password", "", "Outlook password or app password")
	flag.Parse()

	if *username == "" || *password == "" {
		log.Fatal("Both -username and -password flags are required")
	}

	// Connect to IMAP server
	log.Println("Connecting to server...")
	c, err := client.DialTLS(imapServer, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(*username, *password); err != nil {
		if err.Error() == "LOGIN failed" || err.Error() == "authentication failed" {
			log.Fatal("Login failed: invalid username or password (check if app password is needed)")
		}
		log.Fatalf("Login error: %v", err)
	}
	log.Println("Logged in")

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	if mbox.Messages == 0 {
		log.Println("No messages in INBOX")
		return
	}

	// Fetch envelopes
	seqset := new(imap.SeqSet)
	seqset.AddRange(1, mbox.Messages)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	var records [][]string

	for msg := range messages {
		name, email, subject := parseSenderInfo(msg)
		fmt.Printf("From: %s <%s>\nSubject: %s\n\n", name, email, subject)

		if name != "" || email != "" || subject != "" {
			records = append(records, []string{name, email, subject})
		}
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	if err := writeToCSV(records, "emails.csv"); err != nil {
		log.Fatal("Failed to write CSV:", err)
	}

	log.Println("Done. Output written to emails.csv")
}

// parseSenderInfo extracts name, email, and subject from an IMAP message.
func parseSenderInfo(msg *imap.Message) (string, string, string) {
	if msg.Envelope != nil && len(msg.Envelope.From) > 0 {
		from := msg.Envelope.From[0]
		name := from.PersonalName
		email := from.MailboxName + "@" + from.HostName
		subject := msg.Envelope.Subject
		return name, email, subject
	}
	return "", "", ""
}

// writeToCSV writes email metadata to a CSV file.
func writeToCSV(records [][]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write([]string{"Name", "Email", "Subject"}); err != nil {
		return err
	}

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
