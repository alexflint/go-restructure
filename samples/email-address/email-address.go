package main

import (
	"fmt"

	"github.com/alexflint/go-restructure"
)

type Hostname struct {
	Domain string   `\w+`
	_      struct{} `\.`
	TLD    string   `\w+`
}

type EmailAddress struct {
	_    struct{} `^`
	User string   `[a-zA-Z0-9._%+-]+`
	_    struct{} `@`
	Host *Hostname
	_    struct{} `$`
}

func main() {
	var addr EmailAddress
	success, _ := restructure.Find(&addr, "joe@example.com")
	if success {
		fmt.Println(addr.User)        // prints "joe"
		fmt.Println(addr.Host.Domain) // prints "example"
		fmt.Println(addr.Host.TLD)    // prints "com"
	}
}
