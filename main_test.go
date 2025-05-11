package main

import (
	"os"
	"testing"

	"github.com/emersion/go-imap"
)

func TestParseSenderInfo(t *testing.T) {
	env := &imap.Envelope{
		From: []*imap.Address{
			{
				PersonalName: "Alice Smith",
				MailboxName:  "alice",
				HostName:     "example.com",
			},
		},
		Subject: "Hello from Alice",
	}
	msg := &imap.Message{Envelope: env}
	name, email, subject := parseSenderInfo(msg)

	if name != "Alice Smith" {
		t.Errorf("Expected name 'Alice Smith', got '%s'", name)
	}
	if email != "alice@example.com" {
		t.Errorf("Expected email 'alice@example.com', got '%s'", email)
	}
	if subject != "Hello from Alice" {
		t.Errorf("Expected subject 'Hello from Alice', got '%s'", subject)
	}
}

func TestWriteToCSV(t *testing.T) {
	testData := [][]string{
		{"Alice Smith", "alice@example.com", "Hello from Alice"},
		{"Bob Jones", "bob@example.com", "Meeting Update"},
	}

	filename := "test_emails.csv"
	err := writeToCSV(testData, filename)
	if err != nil {
		t.Fatalf("writeToCSV returned error: %v", err)
	}

	// Cleanup
	defer os.Remove(filename)

	// Check file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("CSV file was not created")
	}
}
