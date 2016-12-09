package template

import (
	"errors"
	"github.com/apaxa-go/helper/strconvh"
	"io"
	"reflect"
	"regexp"
)

func Execute(w io.Writer, strings map[string]string, order []string, options map[string]Options, arguments ...interface{}) error {
	argNum := 0
	for _, name := range order {
		//log.Print(argNum)
		// Prefix
		if options[name].Prefix {
			if p, ok := arguments[argNum].(string); ok {
				if p != "" {
					if _, err := w.Write([]byte(p)); err != nil {
						return errors.New("While writing prefix for template \"" + name + "\": " + err.Error())
					}
				}
			} else {
				//log.Print(arguments[argNum])
				//log.Print(strings[name])
				//log.Print(options[name])
				return errors.New("Prefix for template \"" + name + "\" is not string (argument #" + strconvh.FormatInt(argNum) + "), it is of type " + reflect.TypeOf(arguments[argNum]).String())
			}
			argNum++
		}

		// It self
		writeItself := true
		if options[name].Optional {
			if b, ok := arguments[argNum].(bool); ok {
				writeItself = b
			} else {
				return errors.New("Optional flag for template \"" + name + "\" is not bool (argument #" + strconvh.FormatInt(argNum) + "), it is of type " + reflect.TypeOf(arguments[argNum]).String())
			}
			argNum++
		}
		if len(strings[name]) > 0 && writeItself { // Ignore empty template part
			if _, err := w.Write([]byte(strings[name])); err != nil {
				return errors.New("While writing template \"" + name + "\": " + err.Error())
			}
		}

		// Suffix
		if options[name].Suffix {
			if s, ok := arguments[argNum].(string); ok {
				if s != "" {
					if _, err := w.Write([]byte(s)); err != nil {
						return errors.New("While writing suffix for template \"" + name + "\": " + err.Error())
					}
				}
			} else {
				return errors.New("Suffix for template \"" + name + "\" is not string (argument #" + strconvh.FormatInt(argNum) + "), it is of type " + reflect.TypeOf(arguments[argNum]).String())
			}
			argNum++
		}
	}
	if argNum!=len(arguments){
		return errors.New("Too many arguments: required "+strconvh.FormatInt(argNum)+", but given "+strconvh.FormatInt(len(arguments)))
	}
	return nil
}

func ExecuteExcept(w io.Writer, strings map[string]string, order []string, options map[string]Options, except string, arguments ...interface{}) error {
	re, err := regexp.Compile(except)
	if err != nil {
		return err
	}
	var newOrder []string
	for _, s := range order {
		if !re.MatchString(s) {
			newOrder = append(newOrder, s)
		}
	}
	return Execute(w, strings, newOrder, options, arguments...)
}

func ExecuteOnly(w io.Writer, strings map[string]string, order []string, options map[string]Options, only string, arguments ...interface{}) error {
	re, err := regexp.Compile(only)
	if err != nil {
		return err
	}
	var newOrder []string
	for _, s := range order {
		if re.MatchString(s) {
			newOrder = append(newOrder, s)
		}
	}
	return Execute(w, strings, newOrder, options, arguments...)
}

func ExecuteFile(w io.Writer, filename string, arguments ...interface{}) error {
	strings, order, options, err := ParseFile(filename)
	if err != nil {
		return err
	}
	return Execute(w, strings, order, options, arguments...)
}

func ExecuteFileExcept(w io.Writer, filename string, except string, arguments ...interface{}) error {
	strings, order, options, err := ParseFile(filename)
	if err != nil {
		return err
	}
	return ExecuteExcept(w, strings, order, options, except, arguments...)
}

func ExecuteFileOnly(w io.Writer, filename string, only string, arguments ...interface{}) error {
	strings, order, options, err := ParseFile(filename)
	if err != nil {
		return err
	}
	return ExecuteOnly(w, strings, order, options, only, arguments...)
}
