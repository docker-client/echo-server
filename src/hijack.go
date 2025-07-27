package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Action interface {
	Execute(w *bufio.ReadWriter, index int) (bool, error) // returns whether to continue
}

type ResponseAction struct {
	Text string
}

type DelayAction struct {
	Duration time.Duration
}

type CloseAction struct {
}

func (a ResponseAction) Execute(w *bufio.ReadWriter, index int) (bool, error) {
	_, err := fmt.Fprintf(w, "[%d] scripted: %s\r\n", index, a.Text)
	w.Flush()
	return true, err
}

func (a DelayAction) Execute(w *bufio.ReadWriter, index int) (bool, error) {
	time.Sleep(a.Duration)
	return true, nil
}

func (a CloseAction) Execute(w *bufio.ReadWriter, index int) (bool, error) {
	_, err := fmt.Fprintf(w, "[%d] scripted: closing connection\r\n", index)
	w.Flush()
	return false, err
}

type RawAction struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Delay int    `json:"ms,omitempty"`
}

type Script struct {
	Actions []RawAction `json:"actions"`
}

func parseScript(jsonBody io.Reader) ([]Action, error) {
	var script Script
	decoder := json.NewDecoder(jsonBody)
	if err := decoder.Decode(&script); err != nil {
		return nil, err
	}

	var actions []Action
	for _, raw := range script.Actions {
		switch strings.ToLower(raw.Type) {
		case "response":
			actions = append(actions, ResponseAction{Text: raw.Text})
		case "delay":
			actions = append(actions, DelayAction{Duration: time.Duration(raw.Delay) * time.Millisecond})
		case "close":
			actions = append(actions, CloseAction{})
		default:
			actions = append(actions, ResponseAction{Text: "[unknown action: " + raw.Type + "]"})
		}
	}
	return actions, nil
}

func ParseScriptFromQuery(q url.Values) []Action {
	var keys []string
	for k := range q {
		keys = append(keys, k)
	}
	// Sort keys for deterministic order (e.g. res1, res2, delay1, ...)
	//sort.Strings(keys)

	var actions []Action
	for _, k := range keys {
		vals := q[k]
		for _, val := range vals {
			switch {
			case strings.HasPrefix(k, "res"):
				actions = append(actions, ResponseAction{Text: val})
			case strings.HasPrefix(k, "delay"):
				if ms, err := strconv.Atoi(val); err == nil {
					actions = append(actions, DelayAction{Duration: time.Duration(ms) * time.Millisecond})
				}
			case k == "close":
				actions = append(actions, CloseAction{})
			}
		}
	}
	return actions
}

func DefaultScript(counterStart *int) iter.Seq2[int, Action] {
	return func(yield func(int, Action) bool) {
		i := 0
		for {
			val := *counterStart
			if !yield(i, ResponseAction{Text: fmt.Sprintf("counter: %d", val)}) {
				return
			}
			if !yield(i, DelayAction{Duration: 750 * time.Millisecond}) {
				return
			}
			*counterStart++
			i++
		}
	}
}

func readCommands(reader *bufio.Reader, out chan<- string) {
	defer close(out)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		out <- strings.TrimSpace(line)
	}
}

func upgradeHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.EqualFold(r.Header.Get("Connection"), "Upgrade") ||
		!strings.EqualFold(r.Header.Get("Upgrade"), "testproto") {
		http.Error(w, "Upgrade required", http.StatusUpgradeRequired)
		return
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Switching Protocols")

	// Send Upgrade response
	fmt.Fprintf(bufrw, "HTTP/1.1 101 Switching Protocols\r\n")
	fmt.Fprintf(bufrw, "Upgrade: testproto\r\n")
	fmt.Fprintf(bufrw, "Connection: Upgrade\r\n\r\n")
	bufrw.Flush()

	counter := 0
	var actions iter.Seq2[int, Action]
	useInteractive := false

	customActions, err := parseScript(r.Body)
	if (err != nil || len(customActions) == 0) && len(r.URL.Query()) > 0 {
		customActions = ParseScriptFromQuery(r.URL.Query())
	}

	if err != nil || len(customActions) == 0 {
		fmt.Fprintf(bufrw, "[X] Error parsing script or no actions, using default behavior.\r\n")
		bufrw.Flush()
		actions = DefaultScript(&counter)
		useInteractive = true
	} else {
		log.Printf("Actions: %v\n", customActions)
		actions = func(yield func(int, Action) bool) {
			for i, action := range customActions {
				if !yield(i, action) {
					return
				}
			}
			//			return
		}
	}

	var commandCh chan string
	if useInteractive {
		commandCh = make(chan string)
		go readCommands(bufrw.Reader, commandCh)
	}

	for i, action := range actions {
		if useInteractive {
			select {
			case cmd, ok := <-commandCh:
				if !ok {
					conn.Close()
					return
				}
				switch {
				case cmd == "reset":
					counter = 0
					fmt.Fprintf(bufrw, "[%d] counter reset\r\n", i)
					bufrw.Flush()
					continue
				case strings.HasPrefix(cmd, "set "):
					parts := strings.SplitN(cmd, " ", 2)
					if val, err := strconv.Atoi(parts[1]); err == nil {
						counter = val
						fmt.Fprintf(bufrw, "[%d] counter set to %d\r\n", i, val)
						bufrw.Flush()
						continue
					}
				default:
					fmt.Fprintf(bufrw, "[%d] unknown command: %s\r\n", i, cmd)
					bufrw.Flush()
				}
			default:
				// no input, continue with scripted action
			}
		}

		cont, err := action.Execute(bufrw, i)
		if err != nil || !cont {
			conn.Close()
			return
		}
	}

	conn.Close()
}
