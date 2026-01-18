package crop

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type Options struct {
	DPI       float64
	Threshold float64
	Space     int
	CropFrom  string
}

type PageOption struct {
	Number int
	Left   int
	Top    int
	Right  int
	Bottom int
	Output string
}

type PageResult struct {
	PageNo  int
	Media   *types.Rectangle
	Crop    *types.Rectangle
	Output  string
	WasAuto bool
}

func DefaultOptions() Options {
	return Options{
		DPI:       128,
		Threshold: 0.008,
		Space:     5,
		CropFrom:  "center",
	}
}

func normalizeOptions(opts *Options) {
	if opts.DPI <= 0 {
		opts.DPI = 128
	}
	if opts.Space <= 0 {
		opts.Space = 5
	}
	if opts.Threshold <= 0 {
		opts.Threshold = 0.008
	}
	if opts.CropFrom == "" {
		opts.CropFrom = "center"
	}
}

func CropDocument(inputFile, outputFile string, opts Options) error {
	if outputFile == "" {
		return fmt.Errorf("output file is required")
	}
	normalizeOptions(&opts)

	doc, err := fitz.New(inputFile)
	if err != nil {
		return err
	}
	defer doc.Close()

	ctx, err := api.ReadContextFile(inputFile)
	if err != nil {
		return err
	}

	pageCount := doc.NumPage()
	for pageNo := 0; pageNo < pageCount; pageNo++ {
		img, err := doc.ImageDPI(pageNo, opts.DPI)
		if err != nil {
			return fmt.Errorf("render page %d: %w", pageNo, err)
		}
		media, err := pageMediaBox(ctx, pageNo+1)
		if err != nil {
			return fmt.Errorf("page %d mediabox: %w", pageNo, err)
		}
		cropBox := rectFromImage(img, media, opts)
		if err := setCropBox(ctx, pageNo+1, cropBox); err != nil {
			return fmt.Errorf("page %d crop: %w", pageNo, err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return err
	}
	return api.WriteContextFile(ctx, outputFile)
}

func CropPages(inputFile string, pageOptions []PageOption, opts Options) ([]PageResult, error) {
	normalizeOptions(&opts)

	doc, err := fitz.New(inputFile)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	ctx, err := api.ReadContextFile(inputFile)
	if err != nil {
		return nil, err
	}

	if len(pageOptions) == 0 {
		pageOptions = make([]PageOption, 0, doc.NumPage())
		for i := 0; i < doc.NumPage(); i++ {
			pageOptions = append(pageOptions, PageOption{Number: i})
		}
	}

	results := make([]PageResult, 0, len(pageOptions))
	for _, option := range pageOptions {
		pageNo := option.Number
		if pageNo < 0 || pageNo >= doc.NumPage() {
			return nil, fmt.Errorf("page no exceed the page number")
		}
		media, err := pageMediaBox(ctx, pageNo+1)
		if err != nil {
			return nil, fmt.Errorf("page %d mediabox: %w", pageNo, err)
		}

		var cropBox *types.Rectangle
		wasAuto := false
		if option.Left == option.Right || option.Top == option.Bottom {
			img, err := doc.ImageDPI(pageNo, opts.DPI)
			if err != nil {
				return nil, fmt.Errorf("render page %d: %w", pageNo, err)
			}
			cropBox = rectFromImage(img, media, opts)
			wasAuto = true
		} else {
			cropBox = rectFromTopLeft(media, option.Left, option.Top, option.Right, option.Bottom)
		}

		if err := setCropBox(ctx, pageNo+1, cropBox); err != nil {
			return nil, fmt.Errorf("page %d crop: %w", pageNo, err)
		}

		output := option.Output
		if output == "" {
			output = defaultOutputFile(inputFile, pageNo)
		}

		if err := writeSinglePage(ctx, pageNo+1, output); err != nil {
			return nil, fmt.Errorf("page %d write: %w", pageNo, err)
		}

		results = append(results, PageResult{
			PageNo:  pageNo,
			Media:   media,
			Crop:    cropBox,
			Output:  output,
			WasAuto: wasAuto,
		})
	}

	return results, nil
}

func CropAllPagesToSingleFile(inputFile, outputFile string, opts Options) ([]PageResult, error) {
	if outputFile == "" {
		return nil, fmt.Errorf("output file is required")
	}
	normalizeOptions(&opts)

	doc, err := fitz.New(inputFile)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	ctx, err := api.ReadContextFile(inputFile)
	if err != nil {
		return nil, err
	}

	results := make([]PageResult, 0, doc.NumPage())
	for pageNo := 0; pageNo < doc.NumPage(); pageNo++ {
		img, err := doc.ImageDPI(pageNo, opts.DPI)
		if err != nil {
			return nil, fmt.Errorf("render page %d: %w", pageNo, err)
		}
		media, err := pageMediaBox(ctx, pageNo+1)
		if err != nil {
			return nil, fmt.Errorf("page %d mediabox: %w", pageNo, err)
		}
		cropBox := rectFromImage(img, media, opts)
		if err := setCropBox(ctx, pageNo+1, cropBox); err != nil {
			return nil, fmt.Errorf("page %d crop: %w", pageNo, err)
		}
		results = append(results, PageResult{
			PageNo: pageNo,
			Media:  media,
			Crop:   cropBox,
		})
	}

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return nil, err
	}
	if err := api.WriteContextFile(ctx, outputFile); err != nil {
		return nil, err
	}

	for i := range results {
		results[i].Output = outputFile
	}
	return results, nil
}

func rectFromImage(img *image.RGBA, media *types.Rectangle, opts Options) *types.Rectangle {
	leftF, topF, rightF, bottomF := detectFrame(img, opts.Space, opts.Threshold, opts.CropFrom)
	width := media.UR.X - media.LL.X
	height := media.UR.Y - media.LL.Y
	left := int(leftF * width)
	top := int(topF * height)
	right := int(rightF * width)
	bottom := int(bottomF * height)
	return rectFromTopLeft(media, left, top, right, bottom)
}

func rectFromTopLeft(media *types.Rectangle, left, top, right, bottom int) *types.Rectangle {
	height := media.UR.Y - media.LL.Y

	leftX := media.LL.X + float64(left)
	rightX := media.LL.X + float64(right)
	upperY := media.LL.Y + (height - float64(top))
	lowerY := media.LL.Y + (height - float64(bottom))

	llx := math.Min(leftX, rightX)
	urx := math.Max(leftX, rightX)
	lly := math.Min(lowerY, upperY)
	ury := math.Max(lowerY, upperY)

	return types.NewRectangle(llx, lly, urx, ury)
}

func pageMediaBox(ctx *model.Context, pageNumber int) (*types.Rectangle, error) {
	pages, err := ctx.PageBoundaries(types.IntSet{pageNumber: true})
	if err != nil {
		// Fallback to default A4 size.
		return types.RectForDim(595, 842), nil
	}
	if len(pages) == 0 {
		return types.RectForDim(595, 842), nil
	}
	media := pages[0].MediaBox()
	if media == nil {
		return types.RectForDim(595, 842), nil
	}
	return media, nil
}

func setCropBox(ctx *model.Context, pageNumber int, rect *types.Rectangle) error {
	if rect == nil {
		return nil
	}
	pb := &model.PageBoundaries{
		Crop: &model.Box{Rect: rect},
	}
	return ctx.AddPageBoundaries(types.IntSet{pageNumber: true}, pb)
}

func writeSinglePage(ctx *model.Context, pageNumber int, output string) error {
	reader, err := api.ExtractPage(ctx, pageNumber)
	if err != nil {
		return err
	}
	outputDir := filepath.Dir(output)
	outputName := filepath.Base(output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	return api.WritePage(reader, outputDir, outputName, pageNumber)
}

func defaultOutputFile(inputFile string, pageNo int) string {
	ext := filepath.Ext(inputFile)
	base := inputFile[:len(inputFile)-len(ext)]
	return fmt.Sprintf("%s - page %2d.pdf", base, pageNo)
}

func RectString(rect *types.Rectangle) string {
	if rect == nil {
		return "(0, 0), (0, 0)"
	}
	return fmt.Sprintf("(%d, %d), (%d, %d)", int(rect.LL.X), int(rect.LL.Y), int(rect.UR.X), int(rect.UR.Y))
}
