//go:build windows
// +build windows

package exec

func ExecTTY(cmd string, envs []string) error {
	return nil
}
