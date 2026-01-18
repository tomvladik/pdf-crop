package crop

import (
	"image"
	"image/color"
	"testing"
)

// createTestImage creates a simple test image with a specified pattern
func createTestImage(width, height int, nonWhitePixels []image.Point) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Add non-white pixels
	for _, p := range nonWhitePixels {
		if p.X >= 0 && p.X < width && p.Y >= 0 && p.Y < height {
			img.Set(p.X, p.Y, color.Black)
		}
	}

	return img
}

func TestBuildDetectData(t *testing.T) {
	// Create a 10x10 image with a few black pixels
	nonWhite := []image.Point{
		{X: 5, Y: 5},
		{X: 6, Y: 5},
		{X: 5, Y: 6},
	}
	img := createTestImage(10, 10, nonWhite)

	d := buildDetectData(img)

	if d.width != 10 {
		t.Errorf("Expected width 10, got %d", d.width)
	}
	if d.height != 10 {
		t.Errorf("Expected height 10, got %d", d.height)
	}

	// Check that non-white pixels are counted
	count := d.countNonZero(0, 0, 10, 10)
	if count != 3 {
		t.Errorf("Expected 3 non-white pixels, got %d", count)
	}
}

func TestCountNonZero(t *testing.T) {
	// Create a test image with black pixels in a known region
	nonWhite := []image.Point{
		{X: 2, Y: 2}, {X: 3, Y: 2}, {X: 4, Y: 2},
		{X: 2, Y: 3}, {X: 3, Y: 3}, {X: 4, Y: 3},
		{X: 2, Y: 4}, {X: 3, Y: 4}, {X: 4, Y: 4},
	}
	img := createTestImage(10, 10, nonWhite)
	d := buildDetectData(img)

	tests := []struct {
		name     string
		x0, y0   int
		x1, y1   int
		expected int
	}{
		{"Full region", 2, 2, 5, 5, 9},
		{"Partial region", 2, 2, 3, 3, 1},
		{"Empty region", 0, 0, 2, 2, 0},
		{"Out of bounds negative", -1, -1, 1, 1, 0},
		{"Out of bounds positive", 9, 9, 20, 20, 0},
		{"Zero width", 2, 2, 2, 5, 0},
		{"Zero height", 2, 2, 5, 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.countNonZero(tt.x0, tt.y0, tt.x1, tt.y1)
			if result != tt.expected {
				t.Errorf("countNonZero(%d, %d, %d, %d) = %d, expected %d",
					tt.x0, tt.y0, tt.x1, tt.y1, result, tt.expected)
			}
		})
	}
}

func TestDetectCenter(t *testing.T) {
	// Create an image with a dense row and column
	nonWhite := []image.Point{
		// Dense row at y=5
		{X: 0, Y: 5}, {X: 1, Y: 5}, {X: 2, Y: 5}, {X: 3, Y: 5}, {X: 4, Y: 5},
		{X: 5, Y: 5}, {X: 6, Y: 5}, {X: 7, Y: 5}, {X: 8, Y: 5}, {X: 9, Y: 5},
		// Dense column at x=7
		{X: 7, Y: 0}, {X: 7, Y: 1}, {X: 7, Y: 2}, {X: 7, Y: 3}, {X: 7, Y: 4},
		{X: 7, Y: 6}, {X: 7, Y: 7}, {X: 7, Y: 8}, {X: 7, Y: 9},
	}
	img := createTestImage(10, 10, nonWhite)
	d := buildDetectData(img)

	cx, cy := detectCenter(d)

	if cx != 7 {
		t.Errorf("Expected center X=7, got %d", cx)
	}
	if cy != 5 {
		t.Errorf("Expected center Y=5, got %d", cy)
	}
}

func TestDetectTop(t *testing.T) {
	// Create an image with content starting at y=3
	nonWhite := []image.Point{}
	for y := 3; y < 8; y++ {
		for x := 2; x < 8; x++ {
			nonWhite = append(nonWhite, image.Point{X: x, Y: y})
		}
	}
	img := createTestImage(10, 10, nonWhite)
	d := buildDetectData(img)

	cy := 5 // center Y
	space := 1
	threshold := 2

	top := detectTop(d, cy, space, threshold)

	// Should detect somewhere near y=3
	if top < 2 || top > 4 {
		t.Errorf("Expected top around 3, got %d", top)
	}
}

func TestDetectBottom(t *testing.T) {
	// Create an image with content ending at y=7
	nonWhite := []image.Point{}
	for y := 3; y < 8; y++ {
		for x := 2; x < 8; x++ {
			nonWhite = append(nonWhite, image.Point{X: x, Y: y})
		}
	}
	img := createTestImage(10, 10, nonWhite)
	d := buildDetectData(img)

	cy := 5 // center Y
	space := 1
	threshold := 2

	bottom := detectBottom(d, cy, space, threshold)

	// Should detect somewhere near y=8
	if bottom < 7 || bottom > 9 {
		t.Errorf("Expected bottom around 8, got %d", bottom)
	}
}

