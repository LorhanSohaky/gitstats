package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// generateHTMLReportWithFile generates an HTML report with git statistics to a specified file
func generateHTMLReportWithFile(authorStats map[string]*AuthorStats, totalCommits int, outputFile string) {
	// Calculate global statistics
	globalHours := make(map[int]int)
	globalDays := make(map[time.Weekday]int)

	for _, stats := range authorStats {
		for hour, count := range stats.HoursOfDay {
			globalHours[hour] += count
		}
		for day, count := range stats.DaysOfWeek {
			globalDays[day] += count
		}
	}

	// Find max values for heatmap scaling
	maxHourValue := 0
	for _, count := range globalHours {
		if count > maxHourValue {
			maxHourValue = count
		}
	}

	maxDayValue := 0
	for _, count := range globalDays {
		if count > maxDayValue {
			maxDayValue = count
		}
	}

	html := generateHTMLContent(authorStats, totalCommits, globalHours, globalDays, maxHourValue, maxDayValue)

	// Save HTML file with custom name
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.WriteString(html)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("HTML report generated: %s\n", outputFile)
}

// generateHTMLContent creates the complete HTML content
func generateHTMLContent(authorStats map[string]*AuthorStats, totalCommits int, globalHours map[int]int, globalDays map[time.Weekday]int, maxHourValue, maxDayValue int) string {
	html := generateHTMLHeader(totalCommits)
	html += generateHourlySection(globalHours, totalCommits, maxHourValue)
	html += generateDailySection(globalDays, totalCommits, maxDayValue)
	html += generateAuthorsSection(authorStats, totalCommits)
	html += generateAuthorOfMonthSection(authorStats) // Add this line
	html += generateLinesChartSection()
	html += generateJavaScript(globalHours, globalDays, authorStats)
	html += generateHTMLFooter()
	return html
}

// generateHTMLHeader generates the HTML header and initial structure
func generateHTMLHeader(totalCommits int) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Git Statistics Report</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        .chart-container { 
            height: 400px; 
            margin: 30px 0; 
        }
        .table-responsive {
            margin: 20px 0;
        }
        .stats-card {
            box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
            border: 1px solid rgba(0, 0, 0, 0.125);
        }
        .heatmap-cell {
            color: white;
            font-weight: bold;
            text-shadow: 1px 1px 1px rgba(0,0,0,0.5);
        }
    </style>
