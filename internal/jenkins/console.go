package jenkins

import (
	"bytes"
	"context"
	"fmt"
)

// GetConsole fetches the plain-text console log of a build via
// {JobPath(jobName)}/{number}/consoleText.
func (c *Client) GetConsole(ctx context.Context, jobName string, number int) ([]byte, error) {
	path := fmt.Sprintf("%s/%d/consoleText", JobPath(jobName), number)
	b, err := c.GET(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get console %s #%d: %w", jobName, number, err)
	}
	return b, nil
}

// LastNLines returns the last n newline-terminated (or trailing) lines of b.
//
//   - n <= 0           → nil
//   - n >= total lines → bytes.Clone(b) (the entire input, copied)
//   - otherwise        → suffix starting after the nth-from-end newline
//     that is NOT the final byte (so a trailing
//     no-newline line is preserved).
func LastNLines(b []byte, n int) []byte {
	if n <= 0 {
		return nil
	}
	count := 0
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] == '\n' && i != len(b)-1 {
			count++
			if count == n {
				return b[i+1:]
			}
		}
	}
	// fewer than n lines exist
	return bytes.Clone(b)
}
