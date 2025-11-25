package ui

import (
	"fmt"
	"time"
)

// Spinner provides a simple loading indicator
type Spinner struct {
	message string
	active  bool
	done    chan bool
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		done:    make(chan bool),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.active = true

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0

		for {
			select {
			case <-s.done:
				return
			default:
				if s.active {
					fmt.Printf("\r%s %s", frames[i], s.message)
					i = (i + 1) % len(frames)
					time.Sleep(80 * time.Millisecond)
				}
			}
		}
	}()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.active = false
	s.done <- true
	fmt.Print("\r\033[K") // Clear line
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.message = message
}
