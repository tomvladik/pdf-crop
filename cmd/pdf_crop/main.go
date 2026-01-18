package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"pdf-crop/internal/crop"
	"pdf-crop/internal/cli"
)

type args struct {
	InputFile string
	Pages     []crop.PageOption
	Space     int
	Threshold float64
	DPI       float64
}

var errHelp = errors.New("help requested")

func printUsage() {
	fmt.Print(cli.PdfCropUsage())
}

func parseArgs(argv []string) (args, error) {
	parsed := args{
		Space:     5,
		Threshold: 0.008,
		DPI:       128,
	}
	for i := 0; i < len(argv); i++ {
		switch argv[i] {
		case "-h", "--help":
			return parsed, errHelp
		case "-i", "--input_file":
			if i+1 >= len(argv) {
				return parsed, fmt.Errorf("missing value for %s", argv[i])
			}
			parsed.InputFile = argv[i+1]
			i++
		case "-p", "--page":
			if i+6 >= len(argv) {
				return parsed, fmt.Errorf("--page requires 6 arguments")
			}
			pageNo, err := strconv.Atoi(argv[i+1])
			if err != nil {
				return parsed, fmt.Errorf("invalid page number: %w", err)
			}
			left, err := strconv.Atoi(argv[i+2])
			if err != nil {
				return parsed, fmt.Errorf("invalid left value: %w", err)
			}
			top, err := strconv.Atoi(argv[i+3])
			if err != nil {
				return parsed, fmt.Errorf("invalid top value: %w", err)
			}
			right, err := strconv.Atoi(argv[i+4])
			if err != nil {
				return parsed, fmt.Errorf("invalid right value: %w", err)
			}
			bottom, err := strconv.Atoi(argv[i+5])
			if err != nil {
				return parsed, fmt.Errorf("invalid bottom value: %w", err)
			}
			output := argv[i+6]
			parsed.Pages = append(parsed.Pages, crop.PageOption{
				Number: pageNo,
				Left:   left,
				Top:    top,
				Right:  right,
				Bottom: bottom,
				Output: output,
			})
			i += 6
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

	if parsed.InputFile == "" {
		return parsed, fmt.Errorf("-i/--input_file is required")
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

	options := crop.Options{
		DPI:       parsed.DPI,
		Threshold: parsed.Threshold,
		Space:     parsed.Space,
		CropFrom:  "center",
	}

	results, err := crop.CropPages(parsed.InputFile, parsed.Pages, options)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, res := range results {
		fmt.Printf("%d %s %s %s\n", res.PageNo, crop.RectString(res.Media), crop.RectString(res.Crop), res.Output)
	}
}
