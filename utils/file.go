package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// SanitizeFilename removes invalid characters for a filename
func SanitizeFilename(name string) string {
	// Remove restricted characters: / \ : * ? " < > |
	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := re.ReplaceAllString(name, "")
	
	// Replace multiple spaces with a single space
	reSpaces := regexp.MustCompile(`\s+`)
	sanitized = reSpaces.ReplaceAllString(sanitized, " ")
	
	return strings.TrimSpace(sanitized)
}

// PromptOverwrite checks if file exists and asks user for overwrite permission
func PromptOverwrite(filepath string) (bool, error) {
	if _, err := os.Stat(filepath); err == nil {
		fmt.Printf("File '%s' already exists. Overwrite? (y/N): ", filepath)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true, nil
		}
		return false, nil
	}
	
	return true, nil // File doesn't exist, proceed
}

// EnsureDir checks if a directory exists, and creates it if it doesn't
func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}
