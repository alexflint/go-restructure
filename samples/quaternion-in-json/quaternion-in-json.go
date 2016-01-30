package main

import (
	"encoding/json"
	"fmt"

	"github.com/alexflint/go-restructure"
)

var quaternionRegexp = restructure.MustCompile(QuotedQuaternion{}, restructure.Options{})

type RealPart struct {
	Sign string `regexp:"[+-]?"`
	Real string `regexp:"[0-9]+"`
}

type SignedInt struct {
	Sign string `regexp:"[+-]"`
	Real string `regexp:"[0-9]+"`
}

type IPart struct {
	Magnitude SignedInt
	_         struct{} `regexp:"i"`
}

type JPart struct {
	Magnitude SignedInt
	_         struct{} `regexp:"j"`
}

type KPart struct {
	Magnitude SignedInt
	_         struct{} `regexp:"k"`
}

// matches "1+2i+3j+4k", "-1+2k", "-1", etc
type Quaternion struct {
	Real *RealPart
	I    *IPart `regexp:"?"`
	J    *JPart `regexp:"?"`
	K    *KPart `regexp:"?"`
}

// matches the quoted strings `"-1+2i+3j+4k"`, `"3-4k"`, `"12+34i"`, etc
type QuotedQuaternion struct {
	_          struct{} `regexp:"^"`
	_          struct{} `regexp:"\""`
	Quaternion *Quaternion
	_          struct{} `regexp:"\""`
	_          struct{} `regexp:"$"`
}

func (c *QuotedQuaternion) UnmarshalJSON(b []byte) error {
	if !quaternionRegexp.Find(c, string(b)) {
		return fmt.Errorf("%s is not a quaternion number", string(b))
	}
	return nil
}

// this struct is handled by JSON
type Var struct {
	Name  string
	Value *QuotedQuaternion
}

func prettyPrint(x interface{}) string {
	buf, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(buf)
}

func main() {
	src := `{"name": "foo", "value": "1+2i+3j+4k"}`
	var v Var
	err := json.Unmarshal([]byte(src), &v)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(prettyPrint(v))
}