</head>
<body>
    <div class="container-fluid py-4">
        <div class="row">
            <div class="col-12">
                <div class="text-center mb-5">
                    <h1 class="display-4 text-primary">Git Statistics Report</h1>
                    <p class="lead">Total commits analyzed: <span class="badge bg-primary fs-6">` + fmt.Sprintf("%d", totalCommits) + `</span></p>
                </div>
            </div>
        </div>`
}

// generateHourlySection generates the hourly statistics section
func generateHourlySection(globalHours map[int]int, totalCommits, maxHourValue int) string {
	html := `
        <div class="row">
            <div class="col-12">
                <div class="card stats-card mb-5">
                    <div class="card-header bg-primary text-white">
                        <h2 class="card-title mb-0"><i class="bi bi-clock"></i> Commits by Hour of Day</h2>
                    </div>
                    <div class="card-body">
                        <div class="table-responsive">
                            <table class="table table-bordered">
                                <thead class="table-dark">
                                    <tr>
                                        <th scope="col">Hour</th>`

	// Hour headers (horizontal)
	for hour := 0; hour < 24; hour++ {
		html += fmt.Sprintf(`<th scope="col" class="text-center">%02d</th>`, hour)
	}
	html += `</tr>
                                </thead>
                                <tbody>
                                    <tr>
                                        <td><strong>Commits</strong></td>`

	// Commit data by hour with heatmap
	for hour := 0; hour < 24; hour++ {
		count := globalHours[hour]
		intensity := float64(count) / float64(maxHourValue)
		red := int(255 * intensity)
		backgroundColor := fmt.Sprintf("rgba(%d, 0, 0, %.2f)", red, 0.3+intensity*0.7)
		html += fmt.Sprintf(`<td class="text-center heatmap-cell" style="background-color: %s;">%d</td>`, backgroundColor, count)
	}

	html += `</tr>
                                    <tr>
                                        <td><strong>%</strong></td>`

	// Percentages by hour with heatmap
	for hour := 0; hour < 24; hour++ {
		count := globalHours[hour]
		percentage := float64(count) / float64(totalCommits) * 100
		intensity := float64(count) / float64(maxHourValue)
		red := int(255 * intensity)
		backgroundColor := fmt.Sprintf("rgba(%d, 0, 0, %.2f)", red, 0.3+intensity*0.7)
		html += fmt.Sprintf(`<td class="text-center heatmap-cell" style="background-color: %s;">%.2f</td>`, backgroundColor, percentage)
	}

	html += `</tr>
                                </tbody>
                            </table>
                        </div>
                        <div class="chart-container">
                            <canvas id="hoursChart"></canvas>
                        </div>
                    </div>
                </div>
            </div>
        </div>`

	return html
}

// generateDailySection generates the daily statistics section
func generateDailySection(globalDays map[time.Weekday]int, totalCommits, maxDayValue int) string {
	html := `
        <div class="row">
            <div class="col-12">
                <div class="card stats-card mb-5">
                    <div class="card-header bg-success text-white">
                        <h2 class="card-title mb-0"><i class="bi bi-calendar-week"></i> Commits by Day of Week</h2>
                    </div>
                    <div class="card-body">
                        <div class="table-responsive">
                            <table class="table table-bordered">
                                <thead class="table-dark">
                                    <tr>
                                        <th scope="col">Day</th>`

	// Day headers (horizontal)
	dayNames := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for _, dayName := range dayNames {
		html += fmt.Sprintf(`<th scope="col" class="text-center">%s</th>`, dayName)
	}
	html += `</tr>
                                </thead>
                                <tbody>
                                    <tr>
                                        <td><strong>Commits</strong></td>`

	// Commit data by day of week with heatmap
	days := []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}
	for _, day := range days {
		count := globalDays[day]
		intensity := float64(count) / float64(maxDayValue)
		red := int(255 * intensity)
		backgroundColor := fmt.Sprintf("rgba(%d, 0, 0, %.2f)", red, 0.3+intensity*0.7)
		html += fmt.Sprintf(`<td class="text-center heatmap-cell" style="background-color: %s;">%d</td>`, backgroundColor, count)
	}
	html += `</tr>
                                    <tr>
                                        <td><strong>%</strong></td>`

	// Percentages by day of week with heatmap
	for _, day := range days {
		count := globalDays[day]
		percentage := float64(count) / float64(totalCommits) * 100
		intensity := float64(count) / float64(maxDayValue)
		red := int(255 * intensity)
		backgroundColor := fmt.Sprintf("rgba(%d, 0, 0, %.2f)", red, 0.3+intensity*0.7)
		html += fmt.Sprintf(`<td class="text-center heatmap-cell" style="background-color: %s;">%.2f</td>`, backgroundColor, percentage)
	}

	html += `</tr>
                                </tbody>
                            </table>
                        </div>
                        <div class="chart-container">
                            <canvas id="daysChart"></canvas>
                        </div>
                    </div>
                </div>
            </div>
        </div>`

	return html
}

// generateAuthorsSection generates the authors statistics section
func generateAuthorsSection(authorStats map[string]*AuthorStats, totalCommits int) string {
	html := `
    <div class="row">
        <div class="col-12">
            <div class="card stats-card mb-5">
                <div class="card-header bg-info text-white">
                    <h2 class="card-title mb-0"><i class="bi bi-people"></i> List of Authors</h2>
                </div>
                <div class="card-body">
                    <div class="table-responsive">
                        <table class="table table-striped table-hover">
                            <thead class="table-dark">
                                <tr>
                                    <th scope="col">Author</th>
                                    <th scope="col" class="text-center">Commits</th>
                                    <th scope="col" class="text-center">+ lines</th>
                                    <th scope="col" class="text-center">- lines</th>
                                    <th scope="col" class="text-center">First Commit</th>
                                    <th scope="col" class="text-center">Last Commit</th>
                                    <th scope="col" class="text-center">Active Days</th>
                                </tr>
                            </thead>
                            <tbody>`

	// Prepare authors data
	var authors []AuthorInfo
	for name, stats := range authorStats {
		percentage := float64(stats.CommitCount) / float64(totalCommits) * 100
		authors = append(authors, AuthorInfo{
			Name:            name,
			CommitCount:     stats.CommitCount,
			LinesInserted:   stats.LinesInserted,
			LinesDeleted:    stats.LinesDeleted,
			Percentage:      percentage,
			FirstCommitDate: stats.FirstCommitDate,
			LastCommitDate:  stats.LastCommitDate,
		})
	}

	// Sort authors by commit count (descending)
	for i := 0; i < len(authors); i++ {
		for j := i + 1; j < len(authors); j++ {
			if authors[j].CommitCount > authors[i].CommitCount {
				authors[i], authors[j] = authors[j], authors[i]
			}
		}
	}

	for _, author := range authors {
		firstCommitStr := ""
		lastCommitStr := ""

		firstCommitStr = author.FirstCommitDate.Format("2006-01-02")
		lastCommitStr = author.LastCommitDate.Format("2006-01-02")

		activeDays := len(authorStats[author.Name].CountCommitsByDay)

		html += fmt.Sprintf(`
                                <tr>
                                    <td><strong>%s</strong></td>
                                    <td class="text-center">%d (%.2f%%)</td>
                                    <td class="text-center text-success">%d</td>
                                    <td class="text-center text-danger">%d</td>
                                    <td class="text-center">%s</td>
                                    <td class="text-center">%s</td>
                                    <td class="text-center">%d</td>
                                </tr>`, author.Name, author.CommitCount, author.Percentage, author.LinesInserted, author.LinesDeleted, firstCommitStr, lastCommitStr, activeDays)
	}

	html += `
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    </div>`

	return html
}

// generateAuthorOfMonthSection generates the author of month statistics section
func generateAuthorOfMonthSection(authorStats map[string]*AuthorStats) string {
	// Collect all monthly data
	monthlyStats := make(map[string]map[string]int) // [year-month][author] = commit_count

	for authorName, stats := range authorStats {

		if stats.CountCommitsByDay == nil {
			continue
		}

		for date, commits := range stats.CountCommitsByDay {
			dateTime, _ := time.Parse("2006-01-02", date)
			yearMonth := dateTime.Format("2006-01")
			if monthlyStats[yearMonth] == nil {
				monthlyStats[yearMonth] = make(map[string]int)
			}
			monthlyStats[yearMonth][authorName] = commits
		}

	}

	// Sort months chronologically
	var sortedMonths []string
	for month := range monthlyStats {
		sortedMonths = append(sortedMonths, month)
	}
	// Simple sort for YYYY-MM format
	for i := 0; i < len(sortedMonths); i++ {
		for j := i + 1; j < len(sortedMonths); j++ {
			if sortedMonths[i] < sortedMonths[j] {
				sortedMonths[i], sortedMonths[j] = sortedMonths[j], sortedMonths[i]
			}
		}
	}

	html := `
    <div class="row">
        <div class="col-12">
            <div class="card stats-card mb-5">
                <div class="card-header bg-dark text-white">
                    <h2 class="card-title mb-0"><i class="bi bi-calendar-month"></i> Author of Month</h2>
                </div>
                <div class="card-body">
                    <div class="table-responsive">
                        <table class="table table-striped table-hover">
                            <thead class="table-dark">
                                <tr>
                                    <th scope="col">Year-Month</th>
                                    <th scope="col">Top Author</th>
                                    <th scope="col" class="text-center">Commits (%)</th>
                                    <th scope="col">Next Top 5</th>
                                </tr>
                            </thead>
                            <tbody>`

	for _, month := range sortedMonths {
		authors := monthlyStats[month]

		// Convert to slice and sort by commit count
		type authorCommit struct {
			name    string
			commits int
		}

		var authorList []authorCommit
		totalMonthCommits := 0
		for name, commits := range authors {
			authorList = append(authorList, authorCommit{name: name, commits: commits})
			totalMonthCommits += commits
		}

		// Sort by commit count (descending)
		for i := 0; i < len(authorList); i++ {
			for j := i + 1; j < len(authorList); j++ {
				if authorList[j].commits > authorList[i].commits {
					authorList[i], authorList[j] = authorList[j], authorList[i]
				}
			}
		}

		if len(authorList) > 0 {
			topAuthor := authorList[0]
			percentage := float64(topAuthor.commits) / float64(totalMonthCommits) * 100

			// Get next top 5
			var nextTop5 []string
			for i := 1; i < len(authorList) && i <= 5; i++ {
				nextPercentage := float64(authorList[i].commits) / float64(totalMonthCommits) * 100
				nextTop5 = append(nextTop5, fmt.Sprintf("%s (%d, %.1f%%)",
					authorList[i].name, authorList[i].commits, nextPercentage))
			}

			nextTop5Str := ""
			if len(nextTop5) > 0 {
				nextTop5Str = fmt.Sprintf("%s", nextTop5[0])
				for i := 1; i < len(nextTop5); i++ {
					nextTop5Str += fmt.Sprintf(", %s", nextTop5[i])
				}
			}

			html += fmt.Sprintf(`
                                <tr>
                                    <td><strong>%s</strong></td>
                                    <td><strong>%s</strong></td>
                                    <td class="text-center"><span class="badge bg-primary">%d (%.1f%%)</span></td>
                                    <td><small>%s</small></td>
                                </tr>`, month, topAuthor.name, topAuthor.commits, percentage, nextTop5Str)
		}
	}

	html += `
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    </div>`

	return html
}

// generateLinesChartSection generates the lines of code chart section
func generateLinesChartSection() string {
	return `
    <div class="row">
        <div class="col-12">
            <div class="card stats-card mb-5">
                <div class="card-header bg-warning text-dark">
                    <h2 class="card-title mb-0"><i class="bi bi-graph-up"></i> Cumulated Added Lines of Code per Author</h2>
                </div>
                <div class="card-body">
                    <div class="chart-container">
                        <canvas id="linesChart"></canvas>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>`
}

// generateJavaScript generates the JavaScript code for charts
func generateJavaScript(globalHours map[int]int, globalDays map[time.Weekday]int, authorStats map[string]*AuthorStats) string {
	// Prepare authors data for sorting
	var authors []AuthorInfo
	for name, stats := range authorStats {
		authors = append(authors, AuthorInfo{
			Name:          name,
			CommitCount:   stats.CommitCount,
			LinesInserted: stats.LinesInserted,
			LinesDeleted:  stats.LinesDeleted,
		})
	}

	// Sort authors by commit count (descending)
	for i := 0; i < len(authors); i++ {
		for j := i + 1; j < len(authors); j++ {
			if authors[j].CommitCount > authors[i].CommitCount {
				authors[i], authors[j] = authors[j], authors[i]
			}
		}
	}

	html := `
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        // Hours chart
        const hoursCtx = document.getElementById('hoursChart').getContext('2d');
        const hoursChart = new Chart(hoursCtx, {
            type: 'bar',
            data: {
                labels: [`

	// Hour labels
	for hour := 0; hour < 24; hour++ {
		html += fmt.Sprintf(`'%02d:00'`, hour)
		if hour < 23 {
			html += ","
		}
	}

	html += `],
                datasets: [{
                    label: 'Commits by Hour',
                    data: [`

	// Hour data
	for hour := 0; hour < 24; hour++ {
		count := globalHours[hour]
		html += fmt.Sprintf(`%d`, count)
		if hour < 23 {
			html += ","
		}
	}

	html += `],
                    backgroundColor: 'rgba(13, 110, 253, 0.8)',
                    borderColor: 'rgba(13, 110, 253, 1)',
                    borderWidth: 2,
                    borderRadius: 4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'top'
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.1)'
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        }
                    }
                }
            }
        });

        // Days chart
        const daysCtx = document.getElementById('daysChart').getContext('2d');
        const daysChart = new Chart(daysCtx, {
            type: 'bar',
            data: {
                labels: ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'],
                datasets: [{
                    label: 'Commits by Day of Week',
                    data: [`

	// Day data
	days := []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}
	for i, day := range days {
		count := globalDays[day]
		html += fmt.Sprintf(`%d`, count)
		if i < len(days)-1 {
			html += ","
		}
	}

	html += `],
                    backgroundColor: 'rgba(25, 135, 84, 0.8)',
                    borderColor: 'rgba(25, 135, 84, 1)',
                    borderWidth: 2,
                    borderRadius: 4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'top'
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.1)'
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        }
                    }
                }
            }
        });

        // Lines chart
        const linesCtx = document.getElementById('linesChart').getContext('2d');
        const linesChart = new Chart(linesCtx, {
            type: 'bar',
            data: {
                labels: [`

	// Author names for chart
	for i, author := range authors {
		html += fmt.Sprintf(`'%s'`, author.Name)
		if i < len(authors)-1 {
			html += ","
		}
	}

	html += `],
                datasets: [{
                    label: 'Lines Added',
                    data: [`

	// Lines added data
	for i, author := range authors {
		html += fmt.Sprintf(`%d`, author.LinesInserted)
		if i < len(authors)-1 {
			html += ","
		}
	}

	html += `],
                    backgroundColor: 'rgba(255, 193, 7, 0.8)',
                    borderColor: 'rgba(255, 193, 7, 1)',
                    borderWidth: 2,
                    borderRadius: 4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'top'
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.1)'
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        }
                    }
                }
            }
        });
    </script>`

	return html
}

// generateHTMLFooter generates the HTML footer
func generateHTMLFooter() string {
	return `
</body>
</html>`
}
