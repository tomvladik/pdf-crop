package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"pdf-crop/internal/crop"
)

type args struct {
	Dir       string
	Threshold float64
	Space     int
	DPI       float64
}

var errHelp = errors.New("help requested")

func printUsage() {
	fmt.Println("crop_all_pdf - Crop all PDFs in a directory")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  crop_all_pdf --dir <path> [--threshold <float>] [--space <int>] [--dpi <float>]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -d, --dir           Directory containing PDFs (default: current directory)")
	fmt.Println("      --threshold      Detection threshold (default: 0.1)")
	fmt.Println("      --space          Extra whitespace in points (default: 5)")
	fmt.Println("      --dpi            Rasterization DPI (default: 128)")
	fmt.Println("  -h, --help          Show this help and exit")
}

func parseArgs(argv []string) (args, error) {
	parsed := args{
		Threshold: 0.1,
		Space:     5,
		DPI:       128,
	}
	for i := 0; i < len(argv); i++ {
		switch argv[i] {
		case "-h", "--help":
			return parsed, errHelp
		case "-d", "--dir":
			if i+1 >= len(argv) {
				return parsed, fmt.Errorf("missing value for %s", argv[i])
			}
			parsed.Dir = argv[i+1]
			i++
		case "--threshold":
			if i+1 >= len(argv) {
				return parsed, fmt.Errorf("missing value for --threshold")
			}
			val, err := strconv.ParseFloat(argv[i+1], 64)
			if err != nil {
				return parsed, fmt.Errorf("invalid --threshold: %w", err)
			}
			parsed.Threshold = val
			i++
		case "--space":
			if i+1 >= len(argv) {
				return parsed, fmt.Errorf("missing value for --space")
			}
			val, err := strconv.Atoi(argv[i+1])
			if err != nil {
				return parsed, fmt.Errorf("invalid --space: %w", err)
			}
			parsed.Space = val
			i++
		case "--dpi":
			if i+1 >= len(argv) {
				return parsed, fmt.Errorf("missing value for --dpi")
			}
			val, err := strconv.ParseFloat(argv[i+1], 64)
			if err != nil {
				return parsed, fmt.Errorf("invalid --dpi: %w", err)
			}
			parsed.DPI = val
			i++
		default:
			return parsed, fmt.Errorf("unknown argument: %s", argv[i])
		}
	}
	return parsed, nil
}

func main() {
	parsed, err := parseArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, errHelp) {
			printUsage()
			os.Exit(0)
		}
		// Print error and usage to stdout as requested
		fmt.Println(err)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}

	if parsed.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		parsed.Dir = cwd
	}

	entries, err := os.ReadDir(parsed.Dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	options := crop.Options{
		DPI:       parsed.DPI,
		Threshold: parsed.Threshold,
		Space:     parsed.Space,
		CropFrom:  "center",
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".pdf" && filepath.Ext(entry.Name()) != ".PDF" {
			continue
		}
		inputPath := filepath.Join(parsed.Dir, entry.Name())
		outputPath := filepath.Join(parsed.Dir, "cropped_"+entry.Name())
		fmt.Printf("Processing: %s -> %s\n", inputPath, outputPath)
		if _, err := crop.CropAllPagesToSingleFile(inputPath, outputPath, options); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", entry.Name(), err)
			continue
		}
		fmt.Printf("Successfully processed: %s\n", entry.Name())
	}
}