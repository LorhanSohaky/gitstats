package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/schollz/progressbar/v3"
)

// collectGitStats collects statistics from a git repository including submodules
func collectGitStats(repoPath string) (map[string]*AuthorStats, int, error) {
	err := os.Chdir(repoPath)
	if err != nil {
		log.Fatal(err)
	}

	// Open the repository
	_, err = git.PlainOpen(repoPath)
	if err != nil {
		return nil, 0, err
	}

	// Create map to store statistics per author
	authorStats := make(map[string]*AuthorStats)
	totalCommits := 0

	// Collect stats from main repository
	mainStats, mainCommits, err := collectRepoStats(repoPath, "Main Repository")
	if err != nil {
		return nil, 0, err
	}

	// Merge main repository stats
	mergeAuthorStats(authorStats, mainStats)
	totalCommits += mainCommits

	// Get submodules and collect their stats
	submodules, err := listSubmoduleFolders(repoPath)
	if err != nil {
		log.Printf("Warning: Could not list submodules: %v", err)
	} else {
		for _, submodule := range submodules {
			submodulePath := repoPath + "/" + submodule

			// Check if submodule is initialized
			if _, err := os.Stat(submodulePath + "/.git"); os.IsNotExist(err) {
				log.Printf("Skipping uninitialized submodule: %s", submodule)
				continue
			}

			log.Printf("Collecting stats from submodule: %s", submodule)
			subStats, subCommits, err := collectRepoStats(submodulePath, submodule)
			if err != nil {
				log.Printf("Warning: Could not collect stats from submodule %s: %v", submodule, err)
				continue
			}

			// Merge submodule stats with main stats
			mergeAuthorStats(authorStats, subStats)
			totalCommits += subCommits
		}
	}

	return authorStats, totalCommits, nil
}

// collectRepoStats collects statistics from a single repository
func collectRepoStats(repoPath string, repoName string) (map[string]*AuthorStats, int, error) {
	// Get total commit count
	cmd := exec.Command("git", "-C", repoPath, "rev-list", "--count", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, 0, err
	}

	outputStr := strings.TrimSpace(string(output))
	totalCommits, err := strconv.Atoi(outputStr)
	if err != nil {
		return nil, 0, err
	}

	if totalCommits == 0 {
		log.Printf("No commits found in %s", repoName)
		return make(map[string]*AuthorStats), 0, nil
	}

	// Create map to store statistics per author
	authorStats := make(map[string]*AuthorStats)

	// Progress bar
	bar := progressbar.Default(int64(totalCommits))
	bar.Describe(fmt.Sprintf("Processing %s", repoName))

	cmd = exec.Command("git", "-C", repoPath, "log", "--shortstat", "--pretty=format:%at %aN", "HEAD")
	output, err = cmd.Output()
	if err != nil {
		return nil, 0, err
	}

	outputStr = strings.TrimSpace(string(output))

	commitsData, err := extractRawGitStats(outputStr)
	if err != nil {
		return nil, 0, err
	}

	for _, commitData := range commitsData {
		if commitData == nil {
			return nil, 0, errors.New("commit data is nil")
		}

		authorName := commitData.author

		if _, exists := authorStats[authorName]; !exists {
			authorStats[authorName] = &AuthorStats{
				DaysOfWeek:        make(map[time.Weekday]int),
				HoursOfDay:        make(map[int]int),
				CommitCount:       0,
				LinesInserted:     0,
				LinesDeleted:      0,
				CountCommitsByDay: make(map[string]int),
			}
		}

		stats := authorStats[authorName]

		// Count commits
		stats.CommitCount++

		timestamp := time.Unix(int64(commitData.timestamp), 0)

		// Assign FirstCommitDate if not set or if this commit is earlier
		if stats.FirstCommitDate.IsZero() || timestamp.Before(stats.FirstCommitDate) {
			stats.FirstCommitDate = timestamp
		}

		if stats.LastCommitDate.IsZero() || timestamp.After(stats.LastCommitDate) {
			stats.LastCommitDate = timestamp
		}

		stats.CountCommitsByDay[timestamp.Format("2006-01-02")]++

		// Count day of week
		stats.DaysOfWeek[timestamp.Weekday()]++

		// Count hour of day
		stats.HoursOfDay[timestamp.Hour()]++

		stats.LinesInserted += commitData.insertions
		stats.LinesDeleted += commitData.deletions

		bar.Add(1)
	}

	bar.Finish()
	return authorStats, totalCommits, nil
}

// mergeAuthorStats merges statistics from one map into another
func mergeAuthorStats(target map[string]*AuthorStats, source map[string]*AuthorStats) {
	for authorName, sourceStats := range source {
		if _, exists := target[authorName]; !exists {
			// Create new entry with initialized maps
			target[authorName] = &AuthorStats{
				DaysOfWeek:        make(map[time.Weekday]int),
				HoursOfDay:        make(map[int]int),
				CommitCount:       0,
				LinesInserted:     0,
				LinesDeleted:      0,
				CountCommitsByDay: make(map[string]int),
			}
		}

		targetStats := target[authorName]

		// Merge commit counts
		targetStats.CommitCount += sourceStats.CommitCount
		targetStats.LinesInserted += sourceStats.LinesInserted
		targetStats.LinesDeleted += sourceStats.LinesDeleted

		// Merge first/last commit dates
		if targetStats.FirstCommitDate.IsZero() || (!sourceStats.FirstCommitDate.IsZero() && sourceStats.FirstCommitDate.Before(targetStats.FirstCommitDate)) {
			targetStats.FirstCommitDate = sourceStats.FirstCommitDate
		}

		if targetStats.LastCommitDate.IsZero() || (!sourceStats.LastCommitDate.IsZero() && sourceStats.LastCommitDate.After(targetStats.LastCommitDate)) {
			targetStats.LastCommitDate = sourceStats.LastCommitDate
		}

		// Merge day-by-day commits
		for date, count := range sourceStats.CountCommitsByDay {
			targetStats.CountCommitsByDay[date] += count
		}

		// Merge days of week
		for day, count := range sourceStats.DaysOfWeek {
			targetStats.DaysOfWeek[day] += count
		}

		// Merge hours of day
		for hour, count := range sourceStats.HoursOfDay {
			targetStats.HoursOfDay[hour] += count
		}
	}
}

