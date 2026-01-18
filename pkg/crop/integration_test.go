package crop

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// makeTestImage creates a simple white image with a central black rectangle.
func makeTestImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// Fill white
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.White)
		}
	}
	// Add central black content to crop around
	mx, my := w/2, h/2
	for y := my - (h / 5); y < my+(h/5); y++ {
		for x := mx - (w / 5); x < mx+(w/5); x++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, color.Black)
			}
		}
	}
	return img
}

// createPDFViaImport creates a single-page PDF from an image using pdfcpu ImportImagesFile.
func createPDFViaImport(t *testing.T, imgPath, pdfPath string) {
	imp, err := api.Import("", types.POINTS)
	if err != nil {
		t.Fatalf("import config: %v", err)
	}
	// Create PDF from image; outFile created if needed.
	if err := api.ImportImagesFile([]string{imgPath}, pdfPath, imp, nil); err != nil {
		t.Fatalf("import image to pdf: %v", err)
	}
}

func writePNG(t *testing.T, p string, img image.Image) {
	f, err := os.Create(p)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}

func TestCropPages_OnImportedImagePDF(t *testing.T) {
	// Temp workspace
	tdir := t.TempDir()
	pngPath := filepath.Join(tdir, "fixture.png")
	pdfPath := filepath.Join(tdir, "fixture.pdf")

	img := makeTestImage(600, 800)
	writePNG(t, pngPath, img)
	createPDFViaImport(t, pngPath, pdfPath)

	// Run crop on all pages (auto detection)
	opts := Options{DPI: 128, Threshold: 0.05, Space: 5, CropFrom: "center"}
	results, err := CropPages(pdfPath, nil, opts)
	if err != nil {
		t.Fatalf("CropPages: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected at least one page result")
	}
	res := results[0]
	if res.Crop == nil || res.Media == nil {
		t.Fatalf("missing crop or media boxes")
	}
	// Expect crop tighter than media (since image has central content)
	if !(int(res.Crop.LL.X) > int(res.Media.LL.X) && int(res.Crop.UR.X) < int(res.Media.UR.X)) {
		t.Errorf("expected X crop tighter than media: crop=%s media=%s", RectString(res.Crop), RectString(res.Media))
	}
	if !(int(res.Crop.LL.Y) > int(res.Media.LL.Y) && int(res.Crop.UR.Y) < int(res.Media.UR.Y)) {
		t.Errorf("expected Y crop tighter than media: crop=%s media=%s", RectString(res.Crop), RectString(res.Media))
	}
	// Validate a PDF was written into the directory
	outDir := filepath.Dir(res.Output)
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	found := false
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".pdf" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a single-page PDF in %s, found none", outDir)
	}
}

func createMultiPagePDFViaImport(t *testing.T, imgPaths []string, pdfPath string) {
	// Create individual single-page PDFs then merge to ensure proper page dictionaries.
	tmpPDFs := make([]string, 0, len(imgPaths))
	for i, p := range imgPaths {
		out := filepath.Join(filepath.Dir(pdfPath), filepath.Base(pdfPath)+fmt.Sprintf("_%d.pdf", i+1))
		imp, err := api.Import("", types.POINTS)
		if err != nil {
			t.Fatalf("import config: %v", err)
		}
		if err := api.ImportImagesFile([]string{p}, out, imp, nil); err != nil {
			t.Fatalf("import image to pdf: %v", err)
		}
		tmpPDFs = append(tmpPDFs, out)
	}
	if err := api.MergeCreateFile(tmpPDFs, pdfPath, false, nil); err != nil {
		t.Fatalf("merge pdfs: %v", err)
	}
}

func TestCropAllPagesToSingleFile_OnImportedImagePDF(t *testing.T) {
	tdir := t.TempDir()
	pngPath := filepath.Join(tdir, "fixture.png")
	pdfPath := filepath.Join(tdir, "fixture.pdf")
	outPath := filepath.Join(tdir, "cropped.pdf")

	img := makeTestImage(500, 700)
	writePNG(t, pngPath, img)
	createPDFViaImport(t, pngPath, pdfPath)

	opts := Options{DPI: 128, Threshold: 0.05, Space: 5, CropFrom: "center"}
	results, err := CropAllPagesToSingleFile(pdfPath, outPath, opts)
	if err != nil {
		t.Fatalf("CropAllPagesToSingleFile: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected at least one page result")
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file written: %v", err)
	}
	// Basic sanity: each result should have crop box set
	for i, r := range results {
		if r.Crop == nil || r.Media == nil {
			t.Errorf("page %d: missing crop or media boxes", i)
		}
	}
}

