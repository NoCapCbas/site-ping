# Email Verification Tool

## Overview
This Go program is designed to validate and verify email addresses. It checks the format of the email, verifies the domain's MX, SPF, and DMARC records, and attempts to verify the email address via an SMTP connection.

## Features
1. **Email Format Validation**: Ensures the email follows a standard format using regular expressions.
2. **Domain Verification**:
   - **MX Records**: Checks if the domain has MX records.
   - **SPF Records**: Checks if the domain has SPF records.
   - **DMARC Records**: Checks if the domain has DMARC records.
3. **Email Verification**: Attempts to verify the email by connecting to the domain's SMTP server and performing a recipient verification.

## Prerequisites
- Go installed on your machine.
- Internet access to perform DNS lookups and connect to SMTP servers.

## Installation
1. Clone the repository or download the source code.
2. Navigate to the directory containing the source code.

## Usage
1. Open a terminal and navigate to the directory containing the source code.
2. Run the program using the following command:
```bash
go run main.go
```
3. When prompted, enter the email address you wish to verify.

## Example
```sh
$ go run main.go
Please enter a valid email> test@example.com
domain: example.com, hasMX: true, hasSPF: true, spfRecord: v=spf1 include:_spf.example.com ~all, hasDMARC: true, dmarcRecord: v=DMARC1; p=none; rua=mailto:dmarc-reports@example.com
isValid: true
```

## Functions
# main

    Description: Main entry point of the program. Reads the user's input, validates the email format, verifies the domain, and attempts to verify the email address.

# verifyDomain

    Description: Verifies the domain's MX, SPF, and DMARC records.
    Parameters: domain (string) - The domain part of the email address.
    Returns: bool - Indicates if the domain has MX records, []*net.MX - The list of MX records.

# validateEmail

    Description: Validates the email format using a regular expression.
    Parameters: email (string) - The email address to validate.
    Returns: bool - Indicates if the email format is valid.

# verifyEmail

    Description: Attempts to verify the email address by connecting to the domain's SMTP server.
    Parameters: email (string) - The email address to verify, mxRecords ([]*net.MX) - The list of MX records for the domain.
    Returns: bool - Indicates if the email address is valid.

inspo: https://verifalia.com
