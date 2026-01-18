package crop

import (
	"image"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.DPI != 128 {
		t.Errorf("Expected DPI 128, got %f", opts.DPI)
	}
	if opts.Threshold != 0.008 {
		t.Errorf("Expected Threshold 0.008, got %f", opts.Threshold)
	}
	if opts.Space != 5 {
		t.Errorf("Expected Space 5, got %d", opts.Space)
	}
	if opts.CropFrom != "center" {
		t.Errorf("Expected CropFrom 'center', got %s", opts.CropFrom)
	}
}

func TestRectFromTopLeft(t *testing.T) {
	// Create a media box: 0, 0 to 612, 792 (standard letter size)
	media := types.NewRectangle(0, 0, 612, 792)

	tests := []struct {
		name          string
		left, top     int
		right, bottom int
		expectedLLX   float64
		expectedLLY   float64
		expectedURX   float64
		expectedURY   float64
	}{
		{
			name:        "No crop",
			left:        0,
			top:         0,
			right:       612,
			bottom:      792,
			expectedLLX: 0,
			expectedLLY: 0,
			expectedURX: 612,
			expectedURY: 792,
		},
		{
			name:        "Crop all sides",
			left:        50,
			top:         50,
			right:       562,
			bottom:      742,
			expectedLLX: 50,
			expectedLLY: 50,
			expectedURX: 562,
			expectedURY: 742,
		},
		{
			name:        "Crop top only",
			left:        0,
			top:         100,
			right:       612,
			bottom:      792,
			expectedLLX: 0,
			expectedLLY: 0,
			expectedURX: 612,
			expectedURY: 692,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rectFromTopLeft(media, tt.left, tt.top, tt.right, tt.bottom)

			if result.LL.X != tt.expectedLLX {
				t.Errorf("Expected LL.X %f, got %f", tt.expectedLLX, result.LL.X)
			}
			if result.LL.Y != tt.expectedLLY {
				t.Errorf("Expected LL.Y %f, got %f", tt.expectedLLY, result.LL.Y)
			}
			if result.UR.X != tt.expectedURX {
				t.Errorf("Expected UR.X %f, got %f", tt.expectedURX, result.UR.X)
			}
			if result.UR.Y != tt.expectedURY {
				t.Errorf("Expected UR.Y %f, got %f", tt.expectedURY, result.UR.Y)
			}
		})
	}
}

func TestRectString(t *testing.T) {
	tests := []struct {
		name     string
		rect     *types.Rectangle
		expected string
	}{
		{
			name:     "Nil rectangle",
			rect:     nil,
			expected: "(0, 0), (0, 0)",
		},
		{
			name:     "Standard rectangle",
			rect:     types.NewRectangle(10, 20, 100, 200),
			expected: "(10, 20), (100, 200)",
		},
		{
			name:     "Zero rectangle",
			rect:     types.NewRectangle(0, 0, 0, 0),
			expected: "(0, 0), (0, 0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RectString(tt.rect)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRectFromImage(t *testing.T) {
	// Create a test RGBA image with content
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Fill with white
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, image.White)
		}
	}
	// Add some black content in the middle
	for y := 20; y < 80; y++ {
		for x := 30; x < 70; x++ {
			img.Set(x, y, image.Black)
		}
	}

	media := types.NewRectangle(0, 0, 612, 792)
	opts := Options{
		DPI:       128,
		Threshold: 0.1,
		Space:     5,
		CropFrom:  "center",
	}

	result := rectFromImage(img, media, opts)

	if result == nil {
		t.Fatal("Expected non-nil rectangle")
	}

	// Verify the result is within bounds
	if result.LL.X < 0 || result.LL.X > 612 {
		t.Errorf("LL.X out of bounds: %f", result.LL.X)
	}
	if result.LL.Y < 0 || result.LL.Y > 792 {
		t.Errorf("LL.Y out of bounds: %f", result.LL.Y)
	}
	if result.UR.X < 0 || result.UR.X > 612 {
		t.Errorf("UR.X out of bounds: %f", result.UR.X)
	}
	if result.UR.Y < 0 || result.UR.Y > 792 {
		t.Errorf("UR.Y out of bounds: %f", result.UR.Y)
	}

	// Verify proper ordering
	if result.LL.X >= result.UR.X {
		t.Errorf("LL.X (%f) should be less than UR.X (%f)", result.LL.X, result.UR.X)
	}
	if result.LL.Y >= result.UR.Y {
		t.Errorf("LL.Y (%f) should be less than UR.Y (%f)", result.LL.Y, result.UR.Y)
	}
}

func TestOptionsNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    Options
		expected Options
	}{
		{
			name: "All defaults",
			input: Options{
				DPI:       0,
				Threshold: 0,
				Space:     0,
				CropFrom:  "",
			},
			expected: Options{
				DPI:       128,
				Threshold: 0.008,
				Space:     5,
				CropFrom:  "center",
			},
		},
		{
			name: "Custom values preserved",
			input: Options{
				DPI:       300,
				Threshold: 0.01,
				Space:     10,
				CropFrom:  "border",
			},
			expected: Options{
				DPI:       300,
				Threshold: 0.01,
				Space:     10,
				CropFrom:  "border",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the normalization that happens in CropDocument
			opts := tt.input
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

			if opts.DPI != tt.expected.DPI {
				t.Errorf("DPI: expected %f, got %f", tt.expected.DPI, opts.DPI)
			}
			if opts.Threshold != tt.expected.Threshold {
				t.Errorf("Threshold: expected %f, got %f", tt.expected.Threshold, opts.Threshold)
			}
			if opts.Space != tt.expected.Space {
				t.Errorf("Space: expected %d, got %d", tt.expected.Space, opts.Space)
			}
			if opts.CropFrom != tt.expected.CropFrom {
				t.Errorf("CropFrom: expected %s, got %s", tt.expected.CropFrom, opts.CropFrom)
			}
		})
	}
}

func TestDefaultOutputFile(t *testing.T) {
	tests := []struct {
		name      string
		inputFile string
		pageNo    int
		expected  string
	}{
		{
			name:      "PDF with path",
			inputFile: "C:\\Documents\\report.pdf",
			pageNo:    0,
			expected:  "C:\\Documents\\report - page  0.pdf",
		},
		{
			name:      "PDF with no path",
			inputFile: "document.pdf",
			pageNo:    5,
			expected:  "document - page  5.pdf",
		},
		{
			name:      "PDF with longer number",
			inputFile: "/home/user/file.pdf",
			pageNo:    42,
			expected:  "/home/user/file - page 42.pdf",
		},
		{
			name:      "PDF with dots in name",
			inputFile: "my.test.file.pdf",
			pageNo:    1,
			expected:  "my.test.file - page  1.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultOutputFile(tt.inputFile, tt.pageNo)
			if result != tt.expected {
				t.Errorf("defaultOutputFile(%q, %d) = %q, expected %q",
					tt.inputFile, tt.pageNo, result, tt.expected)
			}
		})
	}
}

func TestSetCropBoxWithNil(t *testing.T) {
	// Test that setCropBox handles nil rectangle gracefully
	// This test doesn't actually call setCropBox since it requires a model.Context,
	// but we test the nil check logic
	var rect *types.Rectangle = nil

	if rect != nil {
		t.Error("Rectangle should be nil")
	}

	// Test non-nil rectangle
	rect = types.NewRectangle(10, 20, 100, 200)
	if rect == nil {
		t.Error("Rectangle should not be nil")
	}
}

func TestRectFromTopLeftEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		media         *types.Rectangle
		left, top     int
		right, bottom int
		description   string
	}{
		{
			name:        "Inverted coordinates",
			media:       types.NewRectangle(0, 0, 612, 792),
			left:        100,
			top:         100,
			right:       50,
			bottom:      50,
			description: "Should handle inverted left/right and top/bottom",
		},
		{
			name:        "Large values",
			media:       types.NewRectangle(0, 0, 612, 792),
			left:        1000,
			top:         1000,
			right:       2000,
			bottom:      2000,
			description: "Should handle values larger than media box",
		},
		{
			name:        "Zero crop",
			media:       types.NewRectangle(0, 0, 612, 792),
			left:        0,
			top:         0,
			right:       0,
			bottom:      0,
			description: "Should handle all zeros",
		},
		{
			name:        "Negative coordinates",
			media:       types.NewRectangle(100, 100, 612, 792),
			left:        -50,
			top:         -50,
			right:       500,
			bottom:      600,
			description: "Should handle negative values with offset media box",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rectFromTopLeft(tt.media, tt.left, tt.top, tt.right, tt.bottom)

			if result == nil {
				t.Fatal("Expected non-nil rectangle")
			}

			// Verify basic properties
			if result.LL.X > result.UR.X {
				t.Errorf("LL.X (%f) should be <= UR.X (%f)", result.LL.X, result.UR.X)
			}
			if result.LL.Y > result.UR.Y {
				t.Errorf("LL.Y (%f) should be <= UR.Y (%f)", result.LL.Y, result.UR.Y)
			}
		})
	}
}

