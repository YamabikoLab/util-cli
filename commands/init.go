package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
)

func RunInit(_ *cobra.Command, _ []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".util-cli")
	err = os.MkdirAll(configDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	srcFile, err := os.Open("config.yml")
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()

	destFilePath := filepath.Join(configDir, "config.yml")
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("copying file: %w", err)
	}

	// Close files before trying to remove them
	srcFile.Close()
	destFile.Close()

	err = os.Remove("config.yml")
	if err != nil {
		return fmt.Errorf("removing original file: %w", err)
	}

	fmt.Printf("Config file has been moved to: %s\n", destFilePath)
	fmt.Println("Please edit this file as per your requirements.")

	return nil
}
