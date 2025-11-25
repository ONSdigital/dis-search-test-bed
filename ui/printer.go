package ui

import (
	"fmt"
	"os"
)

// Printer handles formatted output to the console
type Printer struct {
	verbose bool
}

// NewPrinter creates a new printer
func NewPrinter(verbose bool) *Printer {
	return &Printer{verbose: verbose}
}

// Info prints an informational message
func (p *Printer) Info(format string, args ...interface{}) {
	fmt.Printf("ℹ️  "+format+"\n", args...)
}

// Success prints a success message
func (p *Printer) Success(format string, args ...interface{}) {
	fmt.Printf("✅ "+format+"\n", args...)
}

// Warning prints a warning message
func (p *Printer) Warning(format string, args ...interface{}) {
	fmt.Printf("⚠️  "+format+"\n", args...)
}

// Error prints an error message
func (p *Printer) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "❌ "+format+"\n", args...)
}

// Debug prints a debug message (only if verbose)
func (p *Printer) Debug(format string, args ...interface{}) {
	if p.verbose {
		fmt.Printf("🔍 "+format+"\n", args...)
	}
}

// Section prints a section header
func (p *Printer) Section(title string) {
	fmt.Println()
	fmt.Println(repeatChar("=", 60))
	fmt.Printf("  %s\n", title)
	fmt.Println(repeatChar("=", 60))
	fmt.Println()
}

// Celebrate prints a celebration message
func (p *Printer) Celebrate(format string, args ...interface{}) {
	fmt.Println()
	fmt.Println(repeatChar("=", 60))
	fmt.Printf("🎉 "+format+"\n", args...)
	fmt.Println(repeatChar("=", 60))
	fmt.Println()
}

func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}
