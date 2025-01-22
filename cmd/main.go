package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"net/textproto"
	"regexp"
	"strings"
	"time"
)

// EmailResponse represents the JSON response structure.
type EmailResponse struct {
	Email       string `json:"email"`
	IsValid     bool   `json:"is_valid"`
	HasMX       bool   `json:"has_mx"`
	HasSPF      bool   `json:"has_spf"`
	SPFRecord   string `json:"spf_record"`
	HasDMARC    bool   `json:"has_dmarc"`
	DMARCRecord string `json:"dmarc_record"`
	Error       string `json:"error,omitempty"`
}

func main() {
	http.HandleFunc("/verify-email", handleEmailVerification)
	http.HandleFunc("/", handleIndex)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

// handleEmailVerification handles HTTP requests to verify an email address.
func handleEmailVerification(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		log.Printf("Email parameter is missing")
		http.Error(w, "Email parameter is missing", http.StatusBadRequest)
		return
	}

	// Validate the email format
	if !validateEmail(email) {
		log.Printf("Invalid email format: %s", email)
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Extract domain from email and verify domain records
	domain := strings.Split(email, "@")[1]
	hasMX, mxRecords, hasSPF, spfRecord, hasDMARC, dmarcRecord := verifyDomain(domain)

	// Verify the email using SMTP if the domain has MX records
	isValid := false
	var errorMsg string
	if hasMX {
		log.Printf("Verifying email: %s", email)
		isValid, errorMsg = verifyEmail(email, mxRecords)
	} else {
		log.Printf("No MX records found for the domain: %s", domain)
		errorMsg = "No MX records found for the domain"
	}

	// Create the response and encode it as JSON
	response := EmailResponse{
		Email:       email,
		IsValid:     isValid,
		HasMX:       hasMX,
		HasSPF:      hasSPF,
		SPFRecord:   spfRecord,
		HasDMARC:    hasDMARC,
		DMARCRecord: dmarcRecord,
		Error:       errorMsg,
	}
	log.Printf("Response: %+v", response)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// verifyDomain checks the MX, SPF, and DMARC records for a domain.
func verifyDomain(domain string) (bool, []*net.MX, bool, string, bool, string) {
	var hasMX, hasSPF, hasDMARC bool
	var spfRecord, dmarcRecord string

	// Check domain MX records
	mxRecords, err := net.LookupMX(domain)
	if err == nil && len(mxRecords) > 0 {
		hasMX = true
	}

	// Check domain SPF records
	txtRecords, err := net.LookupTXT(domain)
	if err == nil {
		for _, record := range txtRecords {
			if strings.HasPrefix(record, "v=spf1") {
				hasSPF = true
				spfRecord = record
				break
			}
		}
	}

	// Check domain DMARC records
	dmarcRecords, err := net.LookupTXT("_dmarc." + domain)
	if err == nil {
		for _, record := range dmarcRecords {
			if strings.HasPrefix(record, "v=DMARC1") {
				hasDMARC = true
				dmarcRecord = record
				break
			}
		}
	}

	return hasMX, mxRecords, hasSPF, spfRecord, hasDMARC, dmarcRecord
}

// validateEmail uses regex to validate the format of an email address.
func validateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// verifyEmail attempts to verify an email address by connecting to its domain SMTP server.
func verifyEmail(email string, mxRecords []*net.MX) (bool, string) {
	domain := strings.Split(email, "@")[1]

	for _, mx := range mxRecords {
		dialer := net.Dialer{
			Timeout: 5 * time.Second,
		}

		serverAddr := fmt.Sprintf("%s:25", mx.Host)
		conn, err := dialer.Dial("tcp", serverAddr)
		if err != nil {
			log.Printf("Connection error for %s: %v", mx.Host, err)
			continue
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, mx.Host)
		if err != nil {
			log.Printf("SMTP client error for %s: %v", mx.Host, err)
			continue
		}
		defer client.Close()

		err = client.Hello(domain)
		if err != nil {
			log.Printf("HELO error for %s: %v", mx.Host, err)
			continue
		}

		err = client.Mail("test@" + domain)
		if err != nil {
			log.Printf("MAIL FROM error for %s: %v", mx.Host, err)
			continue
		}

		err = client.Rcpt(email)
		if err != nil {
			log.Printf("RCPT TO error type: %T", err)
			log.Printf("RCPT TO error for %s: %v", mx.Host, err)

			switch e := err.(type) {
			case *textproto.Error:
				return false, fmt.Sprintf("SMTP Error Code %d: %s", e.Code, e.Msg)
			default:
				return false, fmt.Sprintf("Delivery failed: %v", err)
			}
		}

		return true, ""
	}

	return false, "Could not verify email: all MX servers failed"
}
