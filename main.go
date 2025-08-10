package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <repository-path>\n", os.Args[0])
		os.Exit(1)
	}

	repoPath := os.Args[1]

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Analyzing repository: %s\n", repoPath)

	// Collect git statistics
	authorStats, totalCommits, err := collectGitStats(repoPath)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Chdir(currentDir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total commits: %d\n", totalCommits)

	// Print statistics to console
	// printStats(authorStats)

	// Generate HTML report
	generateHTMLReport(authorStats, totalCommits)
}
