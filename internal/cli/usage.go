package cli

func PdfCropUsage() string {
	return "pdf_crop - Crop PDF pages using raster detection\n\n" +
		"Usage:\n" +
		"  pdf_crop -i <input.pdf> [--threshold <float>] [--space <int>] [--dpi <float>]\n" +
		"  pdf_crop -i <input.pdf> -p <page> <left> <top> <right> <bottom> <out.pdf> [repeatable]\n\n" +
		"Options:\n" +
		"  -i, --input_file    Path to input PDF (required)\n" +
		"  -p, --page          Per-page crop + output: page left top right bottom out.pdf (can repeat)\n" +
		"      --threshold      Detection threshold (default: 0.008)\n" +
		"      --space          Extra whitespace in points (default: 5)\n" +
		"      --dpi            Rasterization DPI (default: 128)\n" +
		"  -h, --help          Show this help and exit\n"
}

func CropAllPdfUsage() string {
	return "crop_all_pdf - Crop all PDFs in a directory\n\n" +
		"Usage:\n" +
		"  crop_all_pdf --dir <path> [--threshold <float>] [--space <int>] [--dpi <float>]\n\n" +
		"Options:\n" +
		"  -d, --dir           Directory containing PDFs (default: current directory)\n" +
		"      --threshold      Detection threshold (default: 0.1)\n" +
		"      --space          Extra whitespace in points (default: 5)\n" +
		"      --dpi            Rasterization DPI (default: 128)\n" +
		"  -h, --help          Show this help and exit\n"
}
