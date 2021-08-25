package main

import (
	"github.com/alexflint/go-restructure"
	"github.com/kr/pretty"
)

// PhoneNumber matches "123-456-7890"
type PhoneNumber struct {
	Area   string   `\d{3}`
	_      struct{} `-`
	Group  string   `\d{3}`
	_      struct{} `-`
	Digits string   `\d{4}`
}

// EmailAddress matches "greg@example.com"
type EmailAddress struct {
	User string   `\w+`
	_    struct{} `@`
	Host string   `.+`
}

// ContactInfo matches a PhoneNumber or an Email Address
type ContactInfo interface{}

// ContactLine matches "please contact 123-456-7890" or "please contact greg@example.com"
type ContactLine struct {
	Prelude string `please contact `
	Contact ContactInfo
}

func main() {
	var contact ContactInfo
	restructure.RegisterUnion(&contact, EmailAddress{}, PhoneNumber{})

	pattern := restructure.MustCompile(ContactLine{}, restructure.Options{})

	var line ContactLine
	pattern.Find(&line, "please contact greg@example.com")
	pretty.Println(line)

	pattern.Find(&line, "please contact 123-456-7890")
	pretty.Println(line)
}
