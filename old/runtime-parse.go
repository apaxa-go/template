package template

import (
	"errors"
	"github.com/apaxa-go/helper/strconvh"
	"io/ioutil"
	"log"
	"regexp"
)

const (
	OpeningTag        = "<?"
	CloseTag          = "?>"
	TemplatePartName1 = "[[:alnum:]]*"                  // Default name
	TemplatePartName2 = "(?:[[:alnum:]]*[[:alpha:]])?#" // Name with suffix request (end with '#')
	TemplatePartName  = "(?:" + TemplatePartName1 + ")|(?:" + TemplatePartName2 + ")"
	MainRegexp        = "[[:space:]]*(" + TemplatePartName + ")?[[:space:]]*(\\+p?o?s?)?[[:space:]]*"
)

type Options struct {
	Prefix   bool
	Optional bool
	Suffix   bool
}

func defaultOptions() (o Options) {
	o.Prefix = true
	o.Optional, o.Suffix = false, false
	return
}

func (o Options) String() string {
	replacer := map[bool]string{
		false: "no",
		true:  "yes",
	}
	return "prefix: " + replacer[o.Prefix] + ", optional: " + replacer[o.Optional] + ", suffix: " + replacer[o.Suffix]
}

var re = regexp.MustCompile(regexp.QuoteMeta(OpeningTag) + MainRegexp + regexp.QuoteMeta(CloseTag))

// TODO move to other package
func ByteIsNumChar(b byte) bool {
	return b >= '0' && b <= '9'
}

func ParseString(s string) (strings map[string]string, order []string, options map[string]Options, err error) {
	strings = make(map[string]string)
	options = make(map[string]Options)
	seps := re.FindAllStringSubmatchIndex(s, -1)

	// Check for omitted first splitter
	if len(seps) == 0 || seps[0][0] != 0 {
		seps = append([][]int{[]int{0, 0, -1, -1, -1, -1}}, seps...)
	}

	var lastSuffixes = make(map[string]int)

	for i, sep := range seps {
		// Name

		name := ""
		if sep[2] >= 0 && sep[3] >= 0 { // name may be omitted in template
			name = s[sep[2]:sep[3]]
		}

		// Name -> Base name [+ suffix]
		var baseName string
		var autoSuffixRequired bool = false
		var currentSuffix = -1

		//log.Print(name)
		switch {
		case name == "":
			//log.Print("Empty name")
			baseName = ""
			autoSuffixRequired = true
		case name[len(name)-1] == '#':
			//log.Print("#-ended name")
			baseName = name[:len(name)-1]
			autoSuffixRequired = true
		case ByteIsNumChar(name[len(name)-1]):
			//log.Print("Num-ended name")
			suffixLen := 1
			for suffixLen < len(name) && ByteIsNumChar(name[len(name)-1-suffixLen]) {
				suffixLen++
			}
			baseName = name[:len(name)-suffixLen]
			if currentSuffix, err = strconvh.ParseInt(name[len(name)-suffixLen:]); err != nil {
				err = errors.New("Unable to parse suffix: " + err.Error())
				return
			}
			//log.Print(name, " ", suffixLen, " ", baseName, " ", currentSuffix)
		default:
			//log.Print("Simple name")
			baseName = name
		}

		// Extract suffix from map if required
		if autoSuffixRequired {
			var ok bool
			if currentSuffix, ok = lastSuffixes[baseName]; !ok {
				currentSuffix = 0
			} else {
				currentSuffix++
			}
		}

		// Save current suffix to map
		if currentSuffix >= 0 {
			lastSuffixes[baseName] = currentSuffix
		}

		// Base name [+ suffix] -> name
		name = baseName
		if currentSuffix >= 0 {
			name += strconvh.FormatInt(currentSuffix)
		}

		// Check for name duplication
		if _, ok := strings[name]; ok {
			err = errors.New("Duplicate template name detected: " + name + ".")
			return
		}

		// options
		var opt Options
		if sep[4] >= 0 && sep[5] >= 0 { // Options passed
			//log.Print(s[sep[0]:sep[1]], sep[4], sep[5])
			optStr := s[sep[4]+1 : sep[5]] // Skip '+'
			//log.Print(optStr)
			//log.Print(len(optStr))
			for _, o := range optStr {
				//log.Print(o)
				switch o {
				case 'p':
					opt.Prefix = true
				case 'o':
					opt.Optional = true
				case 's':
					opt.Suffix = true
				default:
					err = errors.New("Unknown option '" + string(o) + "'")
					return
				}
			}
		} else {
			opt = defaultOptions()
		}

		// Value
		var end int
		if i+1 < len(seps) {
			end = seps[i+1][0]
		} else {
			end = len(s)
		}

		value := s[sep[1]:end]

		//

		strings[name] = value
		order = append(order, name)
		options[name] = opt
	}

	return
}

func ParseFile(filename string) (strings map[string]string, order []string, options map[string]Options, err error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	return ParseString(string(b))
}

func MustParseFile(filename string) (strings map[string]string, order []string, options map[string]Options) {
	strings, order, options, err := ParseFile(filename)
	if err != nil {
		log.Panic("Unable to parse template: ", err) // TODO change log.Panic to smth other?
	}
	return
}

/*
TODO implement this function
func CompileSimple(w http.ResponseWriter, filename string) error {
	t, o, err := template.ParseFile(filename)
	if err != nil {
		log.Print(err)
		return err
	}

	for _, name := range o {
		if _, err := w.Write([]byte(t[name])); err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}
*/
