package restructure

import (
	"io/ioutil"
	"os"
	"regexp"
	"testing"
)

var src = `
The US economy went through an economic downturn following the financial 
crisis of 2007–08, with output as late as 2013 still below potential
according to the Congressional Budget Office.[57] The economy, however,
began to recover in the second half of 2009, and as of November 2015,
unemployment had declined from a high of 10% to 5%; the government's
broader U-6 unemployment rate, which includes the part-time underemployed,
was 9.8% (it had reached 16% in 2009).[13] At 11.3%, the U.S. has one of
the lowest labor union participation rates in the OECD.[58] Households
living on less than $2 per day before government benefits, doubled from
1996 levels to 1.5 million households in 2011, including 2.8 million
children.[59] The gap in income between rich and poor is greater in the
United States than in any other developed country.[60] Total public and
private debt was $50 trillion at the end of the first quarter of 2010,
or 3.5 times GDP.[61] In December 2014, public debt was slightly more
than 100% of GDP.[62] Domestic financial assets totaled $131 trillion
and domestic financial liabilities totaled $106 trillion.[63]
`

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

func BenchmarkParseFloat(b *testing.B) {
	pattern := MustCompile(Float{}, Options{})
	var f Float
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern.Find(&f, src)
	}
}

func BenchmarkParseFloatStdlib(b *testing.B) {
	pattern := regexp.MustCompile(`((?P<Sign>((?P<Ch>[\+\-]))?)(?P<Whole>[0-9]*)(?P<Period>\.?)(?P<Frac>[0-9]+)(?P<Exponent>((?i:E)(?P<Sign>((?P<Ch>[\+\-]))?)(?P<Num>[0-9]+))?))`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern.FindSubmatch([]byte(src))
	}
}

type EmailAddress struct {
	_    struct{} `^`
	User string   `[a-zA-Z0-9._%+-]+`
	_    struct{} `@`
	Host string   `.+`
	_    struct{} `$`
}

func BenchmarkParseEmail(b *testing.B) {
	var addr EmailAddress
	pattern := MustCompile(EmailAddress{}, Options{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern.Find(&addr, "joe@example.com")
	}
}

func BenchmarkParseEmailStdlib(b *testing.B) {
	//pattern := regexp.MustCompile(`(\A(?P<User>[%\+\--\.0-9A-Z_a-z]+)@(?P<Host>((?P<Domain>[0-9A-Z_a-z]+)\.(?P<TLD>[0-9A-Z_a-z]+)))(?-m:$))`)
	pattern := regexp.MustCompile(`(\A(?P<User>[%\+\--\.0-9A-Z_a-z]+)@(?P<Host>.+)(?-m:$))`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern.FindStringSubmatch("joe@example.com")
	}
}

// Import matches "import foo" and "import foo as bar"
type Import struct {
	_       struct{} `^import\s+`
	Package Submatch `\w+`
	Alias   *AsName  `?`
	_       struct{} `$`
}

// AsName matches "as xyz"
type AsName struct {
	_    struct{} `\s+as\s+`
	Name Submatch `\w+`
}

func BenchmarkFindAllImports(b *testing.B) {
	path := os.Getenv("TESTDATA")
	if path == "" {
		b.Skip("skipping because TESTDATA environment var was not set")
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		b.Error(err)
	}
	pattern := MustCompile(Import{}, Options{})
	var imports []Import
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern.FindAll(&imports, string(buf), -1)
	}
}

func BenchmarkFindAllImportsStdlib(b *testing.B) {
	path := os.Getenv("TESTDATA")
	if path == "" {
		b.Skip("skipping because TESTDATA environment var was not set")
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		b.Error(err)
	}
	pattern := regexp.MustCompile(`(\Aimport[\t-\n\f-\r ]+(?P<Package>[0-9A-Z_a-z]+)(?P<Alias>([\t-\n\f-\r ]+as[\t-\n\f-\r ]+(?P<Name>[0-9A-Z_a-z]+))?)(?-m:$))`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern.FindAllSubmatchIndex(buf, -1)
	}
}
