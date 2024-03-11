package vjudge

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var VjStatusMap = map[string]string{
	"Accepted":                  "Accepted",
	"Partial Accepted":          "Partial Accepted",
	"Presentation Error":        "Presentation Error",
	"Wrong Answer":              "Wrong Answer",
	"Incorrect":                 "Wrong Answer",
	"Time Limit Exceed":         "Time Limit Exceed",
	"Terminated due to timeout": "Time Limit Exceed",
	"Memory Limit Exceed":       "Memory Limit Exceed",
	"Output Limit Exceed":       "Output Limit Exceed",
	"Runtime Error":             "Runtime Error",
	"Segmentation Fault":        "Runtime Error",
	"Compile Error":             "Compile Error",
	"Compilation Error":         "Compile Error",
	"Remote OJ Unavailable":     "Judge Error",
	"Judge Failed":              "Judge Error",
	"Unknown Error":             "Judge Error",
	"Submit Failed":             "Judge Error",
}

func getMappedStatus(status string) string {
	for k, v := range VjStatusMap {
		if len(status) < len(k) {
			continue
		}
		if strings.ToLower(status)[0:len(k)] == strings.ToLower(k) {
			return v
		}
	}
	return "Judge Error"
}

func parseScore(info string) float64 {
	re := regexp.MustCompile(`(\d+\.\d+) / (\d+\.\d+)`)
	matches := re.FindStringSubmatch(info)
	if len(matches) != 3 {
		return 0
	}
	score, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}
	total, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return 0
	}
	normalizedScore := min(1, max(0, score/total)) * 100
	return normalizedScore
}

func generateVjMd(solution VjSolution) string {
	// Convert the Unix timestamp (milliseconds) to a readable datetime string in local timezone
	submitTime := time.UnixMilli(solution.SubmitTime).Format("1926-08-17 01:02:03")

	// Create the markdown table, integrating the links directly into it
	markdown := "| Status | Time | Memory | Length | Lang | Submitted | Vjudge | Origin |\n"
	markdown += "|--------|------|--------|--------|------|-----------|--------|--------|\n"
	markdown += "| `" + solution.Status + "` "
	markdown += "| `" + fmt.Sprint(solution.Runtime) + " ms` "
	markdown += "| `" + fmt.Sprint(solution.Memory) + " KB` "
	markdown += "| `" + fmt.Sprint(solution.Length) + " bytes` "
	markdown += "| `" + solution.Language + "` "
	markdown += "| `" + submitTime + "` "
	markdown += "| [Link](https://vjudge.net/solution/" + fmt.Sprint(solution.RunId) + ") "
	markdown += "| [Link](https://vjudge.net/solution/" + fmt.Sprint(solution.RunId) + "/origin) |\n\n"

	// Conditionally adding AdditionalInfo if it's not empty
	if solution.AdditionalInfo != "" {
		markdown += "```\n" + solution.AdditionalInfo + "\n```\n\n"
	}

	// Adding the code block with syntax highlighting
	markdown += "```" + solution.PrismClass + "\n"
	markdown += solution.Code + "\n"
	markdown += "```\n"

	return markdown
}
