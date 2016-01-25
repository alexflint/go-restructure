[![GoDoc](https://godoc.org/github.com/alexflint/go-restructure?status.svg)](https://godoc.org/github.com/alexflint/go-restructure)
[![Build Status](https://travis-ci.org/alexflint/go-restructure.svg?branch=master)](https://travis-ci.org/alexflint/go-restructure)

## Match regular expressions into struct fields

```shell
go get github.com/alexflint/go-restructure
```

This package allows you to express regular expressions by defining a struct, and then capture matched sub-expressions into struct fields. Here is a very simple email address parser:

```go
import "github.com/alexflint/go-restructure"

type EmailAddress struct {
	_    string `^`
	User string `\w+`
	_    string `@`
	Host string `[^@]+`
	_    string `$`
}

func main() {
	var addr EmailAddress
	restructure.Find(&addr, "joe@example.com")
	fmt.Println(addr.User) // prints "joe"
	fmt.Println(addr.Host) // prints "example.com"
}
```
(Note that the above is far too simplistic to be used as a serious email address validator.)

The regular expression that was executed was the concatenation of the struct tags:

```
^(\w+)@([^@]+)$
```

The first submatch was inserted into the `User` field and the second into the `Host` field.

Here is a slightly more sophisticated email address parser that uses nested structs:

```go
import "github.com/alexflint/go-restructure"

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
```

Compare this to using the standard library `FindStringSubmatchIndex` directly:

```go
func main() {
	content := "joe@example.com"
	expr := regexp.MustCompile(`^([a-zA-Z0-9._%+-]+)@((\w+)\.(\w+))$`)
	indices := expr.FindStringSubmatchIndex(content)
	if len(indices) > 0 {
		userBegin, userEnd := indices[2], indices[3]
		var user string
		if userBegin != -1 && userEnd != -1 {
			user = content[userBegin:userEnd]
		}

		domainBegin, domainEnd := indices[6], indices[7]
		var domain string
		if domainBegin != -1 && domainEnd != -1 {
			domain = content[domainBegin:domainEnd]
		}

		tldBegin, tldEnd := indices[8], indices[9]
		var tld string
		if tldBegin != -1 && tldEnd != -1 {
			tld = content[tldBegin:tldEnd]
		}

		fmt.Println(user)   // prints "joe"
		fmt.Println(domain) // prints "example"
		fmt.Println(tld)    // prints "com"
	}
}
```

Also compare this to using the standard library `FindStringSubmatch`:

```go
func main() {
	re := regexp.MustCompile(`^([a-zA-Z0-9._%+-]+)@(\w+)\.(\w+)$`)
	m := re.FindStringSubmatch("joe@example.com")
	if len(m) > 0 {
		user, domain, tld := m[1], m[2], m[3]
		fmt.Println(user)   // prints "joe"
		fmt.Println(domain) // prints "example"
		fmt.Println(tld)    // prints "com"
	}
}
```
