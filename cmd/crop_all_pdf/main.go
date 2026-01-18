package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cli "pdf-crop/internal/cli"
	"pdf-crop/pkg/crop"
)

type args struct {
	Dir       string
	Threshold float64
	Space     int
	DPI       float64
}

var errHelp = errors.New("help requested")

func printUsage() {
	fmt.Print(cli.CropAllPdfUsage())
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
			val, next, err := cli.RequireValue(argv, i, argv[i])
			if err != nil {
				return parsed, err
			}
			parsed.Dir = val
			i = next
		case "--threshold":
			val, next, err := cli.RequireValue(argv, i, "--threshold")
			if err != nil {
				return parsed, err
			}
			parsed.Threshold, err = cli.ParseFloat(val, "--threshold")
			if err != nil {
				return parsed, err
			}
			i = next
		case "--space":
			val, next, err := cli.RequireValue(argv, i, "--space")
			if err != nil {
				return parsed, err
			}
			parsed.Space, err = cli.ParseInt(val, "--space")
			if err != nil {
				return parsed, err
			}
			i = next
		case "--dpi":
			val, next, err := cli.RequireValue(argv, i, "--dpi")
			if err != nil {
				return parsed, err
			}
			parsed.DPI, err = cli.ParseFloat(val, "--dpi")
			if err != nil {
				return parsed, err
			}
			i = next
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
