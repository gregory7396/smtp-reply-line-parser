# smtp-reply-line-parser

Parses raw SMTP reply text into structured multiline response codes and messages, handling continuation lines.

```go
package main

import (
	"fmt"
	"strings"

	smtp_reply_line_parser "smtp-reply-line-parser"
)

func main() {
	raw := "250-First line\r\n250-Second line\r\n250 Done\r\n"
	replies, err := smtp_reply_line_parser.Parse(strings.NewReader(raw))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, r := range replies {
		fmt.Printf("%d %s\n", r.Code, r.Message)
	}
	// Output: 250 First line Second line Done
}
```

## Why

SMTP replies can span several lines. Each line begins with a three-digit code followed by either a hyphen (meaning "more lines for this reply follow") or a space (meaning "this is the final line"). Raw socket reads don't give you reply boundaries — you get bytes and have to reconstruct the structure yourself.

This library does that one job. It reads from an `io.Reader`, groups continuation lines under their reply, and returns a `[]Reply` where each `Reply` has a `Code` (int) and a joined `Message` (string).

## Edge cases

If the input ends mid-reply — i.e. the last line used a hyphen separator and no final space-separated line arrived — the partial reply is returned along with `ErrUnterminatedReply`. This lets a caller buffer more input and retry rather than silently truncating.

Continuation lines whose code differs from the reply's opening code produce an error, since that indicates a protocol violation rather than a new reply.

Message text is trimmed of surrounding whitespace per line. When multiple continuation lines are joined, they are separated by a single space.