// Helper to compute relative fractions within the media box.
func rectFractions(r *types.Rectangle, media *types.Rectangle) (left, top, right, bottom float64) {
	width := media.UR.X - media.LL.X
	height := media.UR.Y - media.LL.Y
	if width == 0 || height == 0 {
		return 0, 0, 0, 0
	}
	left = (r.LL.X - media.LL.X) / width
	right = (r.UR.X - media.LL.X) / width
	// Y increases upwards in PDF coordinates
	top = (r.UR.Y - media.LL.Y) / height
	bottom = (r.LL.Y - media.LL.Y) / height
	return
}

func makeWhiteImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.White)
		}
	}
	return img
}

func makeEdgeContentImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.White)
		}
	}
	// Draw a large border-ish rectangle near the edges (5px margin)
	margin := 5
	for y := margin; y < h-margin; y++ {
		for x := margin; x < w-margin; x++ {
			if y < margin+10 || y > h-margin-10 || x < margin+10 || x > w-margin-10 {
				img.Set(x, y, color.Black)
			}
		}
	}
	return img
}

func TestCropPages_OnEmptyImagePDF(t *testing.T) {
	tdir := t.TempDir()
	pngPath := filepath.Join(tdir, "empty.png")
	pdfPath := filepath.Join(tdir, "empty.pdf")

	img := makeWhiteImage(600, 800)
	writePNG(t, pngPath, img)
	createPDFViaImport(t, pngPath, pdfPath)

	opts := Options{DPI: 128, Threshold: 0.05, Space: 5, CropFrom: "center"}
	results, err := CropPages(pdfPath, nil, opts)
	if err != nil {
		t.Fatalf("CropPages (empty): %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected at least one page result")
	}
	res := results[0]
	if res.Crop == nil || res.Media == nil {
		t.Fatalf("missing crop or media boxes")
	}
	// Bounds sanity
	if res.Crop.LL.X < res.Media.LL.X || res.Crop.UR.X > res.Media.UR.X ||
		res.Crop.LL.Y < res.Media.LL.Y || res.Crop.UR.Y > res.Media.UR.Y {
		t.Errorf("crop out of media bounds: crop=%s media=%s", RectString(res.Crop), RectString(res.Media))
	}
	// Output existence
	outDir := filepath.Dir(res.Output)
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	found := false
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".pdf" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a single-page PDF in %s, found none", outDir)
	}
}

func TestCropPages_BorderModeEdgeContent(t *testing.T) {
	tdir := t.TempDir()
	pngPath := filepath.Join(tdir, "edge.png")
	pdfPath := filepath.Join(tdir, "edge.pdf")

	img := makeEdgeContentImage(800, 800)
	writePNG(t, pngPath, img)
	createPDFViaImport(t, pngPath, pdfPath)

	opts := Options{DPI: 128, Threshold: 0.02, Space: 5, CropFrom: "border"}
	results, err := CropPages(pdfPath, nil, opts)
	if err != nil {
		t.Fatalf("CropPages (border): %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected at least one page result")
	}
	res := results[0]
	if res.Crop == nil || res.Media == nil {
		t.Fatalf("missing crop or media boxes")
	}
	l, tFrac, r, b := rectFractions(res.Crop, res.Media)
	if !(l < 0.2 && tFrac > 0.8 && r > 0.8 && b < 0.2) {
		t.Errorf("unexpected fractions l=%.2f t=%.2f r=%.2f b=%.2f", l, tFrac, r, b)
	}
}

