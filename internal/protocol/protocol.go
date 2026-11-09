package protocol

import "strings"

type Command struct {
	Name string
	Args string
}

type DM struct {
	To      string
	Message string
}

func ParseCommand(message string) (Command, bool) {
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

func ParseDM(args string) (DM, bool) {
	line := strings.TrimSpace(args)
	if line == "" {
		return DM{}, false
	}

	parts := strings.SplitN(line, " ", 2)
	dm := DM{To: parts[0]}
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return DM{}, false
	}
	if len(parts) > 1 {
		dm.Message = parts[1]
	}
	return dm, true
}