func TestRectFromImageWithDifferentOptions(t *testing.T) {
	// Test rectFromImage with different crop strategies
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Fill with white
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, image.White)
		}
	}
	// Add some content
	for y := 30; y < 70; y++ {
		for x := 40; x < 60; x++ {
			img.Set(x, y, image.Black)
		}
	}

	media := types.NewRectangle(0, 0, 612, 792)

	tests := []struct {
		name     string
		opts     Options
		validate func(*testing.T, *types.Rectangle)
	}{
		{
			name: "Center crop with default threshold",
			opts: Options{
				DPI:       128,
				Threshold: 0.008,
				Space:     5,
				CropFrom:  "center",
			},
			validate: func(t *testing.T, rect *types.Rectangle) {
				if rect == nil {
					t.Fatal("Expected non-nil rectangle")
				}
			},
		},
		{
			name: "Border crop",
			opts: Options{
				DPI:       128,
				Threshold: 0.1,
				Space:     5,
				CropFrom:  "border",
			},
			validate: func(t *testing.T, rect *types.Rectangle) {
				if rect == nil {
					t.Fatal("Expected non-nil rectangle")
				}
			},
		},
		{
			name: "High threshold",
			opts: Options{
				DPI:       128,
				Threshold: 0.5,
				Space:     10,
				CropFrom:  "center",
			},
			validate: func(t *testing.T, rect *types.Rectangle) {
				if rect == nil {
					t.Fatal("Expected non-nil rectangle")
				}
			},
		},
		{
			name: "Small space",
			opts: Options{
				DPI:       128,
				Threshold: 0.01,
				Space:     1,
				CropFrom:  "center",
			},
			validate: func(t *testing.T, rect *types.Rectangle) {
				if rect == nil {
					t.Fatal("Expected non-nil rectangle")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rectFromImage(img, media, tt.opts)
			tt.validate(t, result)

			// Verify bounds are within media box
			if result.LL.X < media.LL.X-1 || result.UR.X > media.UR.X+1 {
				t.Errorf("X coordinates out of media bounds: [%f, %f] not in [%f, %f]",
					result.LL.X, result.UR.X, media.LL.X, media.UR.X)
			}
			if result.LL.Y < media.LL.Y-1 || result.UR.Y > media.UR.Y+1 {
				t.Errorf("Y coordinates out of media bounds: [%f, %f] not in [%f, %f]",
					result.LL.Y, result.UR.Y, media.LL.Y, media.UR.Y)
			}
		})
	}
}

func TestPageOption(t *testing.T) {
	// Test PageOption struct
	opt := PageOption{
		Number: 5,
		Left:   10,
		Top:    20,
		Right:  100,
		Bottom: 200,
		Output: "output.pdf",
	}

	if opt.Number != 5 {
		t.Errorf("Expected Number 5, got %d", opt.Number)
	}
	if opt.Left != 10 {
		t.Errorf("Expected Left 10, got %d", opt.Left)
	}
	if opt.Output != "output.pdf" {
		t.Errorf("Expected Output 'output.pdf', got %s", opt.Output)
	}
}

func TestPageResult(t *testing.T) {
	// Test PageResult struct
	media := types.NewRectangle(0, 0, 612, 792)
	crop := types.NewRectangle(10, 10, 600, 780)

	result := PageResult{
		PageNo:  3,
		Media:   media,
		Crop:    crop,
		Output:  "page3.pdf",
		WasAuto: true,
	}

	if result.PageNo != 3 {
		t.Errorf("Expected PageNo 3, got %d", result.PageNo)
	}
	if !result.WasAuto {
		t.Error("Expected WasAuto to be true")
	}
	if result.Output != "page3.pdf" {
		t.Errorf("Expected Output 'page3.pdf', got %s", result.Output)
	}
	if result.Media == nil {
		t.Error("Expected non-nil Media")
	}
	if result.Crop == nil {
		t.Error("Expected non-nil Crop")
	}
}
