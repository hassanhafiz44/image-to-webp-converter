package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"

	"github.com/chai2010/webp"
)

var supportedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".tiff": true,
	".tif":  true,
}

type ConversionResult struct {
	Input        string
	Output       string
	Success      bool
	Message      string
	OriginalSize int64
	NewSize      int64
	Savings      float64
}

func main() {
	// Ensure image decoders are registered
	image.RegisterFormat("jpeg", "\xff\xd8", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "\x89PNG\r\n\x1a\n", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "GIF8", gif.Decode, gif.DecodeConfig)

	quality := flag.Int("q", 80, "WebP quality (1-100)")
	inputDir := flag.String("i", "/app/images", "Input directory")
	outputDir := flag.String("o", "/app/output", "Output directory")
	workers := flag.Int("w", 0, "Number of parallel workers (default: CPU cores)")
	flag.Parse()

	if *quality < 1 {
		*quality = 1
	} else if *quality > 100 {
		*quality = 100
	}

	if *workers <= 0 {
		*workers = runtime.NumCPU()
	}

	start := time.Now()
	startTime := start.Format("2006-01-02 15:04:05 MST")

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║     Image to WebP Converter (Go)           ║")
	fmt.Println("╚════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Started at: %s\n", startTime)
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Printf("  • Input directory:  %s\n", *inputDir)
	fmt.Printf("  • Output directory: %s\n", *outputDir)
	fmt.Printf("  • Quality:          %d%%\n", *quality)
	fmt.Printf("  • Workers:          %d\n", *workers)
	fmt.Println()

	// Validate input directory
	info, err := os.Stat(*inputDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: Input directory does not exist: %s\n", *inputDir)
		os.Exit(1)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Scan for image files
	allFiles := getImageFiles(*inputDir)
	totalFound := len(allFiles)

	if totalFound == 0 {
		fmt.Printf("⚠ No images found in %s\n", *inputDir)
		fmt.Println("  Supported formats: jpg, jpeg, png, gif, bmp, tiff, tif")
		return
	}

	fmt.Printf("Found %d image(s)\n", totalFound)

	// Filter already converted
	files, skipped := filterAlreadyConverted(allFiles, *inputDir, *outputDir)
	totalFiles := len(files)

	if skipped > 0 {
		fmt.Printf("  ⏭ Skipped %d file(s) (already converted)\n", skipped)
	}

	if totalFiles == 0 {
		fmt.Println("  All files already converted. Nothing to do.")
		return
	}

	fmt.Printf("  Converting %d file(s)\n", totalFiles)
	fmt.Println(strings.Repeat("-", 50))

	// Parallel conversion
	var counter atomic.Int64
	var totalOriginal atomic.Int64
	var totalNew atomic.Int64
	var successful atomic.Int64
	var failed atomic.Int64

	sem := make(chan struct{}, *workers)
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		sem <- struct{}{}
		go func(f string) {
			defer wg.Done()
			defer func() { <-sem }()

			result := convertImage(f, *inputDir, *outputDir, float32(*quality))
			current := counter.Add(1)
			filename := filepath.Base(result.Input)

			if result.Success {
				successful.Add(1)
				totalOriginal.Add(result.OriginalSize)
				totalNew.Add(result.NewSize)
				fmt.Printf("[%d/%d] ✓ %s: %s → %s (%.2f%% saved)\n",
					current, totalFiles, filename,
					formatBytes(result.OriginalSize),
					formatBytes(result.NewSize),
					result.Savings)
			} else {
				failed.Add(1)
				fmt.Printf("[%d/%d] ✗ %s: %s\n",
					current, totalFiles, filename, result.Message)
			}
		}(file)
	}
	wg.Wait()

	successCount := successful.Load()
	failCount := failed.Load()
	origTotal := totalOriginal.Load()
	newTotal := totalNew.Load()

	endTime := time.Now().Format("2006-01-02 15:04:05 MST")
	elapsed := time.Since(start)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("CONVERSION SUMMARY")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("  • Total files:      %d\n", totalFiles)
	fmt.Printf("  • Successful:       %d\n", successCount)
	fmt.Printf("  • Failed:           %d\n", failCount)

	if successCount > 0 {
		var totalSavings float64
		if origTotal > 0 {
			totalSavings = float64(int((1.0-float64(newTotal)/float64(origTotal))*10000)) / 100.0
		}
		fmt.Printf("  • Original size:    %s\n", formatBytes(origTotal))
		fmt.Printf("  • New size:         %s\n", formatBytes(newTotal))
		fmt.Printf("  • Total savings:    %.2f%%\n", totalSavings)
	}

	fmt.Printf("  • Start time:       %s\n", startTime)
	fmt.Printf("  • End time:         %s\n", endTime)
	fmt.Printf("  • Time elapsed:     %s\n", formatDuration(elapsed))
	fmt.Println()
	fmt.Printf("Output directory: %s\n", *outputDir)
	fmt.Println()
}