func TestDetectLeft(t *testing.T) {
	// Create an image with content starting at x=2
	nonWhite := []image.Point{}
	for y := 2; y < 8; y++ {
		for x := 2; x < 8; x++ {
			nonWhite = append(nonWhite, image.Point{X: x, Y: y})
		}
	}
	img := createTestImage(10, 10, nonWhite)
	d := buildDetectData(img)

	cx := 5 // center X
	top := 2
	bottom := 8
	space := 1
	threshold := 2

	left := detectLeft(d, cx, top, bottom, space, threshold)

	// Should detect somewhere near x=2
	if left < 1 || left > 3 {
		t.Errorf("Expected left around 2, got %d", left)
	}
}

func TestDetectRight(t *testing.T) {
	// Create an image with content ending at x=7
	nonWhite := []image.Point{}
	for y := 2; y < 8; y++ {
		for x := 2; x < 8; x++ {
			nonWhite = append(nonWhite, image.Point{X: x, Y: y})
		}
	}
	img := createTestImage(10, 10, nonWhite)
	d := buildDetectData(img)

	cx := 5 // center X
	top := 2
	bottom := 8
	space := 1
	threshold := 2

	right := detectRight(d, cx, top, bottom, space, threshold)

	// Should detect somewhere near x=8
	if right < 7 || right > 9 {
		t.Errorf("Expected right around 8, got %d", right)
	}
}

func TestDetectBorder(t *testing.T) {
	// Create an image with content in a specific region
	nonWhite := []image.Point{}
	for y := 2; y < 8; y++ {
		for x := 3; x < 7; x++ {
			nonWhite = append(nonWhite, image.Point{X: x, Y: y})
		}
	}
	img := createTestImage(10, 10, nonWhite)
	d := buildDetectData(img)

	space := 1
	thresholdW := 2
	thresholdH := 2

	top, bottom, left, right := detectBorder(d, space, thresholdW, thresholdH)

	// Check that borders are detected approximately
	if top < 1 || top > 3 {
		t.Errorf("Expected top around 2, got %d", top)
	}
	if bottom < 7 || bottom > 9 {
		t.Errorf("Expected bottom around 8, got %d", bottom)
	}
	if left < 2 || left > 4 {
		t.Errorf("Expected left around 3, got %d", left)
	}
	if right < 6 || right > 8 {
		t.Errorf("Expected right around 7, got %d", right)
	}
}

func TestDetectFrame(t *testing.T) {
	// Create an image with content in the center
	nonWhite := []image.Point{}
	for y := 20; y < 80; y++ {
		for x := 30; x < 70; x++ {
			nonWhite = append(nonWhite, image.Point{X: x, Y: y})
		}
	}
	img := createTestImage(100, 100, nonWhite)

	space := 5
	threshold := 0.1
	cropFrom := "center"

	left, top, right, bottom := detectFrame(img, space, threshold, cropFrom)

	// Verify values are in range [0, 1]
	if left < 0 || left > 1 {
		t.Errorf("Left fraction out of range: %f", left)
	}
	if top < 0 || top > 1 {
		t.Errorf("Top fraction out of range: %f", top)
	}
	if right < 0 || right > 1 {
		t.Errorf("Right fraction out of range: %f", right)
	}
	if bottom < 0 || bottom > 1 {
		t.Errorf("Bottom fraction out of range: %f", bottom)
	}

	// Verify relative positions
	if left >= right {
		t.Errorf("Left (%f) should be less than right (%f)", left, right)
	}
	if top >= bottom {
		t.Errorf("Top (%f) should be less than bottom (%f)", top, bottom)
	}
}

func TestDetectFrameBorder(t *testing.T) {
	// Test border detection mode
	nonWhite := []image.Point{}
	for y := 20; y < 80; y++ {
		for x := 30; x < 70; x++ {
			nonWhite = append(nonWhite, image.Point{X: x, Y: y})
		}
	}
	img := createTestImage(100, 100, nonWhite)

	space := 5
	threshold := 0.1
	cropFrom := "border"

	left, top, right, bottom := detectFrame(img, space, threshold, cropFrom)

	// Verify values are in range [0, 1]
	if left < 0 || left > 1 {
		t.Errorf("Left fraction out of range: %f", left)
	}
	if top < 0 || top > 1 {
		t.Errorf("Top fraction out of range: %f", top)
	}
	if right < 0 || right > 1 {
		t.Errorf("Right fraction out of range: %f", right)
	}
	if bottom < 0 || bottom > 1 {
		t.Errorf("Bottom fraction out of range: %f", bottom)
	}
}

func TestDetectFrameEmptyImage(t *testing.T) {
	// Test with empty (all white) image
	img := createTestImage(100, 100, []image.Point{})

	space := 5
	threshold := 0.1
	cropFrom := "center"

	left, top, right, bottom := detectFrame(img, space, threshold, cropFrom)

	// Should return zeros for empty image
	if left != 0 || top != 0 || right != 0 || bottom != 0 {
		t.Errorf("Expected all zeros for empty image, got (%f, %f, %f, %f)",
			left, top, right, bottom)
	}
}
