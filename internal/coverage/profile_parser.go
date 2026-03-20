package coverage

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// CoverageProfile represents parsed coverage data
type CoverageProfile struct {
	FileCoverage map[string]*FileCoverage
}

// FileCoverage represents coverage for a single file
type FileCoverage struct {
	FilePath string
	Blocks   []CoverageBlock
}

// CoverageBlock represents a single coverage block
type CoverageBlock struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmt   int
	Count     int
}

// ParseCoverageProfile parses a Go coverage profile file
func ParseCoverageProfile(profilePath string) (*CoverageProfile, error) {
	file, err := os.Open(profilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	profile := &CoverageProfile{
		FileCoverage: make(map[string]*FileCoverage),
	}

	scanner := bufio.NewScanner(file)

	// Skip the first line (mode: set/count/atomic)
	if scanner.Scan() {
		// mode line
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		block, filePath, err := parseCoverageLine(line)
		if err != nil {
			continue // Skip malformed lines
		}

		if _, exists := profile.FileCoverage[filePath]; !exists {
			profile.FileCoverage[filePath] = &FileCoverage{
				FilePath: filePath,
				Blocks:   []CoverageBlock{},
			}
		}

		profile.FileCoverage[filePath].Blocks = append(profile.FileCoverage[filePath].Blocks, block)
	}

	return profile, scanner.Err()
}

// parseCoverageLine parses a single line from the coverage profile
// Format: file.go:startLine.startCol,endLine.endCol numStmt count
func parseCoverageLine(line string) (CoverageBlock, string, error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return CoverageBlock{}, "", os.ErrInvalid
	}

	// Split file path and position
	fileAndPos := parts[0]
	colonIdx := strings.LastIndex(fileAndPos, ":")
	if colonIdx == -1 {
		return CoverageBlock{}, "", os.ErrInvalid
	}

	filePath := fileAndPos[:colonIdx]
	position := fileAndPos[colonIdx+1:]

	// Parse position: startLine.startCol,endLine.endCol
	positions := strings.Split(position, ",")
	if len(positions) != 2 {
		return CoverageBlock{}, "", os.ErrInvalid
	}

	start := strings.Split(positions[0], ".")
	end := strings.Split(positions[1], ".")
	if len(start) != 2 || len(end) != 2 {
		return CoverageBlock{}, "", os.ErrInvalid
	}

	startLine, _ := strconv.Atoi(start[0])
	startCol, _ := strconv.Atoi(start[1])
	endLine, _ := strconv.Atoi(end[0])
	endCol, _ := strconv.Atoi(end[1])
	numStmt, _ := strconv.Atoi(parts[1])
	count, _ := strconv.Atoi(parts[2])

	block := CoverageBlock{
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		NumStmt:   numStmt,
		Count:     count,
	}

	return block, filePath, nil
}

// GetLineCoverage determines coverage for each line in a file
func (fc *FileCoverage) GetLineCoverage() map[int]int {
	lineCoverage := make(map[int]int)

	for _, block := range fc.Blocks {
		for line := block.StartLine; line <= block.EndLine; line++ {
			lineCoverage[line] = block.Count
		}
	}

	return lineCoverage
}

// CalculateFunctionCoverage calculates coverage percentage for a function
func CalculateFunctionCoverage(startLine, endLine int, lineCoverage map[int]int) float64 {
	if endLine < startLine {
		return 0.0
	}

	totalLines := 0
	coveredLines := 0

	for line := startLine; line <= endLine; line++ {
		if count, exists := lineCoverage[line]; exists {
			totalLines++
			if count > 0 {
				coveredLines++
			}
		}
	}

	if totalLines == 0 {
		return 0.0
	}

	return float64(coveredLines) / float64(totalLines) * 100.0
}
