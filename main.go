package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

import "github.com/spf13/cobra"

var (
	repoPath    string
	outputFile  string
	showHelp    bool
	showVersion bool
	version     = "1.0.0"
)

var rootCmd = &cobra.Command{
	Use:   "gitstats [flags] <repository-path>",
	Short: "GitStats is a tool for analyzing git repositories",
	Long: `GitStats is a tool for analyzing git repositories and generating
statistics about commits, authors, and more.

Examples:
  gitstats --output report.html /path/to/repo
  gitstats -o custom_report.html /path/to/repo
  gitstats --version
  gitstats --help`,
	Args: func(cmd *cobra.Command, args []string) error {
		// If version flag is set, don't require repository path
		if showVersion || showHelp {
			return nil
		}
		cmd.MarkFlagRequired("output")
		// Otherwise, require exactly one argument
		return cobra.ExactArgs(1)(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Printf("GitStats version %s\n", version)
			return
		}

		repoPath = args[0]

		// Validate repository path
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			log.Fatalf("Error: '%s' is not a valid git repository", repoPath)
		}

		log.SetFlags(0)

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

		// Generate HTML report
		generateHTMLReportWithFile(authorStats, totalCommits, outputFile)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for the HTML report (required)")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version information")
	rootCmd.Flags().BoolVarP(&showHelp, "help", "h", false, "Show help information")

	// We'll handle the required validation in the PreRunE function instead of here
	// to allow version flag to work without requiring output
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