func TestCropPages_SpaceAffectsCropSize(t *testing.T) {
	tdir := t.TempDir()
	pngPath := filepath.Join(tdir, "content.png")
	pdfPath := filepath.Join(tdir, "content.pdf")

	img := makeTestImage(600, 800)
	writePNG(t, pngPath, img)
	createPDFViaImport(t, pngPath, pdfPath)

	// Small space
	resSmall, err := CropPages(pdfPath, nil, Options{DPI: 128, Threshold: 0.05, Space: 2, CropFrom: "center"})
	if err != nil {
		t.Fatalf("CropPages small space: %v", err)
	}
	// Larger space
	resLarge, err := CropPages(pdfPath, nil, Options{DPI: 128, Threshold: 0.05, Space: 25, CropFrom: "center"})
	if err != nil {
		t.Fatalf("CropPages large space: %v", err)
	}
	if len(resSmall) == 0 || len(resLarge) == 0 {
		t.Fatalf("expected results for both runs")
	}
	cs := resSmall[0].Crop
	cl := resLarge[0].Crop
	if cs == nil || cl == nil || resSmall[0].Media == nil {
		t.Fatalf("missing rectangles")
	}
	// Compare crop widths/heights
	wSmall := cs.UR.X - cs.LL.X
	hSmall := cs.UR.Y - cs.LL.Y
	wLarge := cl.UR.X - cl.LL.X
	hLarge := cl.UR.Y - cl.LL.Y
	if !(wLarge >= wSmall && hLarge >= hSmall) {
		t.Errorf("expected larger crop for bigger space: small(%.1fx%.1f) large(%.1fx%.1f)", wSmall, hSmall, wLarge, hLarge)
	}
}

