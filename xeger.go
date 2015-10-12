package xeger

import (
	"fmt"
	"log"
	"regexp"
	"regexp/syntax"
	"strconv"
	"unicode"
)

type Xeger struct {
	re     *syntax.Regexp
	logger Logger
}

func (x *Xeger) Generate() string {
	x.logger.Printf("regex: %s", x.re.String())
	x.logger.Printf("sub %v", x.re.Sub)

	var regexStr string
	for _, r := range x.re.Sub {
		x.logger.Println(">")
		regexStr += makeMatch(*r)
		x.logger.Println(r.String())
		x.logger.Printf("\t op   %s [%v]", OpName(r.Op), r.Op)
		x.logger.Printf("\t rune %v", string(r.Rune))
	}
	x.logger.Printf("potenially match: `%s`", regexStr)
	x.logger.Println()

	return regexStr
}

func NewInverseRegex(s string) (*Xeger, error) {
	_, err := regexp.Compile(s)
	if err != nil {
		return nil, err
	}
	re, err := syntax.Parse(s, syntax.POSIX)
	if err != nil {
		return nil, err
	}
	simp := re.Simplify()

	return &Xeger{re: simp, logger: nopLogger{}}, nil
}

func OpName(op syntax.Op) string {
	switch op {
	case 1:
		return "OpNoMatch"
	case 2:
		return "OpEmptyMatch"
	case 3:
		return "OpLiteral"
	case 4:
		return "OpCharClass"
	case 5:
		return "OpAnyCharNotNL"
	case 6:
		return "OpAnyChar"
	case 7:
		return "OpBeginLine"
	case 8:
		return "OpBeginText"
	case 9:
		return "OpWrodBoundary"
	case 10:
		return "OpNoWordBoundary"
	case 11:
		return "OpCapture"
	case 12:
		return "OpStar"
	case 13:
		return "OpPlus"
	case 14:
		return "OpQuest"
	case 15:
		return "OpRepeat"
	case 16:
		return "OpConcat"
	case 17:
		return "OpAlternate"
	default:
		return "OpUnknown"
	}
}