type CommitStats struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

type CommitData struct {
	timestamp    int
	author       string
	filesChanged int
	insertions   int
	deletions    int
}

func extractRawGitStats(str string) ([]*CommitData, error) {
	// var filesChanged, insertions, deletions int = 0, 0, 0

	var re = regexp.MustCompile(`(?m)^(\d+)\s+(.+?)\n(\s*(\d+)\s+files?\s+changed(?:,\s+(\d+)\s+insertions?\(\+\))?(?:,\s+(\d+)\s+deletions?\(-\))?)?`)

	matches := re.FindAllStringSubmatch(str, -1)

	var commits []*CommitData = []*CommitData{}

	var filesChanged, insertions, deletions int = 0, 0, 0
	var author string
	for _, match := range matches {
		timestamp, _ := strconv.Atoi(match[1])
		author = match[2]
		filesChanged, _ = strconv.Atoi(match[4])
		insertions, _ = strconv.Atoi(match[5])
		deletions, _ = strconv.Atoi(match[6])

		commits = append(commits, &CommitData{
			timestamp:    timestamp,
			author:       author,
			filesChanged: filesChanged,
			insertions:   insertions,
			deletions:    deletions,
		})
	}

	// reverse the order of commits
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	return commits, nil
}

// listSubmoduleFolders lists all submodule folders in the repository
func listSubmoduleFolders(repoPath string) ([]string, error) {
	var submoduleFolders []string

	// Change to repository directory
	err := os.Chdir(repoPath)
	if err != nil {
		return nil, err
	}

	// Check if .gitmodules file exists
	if _, err := os.Stat(".gitmodules"); os.IsNotExist(err) {
		log.Println("No submodules found in this repository")
		return submoduleFolders, nil
	}

	// Use git command to list submodules
	cmd := exec.Command("git", "submodule", "status")
	output, err := cmd.Output()
	if err != nil {
		// If git submodule command fails, try alternative approach
		return listSubmoduleFoldersFromFile(repoPath)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return submoduleFolders, nil
	}

	// Parse git submodule status output
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse submodule status line format: " <commit_hash> <path> (<branch>)"
		// or "-<commit_hash> <path>" for uninitialized submodules
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			path := parts[1]
			submoduleFolders = append(submoduleFolders, path)
		}
	}

	return submoduleFolders, nil
}

// listSubmoduleFoldersFromFile reads submodule paths from .gitmodules file
func listSubmoduleFoldersFromFile(repoPath string) ([]string, error) {
	var submoduleFolders []string

	// Read .gitmodules file
	gitmodulesPath := repoPath + "/.gitmodules"
	content, err := os.ReadFile(gitmodulesPath)
	if err != nil {
		return nil, err
	}

	// Parse .gitmodules content
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "path = ") {
			path := strings.TrimPrefix(line, "path = ")
			path = strings.TrimSpace(path)
			if path != "" {
				submoduleFolders = append(submoduleFolders, path)
			}
		}
	}

	return submoduleFolders, nil
}

// printSubmoduleInfo prints information about submodules
func printSubmoduleInfo(repoPath string) {
	submodules, err := listSubmoduleFolders(repoPath)
	if err != nil {
		log.Printf("Error listing submodules: %v\n", err)
		return
	}

	if len(submodules) == 0 {
		log.Println("No submodules found in this repository")
		return
	}

	log.Printf("Found %d submodule(s):\n", len(submodules))
	log.Println("================================")

	for i, submodule := range submodules {
		log.Printf("%d. %s\n", i+1, submodule)

		// Check if submodule folder exists
		submodulePath := repoPath + "/" + submodule
		if _, err := os.Stat(submodulePath); os.IsNotExist(err) {
			log.Printf("   Status: Not initialized\n")
		} else {
			log.Printf("   Status: Present\n")

			// Try to get submodule commit info
			cmd := exec.Command("git", "-C", repoPath, "ls-tree", "HEAD", submodule)
			if output, err := cmd.Output(); err == nil {
				outputStr := strings.TrimSpace(string(output))
				if outputStr != "" {
					parts := strings.Fields(outputStr)
					if len(parts) >= 3 {
						commitHash := parts[2]
						log.Printf("   Commit: %s\n", commitHash[:8])
					}
				}
			}
		}
		log.Println()
	}
}

// getSubmoduleStats collects statistics including submodule information
func getSubmoduleStats(repoPath string) map[string]interface{} {
	stats := make(map[string]interface{})

	submodules, err := listSubmoduleFolders(repoPath)
	if err != nil {
		stats["error"] = err.Error()
		return stats
	}

	stats["count"] = len(submodules)
	stats["paths"] = submodules

	// Check status of each submodule
	var initialized, uninitialized []string
	for _, submodule := range submodules {
		submodulePath := repoPath + "/" + submodule
		if _, err := os.Stat(submodulePath); os.IsNotExist(err) {
			uninitialized = append(uninitialized, submodule)
		} else {
			initialized = append(initialized, submodule)
		}
	}

	stats["initialized"] = initialized
	stats["uninitialized"] = uninitialized
	stats["initialized_count"] = len(initialized)
	stats["uninitialized_count"] = len(uninitialized)

	return stats
}
