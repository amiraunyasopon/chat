package protocol

import "strings"

type Command struct {
	Name string
	Args string
}

func Parse(message string) (Command, bool) {
	line := strings.TrimSpace(message)
	if line == "" {
		return Command{}, false
	}

	parts := strings.SplitN(line, " ", 2)
	cmd := Command{Name: strings.ToUpper(parts[0])}
	if len(parts) > 1 {
		cmd.Args = parts[1]
	}
	return cmd, true
}
