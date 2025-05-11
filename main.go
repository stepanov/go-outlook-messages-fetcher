package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	imapServer := os.Getenv("OUTLOOK_IMAP_SERVER")
	if imapServer == "" {
		log.Fatal("OUTLOOK_IMAP_SERVER not set in .env file")
	}

	// Parse command-line flags
	username := flag.String("username", "", "Outlook email address")
	password := flag.String("password", "", "Outlook password or app password")
	flag.Parse()

	if *username == "" || *password == "" {
		log.Fatal("Both -username and -password flags are required")
	}

	// Connect to the Outlook IMAP server
	log.Println("Connecting to server...")
	c, err := client.DialTLS(imapServer, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(*username, *password); err != nil {
		log.Fatal(err)
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

	// Set range to fetch
	seqset := new(imap.SeqSet)
	seqset.AddRange(1, mbox.Messages)

	// Fetch envelopes
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	// Create CSV file
	file, err := os.Create("emails.csv")
	if err != nil {
		log.Fatal("Could not create CSV file:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Name", "Email", "Subject"})

	// Process messages
	for msg := range messages {
		if msg.Envelope != nil && len(msg.Envelope.From) > 0 {
			from := msg.Envelope.From[0]
			name := from.PersonalName
			email := from.MailboxName + "@" + from.HostName
			subject := msg.Envelope.Subject

			fmt.Printf("From: %s <%s>\nSubject: %s\n\n", name, email, subject)

			writer.Write([]string{name, email, subject})
		}
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done. Output written to emails.csv")
}
