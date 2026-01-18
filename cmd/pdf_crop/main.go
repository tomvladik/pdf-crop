package main

import (
	"errors"
	"fmt"
	"os"

	cli "pdf-crop/internal/cli"
	"pdf-crop/pkg/crop"
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
			val, next, err := cli.RequireValue(argv, i, argv[i])
			if err != nil {
				return parsed, err
			}
			parsed.InputFile = val
			i = next
		case "-p", "--page":
			vals, next, err := cli.RequireValues(argv, i, 6, "--page")
			if err != nil {
				return parsed, err
			}
			pageNo, err := cli.ParseInt(vals[0], "page number")
			if err != nil {
				return parsed, err
			}
			left, err := cli.ParseInt(vals[1], "left value")
			if err != nil {
				return parsed, err
			}
			top, err := cli.ParseInt(vals[2], "top value")
			if err != nil {
				return parsed, err
			}
			right, err := cli.ParseInt(vals[3], "right value")
			if err != nil {
				return parsed, err
			}
			bottom, err := cli.ParseInt(vals[4], "bottom value")
			if err != nil {
				return parsed, err
			}
			output := vals[5]
			parsed.Pages = append(parsed.Pages, crop.PageOption{
				Number: pageNo,
				Left:   left,
				Top:    top,
				Right:  right,
				Bottom: bottom,
				Output: output,
			})
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
