// This source will generate
//   - core/mani.1
//   - docs/commands.md
//
// and is not included in the final build.

package core

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

//go:embed config.man
var CONFIG_MD []byte

type genManHeaders struct {
	Title   string
	Section string
	Date    string
	Source  string
	Manual  string
	Version string
	Desc    string
}

func CreateManPage(desc string, version string, date string, rootCmd *cobra.Command, cmds ...*cobra.Command) error {
	header := &genManHeaders{
		Title:   "MANI",
		Section: "1",
		Source:  "Mani Manual",
		Manual:  "mani",
		Version: version,
		Date:    date,
		Desc:    desc,
	}

	res := genMan(header, rootCmd, cmds...)
	res = append(res, CONFIG_MD...)
	manPath := filepath.Join("./core/", "mani.1")
	err := ioutil.WriteFile(manPath, res, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Created %s\n", manPath)

	md, err := genDoc(rootCmd, cmds...)
	if err != nil {
		return err
	}

	mdPath := filepath.Join("./docs/", "commands.md")
	err = ioutil.WriteFile(mdPath, md, 0644)
	if err != nil {
		return err
	}
	fmt.Printf("Created %s\n", mdPath)

	return nil
}

func manPreamble(buf io.StringWriter, header *genManHeaders, cmd *cobra.Command, dashedName string) {
	preamble := `.TH "%s" "%s" "%s" "%s" "%s" "%s"`
	cobra.WriteStringAndCheck(buf, fmt.Sprintf(preamble, header.Title, header.Section, header.Date, header.Version, header.Source, header.Manual))

	cobra.WriteStringAndCheck(buf, "\n")

	cobra.WriteStringAndCheck(buf, ".SH NAME\n")
	cobra.WriteStringAndCheck(buf, fmt.Sprintf("%s - %s\n", header.Manual, cmd.Short))

	cobra.WriteStringAndCheck(buf, "\n")

	cobra.WriteStringAndCheck(buf, ".SH SYNOPSIS\n")
	cobra.WriteStringAndCheck(buf, ".B mani [command] [flags]\n")

	cobra.WriteStringAndCheck(buf, "\n")

	cobra.WriteStringAndCheck(buf, ".SH DESCRIPTION\n")

	cobra.WriteStringAndCheck(buf, header.Desc+"\n\n")
}

func manCommand(buf io.StringWriter, cmd *cobra.Command, dashedName string) {
	cobra.WriteStringAndCheck(buf, ".TP\n")
	cobra.WriteStringAndCheck(buf, fmt.Sprintf(`.B %s`, cmd.UseLine()))
	cobra.WriteStringAndCheck(buf, "\n")
	cobra.WriteStringAndCheck(buf, fmt.Sprintf("%s\n\n", cmd.Long))

	nonInheritedFlags := cmd.NonInheritedFlags()
	inheritedFlags := cmd.InheritedFlags()
	if !nonInheritedFlags.HasAvailableFlags() && !inheritedFlags.HasAvailableFlags() {
		return
	}

	cobra.WriteStringAndCheck(buf, "\n.B Available Options:\n")
	cobra.WriteStringAndCheck(buf, ".RS\n")
	cobra.WriteStringAndCheck(buf, ".RS\n")
	if nonInheritedFlags.HasAvailableFlags() {
		manPrintFlags(buf, nonInheritedFlags)
	}

	if inheritedFlags.HasAvailableFlags() && cmd.Name() != "gen" {
		manPrintFlags(buf, inheritedFlags)
		cobra.WriteStringAndCheck(buf, "\n")
	}

	cobra.WriteStringAndCheck(buf, ".RE\n")
	cobra.WriteStringAndCheck(buf, ".RE\n")
}

func manPrintFlags(buf io.StringWriter, flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		if len(flag.Deprecated) > 0 || flag.Hidden {
			return
		}
		format := ""

		if len(flag.Shorthand) > 0 && len(flag.ShorthandDeprecated) == 0 {
			format = fmt.Sprintf("-%s, --%s", flag.Shorthand, flag.Name)
		} else {
			format = fmt.Sprintf("--%s", flag.Name)
		}
		if len(flag.NoOptDefVal) > 0 {
			format += "["
		}
		if flag.Value.Type() == "string" {
			// put quotes on the value
			format += "=%q"
		} else {
			format += "=%s"
		}
		if len(flag.NoOptDefVal) > 0 {
			format += "]"
		}

		format = fmt.Sprintf(`\fB%s\fR`, format)
		format = fmt.Sprintf(format, flag.DefValue)
		format = fmt.Sprintf(".TP\n%s\n%s\n", format, flag.Usage)
		cobra.WriteStringAndCheck(buf, format)
	})
}

func genMan(header *genManHeaders, cmd *cobra.Command, cmds ...*cobra.Command) []byte {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)

	// PREAMBLE
	manPreamble(buf, header, cmd, cmd.CommandPath())
	flags := cmd.NonInheritedFlags()

	// OPTIONS
	cobra.WriteStringAndCheck(buf, ".SH OPTIONS\n")

	// FLAGS
	manPrintFlags(buf, flags)

	buf.WriteString(".SH\nCOMMANDS\n")

	// COMMANDS
	for _, c := range cmds {
		dashCommandName := c.CommandPath()

		cbuf := new(bytes.Buffer)

		if !StringInSlice(c.Name(), []string{"list", "describe"}) {
			manCommand(cbuf, c, dashCommandName)
		}

		if len(c.Commands()) > 0 {
			for _, cc := range c.Commands() {
				// Don't include help command
				if cc.Name() != "help" {
					manCommand(cbuf, cc, dashCommandName)
				}
			}
		}

		buf.Write(cbuf.Bytes())
	}

	return buf.Bytes()
}

func genDoc(cmd *cobra.Command, cmds ...*cobra.Command) ([]byte, error) {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	out := new(bytes.Buffer)
	err := doc.GenMarkdown(cmd, out)
	if err != nil {
		return []byte{}, err
	}

	md := out.String()
	md = strings.Split(md, "### SEE ALSO")[0]
	md = fmt.Sprintf("%s\n\n%s", "# Commands", md)

	for _, c := range cmds {
		if !StringInSlice(c.Name(), []string{"list", "describe"}) {
			cOut := new(bytes.Buffer)
			err := doc.GenMarkdown(c, cOut)
			if err != nil {
				return []byte{}, err
			}

			cMd := cOut.String()
			cMd = strings.Split(cMd, "### SEE ALSO")[0]
			md += cMd
		}

		if len(c.Commands()) > 0 {
			for _, cc := range c.Commands() {
				// Don't include help command
				if cc.Name() != "help" {
					ccOut := new(bytes.Buffer)
					err := doc.GenMarkdown(cc, ccOut)
					if err != nil {
						return []byte{}, err
					}
					ccMd := ccOut.String()
					ccMd = strings.Split(ccMd, "### SEE ALSO")[0]
					md += ccMd
				}
			}
		}
	}

	return []byte(md), nil
}
