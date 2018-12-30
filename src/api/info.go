package main

import (
	"strings"
)

func BuildInfo() (string, error) {
	buf := new(strings.Builder)

	// this part looks fucking terrible, we gotta be able to do better than this
	if _, err := buf.WriteString(Minor); err != nil {
		return "NULL", err
	}

	if _, err := buf.WriteString("."); err != nil {
		return "NULL", err
	}

	if _, err := buf.WriteString(Patch); err != nil {
		return "NULL", err
	}

	if _, err := buf.WriteString("."); err != nil {
		return "NULL", err
	}

	if _, err := buf.WriteString(Major); err != nil {
		return "NULL", err
	}

	if _, err := buf.WriteString(Extra); err != nil {
		return "NULL", err
	}

	if _, err := buf.WriteString(", commit: "); err != nil {
		return "NULL", err
	}

	if _, err := buf.WriteString(Commit); err != nil {
		return "NULL", err
	}

	if _, err := buf.WriteString(", compiled "); err != nil {
		return "NULL", err
	}

	if _, err := buf.WriteString(Date); err != nil {
		return "NULL", err
	}

	return buf.String(), nil
}
