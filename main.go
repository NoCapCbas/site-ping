package main

import (
  "bufio"
  "fmt"
  "log"
  "strings"
  "os"
  "net"
  "net/smtp"
  "regexp"
)

func main() {
  // Scanner to read input from the standard input console
  scanner := bufio.NewScanner(os.Stdin)
  fmt.Printf("Please enter a valid email> ")
  
  // Check if user input exists
  if !scanner.Scan() {
    log.Fatal("Error: Could not read from input")
  }
  
  // Get user input 
  email := scanner.Text()
  // Validate the email format
  if !validateEmail(email) {
    log.Fatal("Error: Email must follow this format (test@example.com)")
  }
  // Extract domain from email
  domain := strings.Split(email, "@")[1]
  // Verify domain (MX, SPF, DMARC)
  hasMX, mxRecords := verifyDomain(domain)
  if hasMX {
    // Verify the email using SMTP
    isValidEmail := verifyEmail(email, mxRecords)
    log.Printf("isValid: %v", isValidEmail)
  } 
  
  // Check for any errors  in the input scanner
  if err := scanner.Err(); err != nil {
    log.Fatal("Error: Could not read from input: %v\n", err)
  }

}

// VerifyDomain checks the MX, SPF, and DMARC records for a domain
func verifyDomain(domain string) (bool, []*net.MX) {

  var hasMX, hasSPF, hasDMARC bool
  var spfRecord, dmarcRecord string
   
  // Check domain MX records
  mxRecords, err := net.LookupMX(domain)
  if err != nil {
    log.Printf("Error: %v\n", err)
  }
  if len(mxRecords) > 0 {
    hasMX = true
  }    
  
  // Check domain spf records
  txtRecords, err := net.LookupTXT(domain)
  if err != nil {
    log.Printf("Error: %v\n", err)
  }

  for _, record := range txtRecords{
    if strings.HasPrefix(record, "v=spf1"){
      hasSPF = true
      spfRecord = record
      break
    }
  }
  
  // Check domain dmarc records
  dmarcRecords, err := net.LookupTXT("_dmarc." + domain)
  if err != nil {
    log.Printf("Error: %v\n", err)
  }

  for _, record := range dmarcRecords {
    if strings.HasPrefix(record, "v=DMARC1"){
      hasDMARC = true
      dmarcRecord = record
      break
    }
  }

  // Print the domain verification results
  fmt.Printf("domain: %v, hasMX: %v, hasSPF: %v, spfRecord: %v, hasDMARC: %v, dmarcRecord: %v\n", domain, hasMX, hasSPF, spfRecord, hasDMARC, dmarcRecord)
  return hasMX, mxRecords
}

// validateEmail uses regex to validate the format of an email address
func validateEmail(email string) bool {
  // Regex to validate email format
  re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
  return re.MatchString(email)
}

// verifyEmail attemtps to verify an email address by connecting to its domain SMTP server
func verifyEmail(email string, mxRecords []*net.MX) bool {
  // Attempt to establish smtp connection and verify email address
  for _, mx := range mxRecords {
    client, err := smtp.Dial(mx.Host + ":25")
    if err != nil {
      log.Printf("Error: %v\n", err)
    }
    defer client.Close()
    
    // Introduction to smtp server
    if err = client.Hello("example.com"); err != nil {
      log.Printf("Error: %v\n", err)
      continue
    }
    
    // Specify the sender email
    if err = client.Mail("verify@example.com"); err != nil {
      log.Printf("Error: %v\n", err)
      continue
    }
    
    // Specify the recipient email
    if err = client.Rcpt(email); err == nil {
      return true
    }
    log.Printf("Error: %v\n", err)
  }
  // Email address is invalid
  return false
}
