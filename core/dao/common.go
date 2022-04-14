package dao

import (
	"fmt"
	"errors"
	"regexp"

	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"
)

type ResourceErrors[T any] struct {
	Resource *T
	Errors []error
}

type Resource interface {
	GetContext() string
	GetContextLine() int
}

// func (re *ResourceErrors[T]) Combine() error {
func CombineErrors(re Resource, errs []error) error {
	var msg = ""
	partsRe := regexp.MustCompile(`line (\d*): (.*)`)

	context := re.GetContext()
	contextLine := re.GetContextLine()

	var errPrefix = text.FgRed.Sprintf("error")
	var ptrPrefix = text.FgBlue.Sprintf("-->")
	for _, err := range errs {
		switch err.(type) {
		case *FoundCyclicDependency: // cyclic dependency error
			msg = fmt.Sprintf("%s: error", msg)
		case *yaml.TypeError: // yaml error
			yamlErrors := err.(*yaml.TypeError)
			for _, yErr := range yamlErrors.Errors {
				match := partsRe.FindStringSubmatch(yErr)
				// In-case matching fails, return unformatted error
				if len(match) != 3 {
					msg = fmt.Sprintf("%s%s: %s\n  %s %s\n\n", msg, errPrefix, yErr, ptrPrefix, context)
				} else {
					msg = fmt.Sprintf("%s%s: %s\n  %s %s:%s\n\n", msg, errPrefix, match[2], ptrPrefix, context, match[1])
				}
			}
		default: // default resource error
			msg = fmt.Sprintf("%s%s: %s\n  %s %s:%d\n\n", msg, errPrefix, err.Error(), ptrPrefix, context, contextLine)
		}
	}

	if msg != "" {
		return errors.New(msg)
	}

	return nil
}

func StringsToErrors(str []string) []error {
	errs := []error{}
	for _, s := range str {
		errs = append(errs, errors.New(s))
	}

	return errs
}
