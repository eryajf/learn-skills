package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RunInteractive starts an interactive chat session
func RunInteractive(assistant *Assistant) error {
	fmt.Println("CNB Assistant - Interactive Mode")
	fmt.Println("Type 'exit' to quit, 'clear' to reset conversation, 'help' for help")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("CNB Assistant> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// Handle special commands
		switch input {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return nil
		case "clear":
			assistant.Reset()
			fmt.Println("Conversation cleared.")
			continue
		case "help":
			printHelp()
			continue
		case "":
			continue
		}

		// Process user message with streaming
		fmt.Println()
		_, err := assistant.ProcessMessageStream(input, func(chunk string) error {
			fmt.Print(chunk)
			return nil
		})
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}

		fmt.Println()
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

func printHelp() {
	fmt.Println(`
Available Commands:
  exit, quit  - Exit the assistant
  clear       - Clear conversation history
  help        - Show this help message

You can ask questions like:
  - "List my repositories"
  - "How do I configure webhooks in CNB?"
  - "Trigger build for demo-app main branch"
  - "Show me the status of build #123"
`)
}
