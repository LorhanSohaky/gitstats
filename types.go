package main

import "time"

// AuthorStats stores statistics for each author
type AuthorStats struct {
	CommitCount       int
	LinesInserted     int
	LinesDeleted      int
	Emails            map[string]bool
	HoursOfDay        map[int]int
	DaysOfWeek        map[time.Weekday]int
	FirstCommitDate   time.Time
	LastCommitDate    time.Time
	CountCommitsByDay map[string]int
}

// AuthorInfo represents author information for reporting
type AuthorInfo struct {
	Name            string
	CommitCount     int
	LinesInserted   int
	LinesDeleted    int
	Percentage      float64
	FirstCommitDate time.Time
	LastCommitDate  time.Time
}
