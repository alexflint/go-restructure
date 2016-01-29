package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/alexflint/go-arg"
	"github.com/alexflint/go-restructure"
)

type DotName struct {
	Dot  string `\.`
	Name string `\w+`
}

type DotExpr struct {
	_    struct{} `^`
	Head string   `foo`
	Tail *DotName `?`
	_    struct{} `$`
}

func prettyPrint(x interface{}) string {
	buf, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(buf)
}

func main() {
	var args struct {
		Str string `arg:"positional"`
	}
	arg.MustParse(&args)

	// Construct the regular expression
	pattern, err := restructure.Compile(&DotExpr{}, restructure.Options{})
	if err != nil {
		log.Fatal(err)
	}

	// Match
	var v DotExpr
	fmt.Println(pattern.Find(&v, args.Str))
}
