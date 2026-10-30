package smtp_reply_line_parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Reply represents a single parsed SMTP reply consisting of a numeric
// three-digit code and the message text. When the reply spans multiple
// physical lines, Message contains the joined message text with all
// individual messages separated by a single space.
type Reply struct {
	Code    int
	Message string
}

// Parse reads raw SMTP reply text from r and returns the replies it
// contains. A reply may span multiple lines: an SMTP reply line uses a '-'
// separator between the code and text to indicate that more lines for the
// same reply follow, and a ' ' separator to mark the final line.
//
// If the last reply on the input is unterminated (its continuation lines
// never end with a final ' ' line), that partial reply is returned with an
// error of kind ErrUnterminatedReply so callers can decide whether to wait
// for more input or reject the connection.
func Parse(r io.Reader) ([]Reply, error) {
	reader := bufio.NewReader(r)
	var replies []Reply
	current := Reply{}
	hasCurrent := false

	for {
		line, err := reader.ReadString('\n')
		hadLine := len(line) > 0

		if hadLine {
			code, msg, cont, perr := parseLine(line)
			if perr != nil {
				return nil, fmt.Errorf("line %d: %w", len(replies)+1, perr)
			}

			if !hasCurrent {
				current = Reply{Code: code, Message: msg}
				hasCurrent = true
			} else if current.Code != code {
				return nil, fmt.Errorf("line %d: expected code %d, got %d", len(replies)+1, current.Code, code)
			} else {
				current.Message = current.Message + " " + msg
			}

			if !cont {
				replies = append(replies, current)
				current = Reply{}
				hasCurrent = false
			}
		}

		if err != nil {
			if err == io.EOF {
				if hasCurrent {
					replies = append(replies, current)
					return replies, ErrUnterminatedReply
				}
				return replies, nil
			}
			return replies, fmt.Errorf("read error: %w", err)
		}
	}
}

// ErrUnterminatedReply indicates that the input ended while a reply was
// still expecting continuation lines. The partial reply is included in the
// returned slice so callers can inspect what was received.
var ErrUnterminatedReply = errors.New("smtp reply line parser: unterminated reply")

// parseLine parses a single SMTP reply line and returns the numeric code,
// the message text, a continuation flag, and any parse error. A line is
// expected to begin with exactly three digits followed by either a space
// (final line) or a hyphen (continuation line), then the message text.
//
// We intentionally accept only ASCII digits for the code. SMTP is defined
// over ASCII, and accepting broader numeric forms would mask genuine
// protocol violations where a server emits non-digit characters.
func parseLine(raw string) (code int, msg string, cont bool, err error) {
	line := strings.TrimRight(raw, "\r\n")

	if len(line) < 4 {
		return 0, "", false, fmt.Errorf("line too short: %q", line)
	}

	for i := 0; i < 3; i++ {
		if line[i] < '0' || line[i] > '9' {
			return 0, "", false, fmt.Errorf("invalid code digit at position %d in %q", i+1, line)
		}
	}

	sep := line[3]
	if sep != ' ' && sep != '-' {
		return 0, "", false, fmt.Errorf("expected space or hyphen after code, got %q", string(sep))
	}

	code = int(line[0]-'0')*100 + int(line[1]-'0')*10 + int(line[2]-'0')
	msg = strings.TrimSpace(line[4:])
	cont = sep == '-'
	return code, msg, cont, nil
}