func TestCropPages_WithExplicitRect(t *testing.T) {
	tdir := t.TempDir()
	pngPath := filepath.Join(tdir, "manual.png")
	pdfPath := filepath.Join(tdir, "manual.pdf")

	img := makeTestImage(600, 800)
	writePNG(t, pngPath, img)
	createPDFViaImport(t, pngPath, pdfPath)

	opts := Options{DPI: 128, Threshold: 0.05, Space: 5, CropFrom: "center"}
	po := []PageOption{{Number: 0, Left: 10, Top: 10, Right: 100, Bottom: 100}}
	results, err := CropPages(pdfPath, po, opts)
	if err != nil {
		t.Fatalf("CropPages explicit: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	res := results[0]
	if res.WasAuto {
		t.Fatalf("expected manual crop, got auto")
	}
	if res.Crop == nil || res.Media == nil {
		t.Fatalf("missing rectangles")
	}
	// Validate coordinates match expected conversion
	expected := rectFromTopLeft(res.Media, 10, 10, 100, 100)
	if int(res.Crop.LL.X) != int(expected.LL.X) || int(res.Crop.LL.Y) != int(expected.LL.Y) ||
		int(res.Crop.UR.X) != int(expected.UR.X) || int(res.Crop.UR.Y) != int(expected.UR.Y) {
		t.Errorf("unexpected crop: got %s want %s", RectString(res.Crop), RectString(expected))
	}
}

func TestCropPages_InvalidPageNumber(t *testing.T) {
	tdir := t.TempDir()
	pngPath := filepath.Join(tdir, "single.png")
	pdfPath := filepath.Join(tdir, "single.pdf")

	img := makeTestImage(500, 500)
	writePNG(t, pngPath, img)
	createPDFViaImport(t, pngPath, pdfPath)

	_, err := CropPages(pdfPath, []PageOption{{Number: 10}}, Options{DPI: 128, Threshold: 0.05, Space: 5, CropFrom: "center"})
	if err == nil {
		t.Fatalf("expected error for invalid page number")
	}
}

func TestCropDocument_WritesOutput(t *testing.T) {
	tdir := t.TempDir()
	pngPath := filepath.Join(tdir, "doc.png")
	pdfPath := filepath.Join(tdir, "doc.pdf")
	outPath := filepath.Join(tdir, "doc_out.pdf")

	img := makeTestImage(400, 400)
	writePNG(t, pngPath, img)
	createPDFViaImport(t, pngPath, pdfPath)

	if err := CropDocument(pdfPath, outPath, Options{DPI: 128, Threshold: 0.05, Space: 5, CropFrom: "center"}); err != nil {
		t.Fatalf("CropDocument: %v", err)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file written: %v", err)
	}
}

func TestCropAllPagesToSingleFile_MultiPage(t *testing.T) {
	tdir := t.TempDir()
	p1 := filepath.Join(tdir, "p1.png")
	p2 := filepath.Join(tdir, "p2.png")
	pdfPath := filepath.Join(tdir, "multi.pdf")
	outPath := filepath.Join(tdir, "multi_out.pdf")

	writePNG(t, p1, makeTestImage(400, 600))
	writePNG(t, p2, makeEdgeContentImage(800, 800))
	createMultiPagePDFViaImport(t, []string{p1, p2}, pdfPath)

	opts := Options{DPI: 128, Threshold: 0.05, Space: 5, CropFrom: "center"}
	results, err := CropAllPagesToSingleFile(pdfPath, outPath, opts)
	if err != nil {
		t.Fatalf("CropAllPagesToSingleFile (multi): %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file written: %v", err)
	}
	for _, r := range results {
		if r.Crop == nil || r.Media == nil {
			t.Errorf("missing rectangles for page %d", r.PageNo)
		}
		if r.Output != outPath {
			t.Errorf("expected Output to equal outPath")
		}
	}
}

func TestCropPages_MixedManualAuto_MultiPage(t *testing.T) {
	tdir := t.TempDir()
	p1 := filepath.Join(tdir, "auto.png")
	p2 := filepath.Join(tdir, "manual.png")
	pdfPath := filepath.Join(tdir, "mixed.pdf")

	writePNG(t, p1, makeTestImage(500, 700))
	writePNG(t, p2, makeTestImage(400, 400))
	createMultiPagePDFViaImport(t, []string{p1, p2}, pdfPath)

	// Page 0 auto, Page 1 manual rect
	pageOpts := []PageOption{
		{Number: 0},
		{Number: 1, Left: 10, Top: 10, Right: 120, Bottom: 120},
	}
	results, err := CropPages(pdfPath, pageOpts, Options{DPI: 128, Threshold: 0.05, Space: 5, CropFrom: "center"})
	if err != nil {
		t.Fatalf("CropPages mixed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].WasAuto {
		t.Errorf("expected page 0 auto")
	}
	if results[1].WasAuto {
		t.Errorf("expected page 1 manual")
	}
	// Ensure outputs exist (pdfcpu may adjust naming); check any PDF in output dir
	for _, r := range results {
		outDir := filepath.Dir(r.Output)
		entries, err := os.ReadDir(outDir)
		if err != nil {
			t.Fatalf("read outDir for page %d: %v", r.PageNo, err)
		}
		found := false
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".pdf" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected a PDF output in %s for page %d", outDir, r.PageNo)
		}
	}
}

func TestCropDocument_ErrorOnMissingOutput(t *testing.T) {
	tdir := t.TempDir()
	pngPath := filepath.Join(tdir, "e.png")
	pdfPath := filepath.Join(tdir, "e.pdf")

	writePNG(t, pngPath, makeTestImage(300, 300))
	createPDFViaImport(t, pngPath, pdfPath)

	if err := CropDocument(pdfPath, "", Options{DPI: 128, Threshold: 0.05, Space: 5, CropFrom: "center"}); err == nil {
		t.Fatalf("expected error for empty output file")
	}
}

func TestPageMediaBox_FallbackA4(t *testing.T) {
	tdir := t.TempDir()
	pngPath := filepath.Join(tdir, "a4.png")
	pdfPath := filepath.Join(tdir, "a4.pdf")

	writePNG(t, pngPath, makeTestImage(300, 300))
	createPDFViaImport(t, pngPath, pdfPath)

	ctx, err := api.ReadContextFile(pdfPath)
	if err != nil {
		t.Fatalf("read ctx: %v", err)
	}
	// Query a non-existent page number to trigger fallback
	rect, err := pageMediaBox(ctx, 99)
	if err != nil {
		t.Fatalf("pageMediaBox error: %v", err)
	}
	if int(rect.UR.X-rect.LL.X) != 595 || int(rect.UR.Y-rect.LL.Y) != 842 {
		t.Fatalf("expected A4 size 595x842, got %dx%d", int(rect.UR.X-rect.LL.X), int(rect.UR.Y-rect.LL.Y))
	}
}
