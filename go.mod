module github.com/alajmo/mani

go 1.16

require (
	github.com/jedib0t/go-pretty/v6 v6.2.4
	github.com/kr/pretty v0.2.1
	github.com/kr/text v0.2.0 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/otiai10/copy v1.6.0
	github.com/spf13/cobra v1.1.3
	github.com/theckman/yacspin v0.8.0
	github.com/alajmo/goph v1.2.2 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace (
	github.com/alajmo/goph v1.2.2 => ../forks/goph
)