func getImageFiles(inputDir string) []string {
	var files []string
	filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if supportedExtensions[ext] {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func getOutputPath(inputPath, inputDir, outputDir string) string {
	rel, err := filepath.Rel(inputDir, inputPath)
	if err != nil {
		rel = filepath.Base(inputPath)
	}
	ext := filepath.Ext(rel)
	out := filepath.Join(outputDir, rel[:len(rel)-len(ext)]+".webp")
	return out
}

func filterAlreadyConverted(files []string, inputDir, outputDir string) ([]string, int) {
	var toConvert []string
	skipped := 0
	for _, f := range files {
		outPath := getOutputPath(f, inputDir, outputDir)
		if _, err := os.Stat(outPath); err == nil {
			skipped++
		} else {
			toConvert = append(toConvert, f)
		}
	}
	return toConvert, skipped
}

func convertImage(inputPath, inputDir, outputDir string, quality float32) ConversionResult {
	outputPath := getOutputPath(inputPath, inputDir, outputDir)

	// Ensure output subdirectory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return ConversionResult{
			Input:   inputPath,
			Output:  outputPath,
			Success: false,
			Message: fmt.Sprintf("Failed to create output dir: %v", err),
		}
	}

	// Get original size
	info, err := os.Stat(inputPath)
	if err != nil {
		return ConversionResult{
			Input:   inputPath,
			Output:  outputPath,
			Success: false,
			Message: fmt.Sprintf("Cannot stat input: %v", err),
		}
	}
	originalSize := info.Size()

	// Open and decode image
	f, err := os.Open(inputPath)
	if err != nil {
		return ConversionResult{
			Input:        inputPath,
			Output:       outputPath,
			Success:      false,
			Message:      fmt.Sprintf("Failed to open image: %v", err),
			OriginalSize: originalSize,
		}
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return ConversionResult{
			Input:        inputPath,
			Output:       outputPath,
			Success:      false,
			Message:      fmt.Sprintf("Failed to decode image: %v", err),
			OriginalSize: originalSize,
		}
	}

	// Encode to WebP
	outFile, err := os.Create(outputPath)
	if err != nil {
		return ConversionResult{
			Input:        inputPath,
			Output:       outputPath,
			Success:      false,
			Message:      fmt.Sprintf("Failed to create output file: %v", err),
			OriginalSize: originalSize,
		}
	}
	defer outFile.Close()

	if err := webp.Encode(outFile, img, &webp.Options{Quality: quality}); err != nil {
		os.Remove(outputPath)
		return ConversionResult{
			Input:        inputPath,
			Output:       outputPath,
			Success:      false,
			Message:      fmt.Sprintf("Failed to encode WebP: %v", err),
			OriginalSize: originalSize,
		}
	}

	// Get new size
	outInfo, err := os.Stat(outputPath)
	if err != nil {
		return ConversionResult{
			Input:        inputPath,
			Output:       outputPath,
			Success:      false,
			Message:      fmt.Sprintf("Failed to stat output: %v", err),
			OriginalSize: originalSize,
		}
	}
	newSize := outInfo.Size()

	var savings float64
	if originalSize > 0 {
		savings = float64(int((1.0-float64(newSize)/float64(originalSize))*10000)) / 100.0
	}

	return ConversionResult{
		Input:        inputPath,
		Output:       outputPath,
		Success:      true,
		Message:      "Converted successfully",
		OriginalSize: originalSize,
		NewSize:      newSize,
		Savings:      savings,
	}
}

func formatBytes(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB"}
	size := float64(bytes)
	unitIndex := 0
	for size >= 1024.0 && unitIndex < len(units)-1 {
		size /= 1024.0
		unitIndex++
	}
	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}

func formatDuration(d time.Duration) string {
	totalSecs := d.Seconds()
	if totalSecs < 1.0 {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	minutes := int(totalSecs) / 60
	secs := totalSecs - float64(minutes*60)
	if minutes > 0 {
		return fmt.Sprintf("%dm %.1fs", minutes, secs)
	}
	return fmt.Sprintf("%.1fs", secs)
}
