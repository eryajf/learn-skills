package cli

import (
	"fmt"
	"strings"
)

// RunOneShot processes a single query and exits
func RunOneShot(assistant *Assistant, args []string) error {
	// Join all arguments as the query
	query := strings.Join(args, " ")

	if query == "" {
		return fmt.Errorf("no query provided")
	}

	// Process the query
	response, err := assistant.ProcessMessage(query)
	if err != nil {
		return fmt.Errorf("failed to process query: %w", err)
	}

	// Print response
	fmt.Println(response)

	return nil
}
