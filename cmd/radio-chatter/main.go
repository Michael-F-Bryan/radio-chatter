/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"
)

func main() {
	cmd := rootCmd()

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