func makeMatch(re syntax.Regexp) string {
	switch re.Op {
	default:
		return fmt.Sprintf("<invalid op" + strconv.Itoa(int(re.Op)) + ">")
	case syntax.OpNoMatch:
		log.Println("OpNoMatch")
		return ""
	case syntax.OpEmptyMatch:
		log.Println("OpEmptyMatch")
		return ""
	case syntax.OpLiteral:
		log.Println("OpLiteral")
		if re.Flags&syntax.FoldCase != 0 {
			// b.WriteString(`(?i:`)
		}
		return string(re.Rune)
	case syntax.OpCharClass:
		log.Println("OpCharClass")

		// b.WriteRune('[')
		if len(re.Rune) == 0 {
			// b.WriteString(`^\x00-\x{10FFFF}`)
		} else if re.Rune[0] == 0 && re.Rune[len(re.Rune)-1] == unicode.MaxRune {
			// Contains 0 and MaxRune.  Probably a negated class.
			// Print the gaps.
			// b.WriteRune('^')
			for i := 1; i < len(re.Rune)-1; i += 2 {
				lo, hi := re.Rune[i]+1, re.Rune[i+1]-1
				// escape(b, lo, lo == '-')
				if lo != hi {
					// b.WriteRune('-')
					// escape(b, hi, hi == '-')
				}
			}
		} else {
			for i := 0; i < len(re.Rune); i += 2 {
				lo, hi := re.Rune[i], re.Rune[i+1]
				// escape(b, lo, lo == '-')
				if lo != hi {
					// b.WriteRune('-')
					// escape(b, hi, hi == '-')
				}
			}
		}
		// b.WriteRune(']')
	case syntax.OpAnyCharNotNL:
		log.Println("OpAnyCharNotNL")
		return "abc"
	case syntax.OpAnyChar:
		log.Println("OpAnyChar")
		return "abc" // and sometimes nl
	case syntax.OpBeginLine:
		log.Println("OpBeginLine")
		// b.WriteRune('^') // make sure this is first?
	case syntax.OpEndLine:
		log.Println("OpEndLine")
		// b.WriteRune('$') // make sure this is last?
	case syntax.OpBeginText:
		log.Println("OpBeginText")
		// b.WriteString(`\A`)
	case syntax.OpEndText:
		log.Println("OpEndText")
		if re.Flags&syntax.WasDollar != 0 {
			// b.WriteString(`(?-m:$)`)
		} else {
			// b.WriteString(`\z`)
		}
	case syntax.OpWordBoundary:
		log.Println("OpWordBoundary")
		return " "
	case syntax.OpNoWordBoundary:
		log.Println("OpNoWordBoundary")
		// b.WriteString(`\B`)
	case syntax.OpCapture:
		log.Println("OpCapture")
		fallthrough
		// if re.Name != "" {
		// 	b.WriteString(`(?P<`)
		// 	b.WriteString(re.Name)
		// 	b.WriteRune('>')
		// } else {
		// 	b.WriteRune('(')
		// }
		// if re.Sub[0].Op != syntax.OpEmptyMatch {
		// 	// writeRegexp(b, re.Sub[0])
		// }
		// b.WriteRune(')')
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest, syntax.OpRepeat:
		log.Println("OpRepeat")
		if sub := re.Sub[0]; sub.Op > syntax.OpCapture || sub.Op == syntax.OpLiteral && len(sub.Rune) > 1 {
			// b.WriteString(`(?:`)
			// writeRegexp(b, sub)
			log.Println("named inner stuff to expand")
			// b.WriteString(`)`)
		} else {
			// writeRegexp(b, sub)
			log.Println("inner stuff to expand")
		}
		// this is the logics!
		thing := re.Sub
		thing2 := re.Sub0

		for _, t := range thing {
			log.Printf(" _>> %v", t)
		}

		for _, t := range thing2 {
			log.Printf(" 0>> %v", t)
		}

		switch re.Op {
		case syntax.OpStar:
			log.Println("OpStar")
			str := string(re.Rune)
			return str + str + str
		case syntax.OpPlus:
			log.Println("OpPlus")
			log.Println("op plus...?")
			return string(re.Rune)
		case syntax.OpQuest:
			log.Println("OpQuest")
			return string(re.Rune)
			// sometimes not
		case syntax.OpRepeat:
			log.Println("OpRepeat")
			// b.WriteRune('{')
			str := ""
			for i := 0; i < re.Min; i++ {
				str += string(re.Rune)
			}
			// consider rand between min and max
		}
		if re.Flags&syntax.NonGreedy != 0 {
			// b.WriteRune('?')
		}
	case syntax.OpConcat:
		log.Println("OpConcat")
		for _, sub := range re.Sub {
			if sub.Op == syntax.OpAlternate {
				// b.WriteString(`(?:`)
				// writeRegexp(b, sub)
				// b.WriteString(`)`)
			} else {
				// writeRegexp(b, sub)
			}
		}
	case syntax.OpAlternate:
		log.Println("OpAlternate")
		for i, sub := range re.Sub {
			if i > 0 {
				// this is specail. huh. sometimes write the second. What is the second?
				// b.WriteRune('|')
			}
			_ = sub
			// writeRegexp(b, sub)
		}
	}
	return ""
}

// generate takes in tokens in the form of:
// [a-z]
// [0-9a-z]
// [0-9][0-9][0-9](?:[0-9][0-9]?)?
// (?-s:.)
// x*
func generate(token string) string {

	return ""
}

type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type nopLogger struct{}

func (nopLogger) Print(v ...interface{})                 {}
func (nopLogger) Printf(format string, v ...interface{}) {}
func (nopLogger) Println(v ...interface{})               {}
