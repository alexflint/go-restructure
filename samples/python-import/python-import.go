package main

import (
	"fmt"

	"github.com/alexflint/go-restructure"
)

var importRegexp = restructure.MustCompile(Import{}, restructure.Options{})

// Import matches "import foo" and "import foo as bar"
type Import struct {
	_       struct{}             `^import\s+`
	Package restructure.Submatch `\w+`
	Alias   *AsName              `?`
	_       struct{}             `$`
}

// AsName matches "as xyz"
type AsName struct {
	_    struct{}             `\s+as\s+`
	Name restructure.Submatch `\w+`
}

func main() {
	var imp Import
	importRegexp.Find(&imp, "import foo as bar")
	fmt.Printf("IMPORT %s (bytes %d...%d)\n", imp.Package.String(), imp.Package.Begin, imp.Package.End)
	fmt.Printf("    AS %s (bytes %d...%d)\n", imp.Alias.Name.String(), imp.Alias.Name.Begin, imp.Alias.Name.End)
}
