package main

import (
	"fmt"

	"github.com/alexflint/go-restructure"
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

var floatRegexp = restructure.MustCompile(Float{}, restructure.Options{})

// Matches "123", "1.23", "1.23e-4", "-12.3E+5", ".123"
type Float struct {
	Begin    restructure.Pos
	Sign     *Sign     `?`
	Whole    string    `[0-9]*`
	Period   struct{}  `\.?`
	Frac     string    `[0-9]+`
	Exponent *Exponent `?`
	End      restructure.Pos
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

func main() {
	var floats []Float
	floatRegexp.FindAll(&floats, src, -1)
	for _, f := range floats {
		fmt.Println(src[f.Begin:f.End])
	}
}
