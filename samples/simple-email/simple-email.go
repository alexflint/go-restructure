package main

import (
	"fmt"

	"github.com/alexflint/go-restructure"
)

type EmailAddress struct {
	_    struct{} `^`
	User string   `\w+`
	_    struct{} `@`
	Host string   `[^@]+`
	_    struct{} `$`
}

func main() {
	var addr EmailAddress
	success, err := restructure.Find(&addr, "joe@example.com")
	if err != nil {
		fmt.Println(err)
	}
	if success {
		fmt.Println(addr.User) // prints "joe"
		fmt.Println(addr.Host) // prints "example.com"
	}
}
