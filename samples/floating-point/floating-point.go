package main

import (
	"encoding/json"
	"fmt"

	"github.com/alexflint/go-restructure"
)

var floatRegexp = restructure.MustCompile(Float{}, restructure.Options{})

// Matches "123", "1.23", "1.23e-4", "-12.3E+5", ".123"
type Float struct {
	Sign     *Sign     `?`
	Whole    string    `[0-9]*`
	Period   struct{}  `\.?`
	Frac     string    `[0-9]+`
	Exponent *Exponent `?`
}

// Matches "+" or "-"
type Sign struct {
	Ch string `[+-]`
}

// Matches "e+4", "E6", "e-03"
type Exponent struct {
	_    struct{} `[eE]`
	Sign *Sign    `?`
	Num  string   `[0-9]+`
}

func prettyPrint(x interface{}) string {
	buf, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(buf)
}

func main() {
	var f Float
	for _, str := range []string{"1.23", "1.23e+45", ".123", "12e3"} {
		floatRegexp.Find(&f, str)
		fmt.Printf("\"%s\" -> %s\n\n", str, prettyPrint(f))
	}
}
