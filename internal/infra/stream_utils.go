package infra

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type ServerSentEvent struct {
	Event string
	Data  string
	ID    string
	Retry string
}

// SSEWriter helps writing SSE events
func WriteSSE(w io.Writer, event ServerSentEvent) error {
	var buf bytes.Buffer
	if event.Event != "" {
		fmt.Fprintf(&buf, "event: %s\n", event.Event)
	}
	if event.ID != "" {
		fmt.Fprintf(&buf, "id: %s\n", event.ID)
	}
	if event.Retry != "" {
		fmt.Fprintf(&buf, "retry: %s\n", event.Retry)
	}
	// Handle multi-line data
	if event.Data != "" {
		lines := strings.Split(event.Data, "\n")
		for _, line := range lines {
			fmt.Fprintf(&buf, "data: %s\n", line)
		}
	}
	fmt.Fprint(&buf, "\n")
	_, err := w.Write(buf.Bytes())
	return err
}

type SSEEventTransformer func(ServerSentEvent) ([]ServerSentEvent, error)

func NewSSETransformer(upstream io.Reader, transformer SSEEventTransformer) io.Reader {
	r, w := io.Pipe()

	go func() {
		defer w.Close()
		scanner := bufio.NewScanner(upstream)

		var currentEvent ServerSentEvent
		inEvent := false

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				if inEvent {
					// End of event, process it
					newEvents, err := transformer(currentEvent)
					if err == nil {
						for _, evt := range newEvents {
							if err := WriteSSE(w, evt); err != nil {
								return
							}
						}
					}
					currentEvent = ServerSentEvent{}
					inEvent = false
				} else {
                    // Keep alive or extra newlines
                    w.Write([]byte("\n"))
                }
				continue
			}

			inEvent = true
			if strings.HasPrefix(line, "event: ") {
				currentEvent.Event = strings.TrimPrefix(line, "event: ")
			} else if strings.HasPrefix(line, "data: ") {
				if currentEvent.Data != "" {
					currentEvent.Data += "\n"
				}
				currentEvent.Data += strings.TrimPrefix(line, "data: ")
			} else if strings.HasPrefix(line, "id: ") {
				currentEvent.ID = strings.TrimPrefix(line, "id: ")
			} else if strings.HasPrefix(line, "retry: ") {
				currentEvent.Retry = strings.TrimPrefix(line, "retry: ")
			} else if strings.HasPrefix(line, ":") {
                // Comment, ignore or pass through? Ignore for now
            }
		}
		// Handle last event if no trailing newline
		if inEvent {
			newEvents, err := transformer(currentEvent)
			if err == nil {
				for _, evt := range newEvents {
					WriteSSE(w, evt)
				}
			}
		}
	}()

	return r
}
