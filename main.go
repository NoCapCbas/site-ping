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

  scanner := bufio.NewScanner(os.Stdin)
  fmt.Printf("Please enter a valid email> ")

  if !scanner.Scan() {
    log.Fatal("Error: Could not read from input")
  }
  
  
  email := scanner.Text()
  if !validateEmail(email) {
    log.Fatal("Error: Email must follow this format (test@example.com)")
  }

  domain := strings.Split(email, "@")[1]
  // Verify domain
  hasMX, mxRecords := verifyDomain(domain)
  if hasMX {
    isValidEmail := verifyEmail(email, mxRecords)
    log.Printf("isValid: %v", isValidEmail)
  } 

  if err := scanner.Err(); err != nil {
    log.Fatal("Error: Could not read from input: %v\n", err)
  }

}

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
  
  fmt.Printf("domain: %v, hasMX: %v, hasSPF: %v, spfRecord: %v, hasDMARC: %v, dmarcRecord: %v\n", domain, hasMX, hasSPF, spfRecord, hasDMARC, dmarcRecord)
  return hasMX, mxRecords
}

func validateEmail(email string) bool {
  // Regex to validate email format
  re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
  return re.MatchString(email)
}

func verifyEmail(email string, mxRecords []*net.MX) bool {
  // Attempt to establish smtp connection and verify email address
  for _, mx := range mxRecords {
    client, err := smtp.Dial(mx.Host + ":25")
    if err != nil {
      log.Printf("Error: %v\n", err)
    }
    defer client.Close()

    if err = client.Hello("example.com"); err != nil {
      log.Printf("Error: %v\n", err)
      continue
    }

    if err = client.Mail("verify@example.com"); err != nil {
      log.Printf("Error: %v\n", err)
      continue
    }

    if err = client.Rcpt(email); err == nil {
      return true
    }
    log.Printf("Error: %v\n", err)
  }
  return false
}


















