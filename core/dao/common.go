package dao

import (
	"fmt"
	"errors"
	"regexp"

	"github.com/jedib0t/go-pretty/v6/text"
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
func FormatErrors(re Resource, errs []error) error {
	var msg = ""
	partsRe := regexp.MustCompile(`line (\d*): (.*)`)

	context := re.GetContext()

	var errPrefix = text.FgRed.Sprintf("error")
	var ptrPrefix = text.FgBlue.Sprintf("-->")
	for _, err := range errs {
		match := partsRe.FindStringSubmatch(err.Error())
		// In-case matching fails, return unformatted error
		if len(match) != 3 {
			contextLine := re.GetContextLine()

			if contextLine == -1 {
				msg = fmt.Sprintf("%s%s: %s\n  %s %s\n\n", msg, errPrefix, err, ptrPrefix, context)
			} else {
				msg = fmt.Sprintf("%s%s: %s\n  %s %s:%d\n\n", msg, errPrefix, err, ptrPrefix, context, contextLine)
			}
		} else {
			msg = fmt.Sprintf("%s%s: %s\n  %s %s:%s\n\n", msg, errPrefix, match[2], ptrPrefix, context, match[1])
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
