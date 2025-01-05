package shuffle

import (
	"bufio"
	"eclipse/internal/logger"
	"eclipse/pkg/services/file"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func ShuffleFiles(evmFilePath, eclipseFilePath string) error {
	lines1, err := file.ReadLines(evmFilePath)
	if err != nil {
		return fmt.Errorf("error reading first file: %w", err)
	}

	lines2, err := file.ReadLines(eclipseFilePath)
	if err != nil {
		return fmt.Errorf("error reading second file: %w", err)
	}

	if len(lines1) != len(lines2) {
		return fmt.Errorf("files have different number of lines: %d vs %d", len(lines1), len(lines2))
	}

	indices := make([]int, len(lines1))
	for i := range indices {
		indices[i] = i
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	shuffled1 := make([]string, len(lines1))
	shuffled2 := make([]string, len(lines2))
	for i, idx := range indices {
		shuffled1[i] = lines1[idx]
		shuffled2[i] = lines2[idx]
	}

	if err := writeLines(evmFilePath, shuffled1); err != nil {
		return fmt.Errorf("error writing first file: %w", err)
	}
	if err := writeLines(eclipseFilePath, shuffled2); err != nil {
		return fmt.Errorf("error writing second file: %w", err)
	}

	logger.Success("Кошельки успешно перемешаны")
	fmt.Println()
	return nil
}

func writeLines(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}
