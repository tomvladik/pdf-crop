package crop

import "image"

type detectData struct {
	width     int
	height    int
	prefixSum []int
	rowCounts []int
	colCounts []int
}

func buildDetectData(img *image.RGBA) detectData {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	ps := make([]int, (width+1)*(height+1))
	rowCounts := make([]int, height)
	colCounts := make([]int, width)

	for y := 0; y < height; y++ {
		rowSum := 0
		rowOffset := y * img.Stride
		for x := 0; x < width; x++ {
			idx := rowOffset + x*4
			r := img.Pix[idx]
			g := img.Pix[idx+1]
			b := img.Pix[idx+2]
			a := img.Pix[idx+3]
			nonWhite := r != 255 || g != 255 || b != 255 || a != 255
			if nonWhite {
				rowSum++
				rowCounts[y]++
				colCounts[x]++
			}
			ps[(y+1)*(width+1)+x+1] = ps[y*(width+1)+x+1] + rowSum
		}
	}

	return detectData{
		width:     width,
		height:    height,
		prefixSum: ps,
		rowCounts: rowCounts,
		colCounts: colCounts,
	}
}

func (d detectData) countNonZero(x0, y0, x1, y1 int) int {
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > d.width {
		x1 = d.width
	}
	if y1 > d.height {
		y1 = d.height
	}
	if x0 >= x1 || y0 >= y1 {
		return 0
	}

	w := d.width + 1
	return d.prefixSum[y1*w+x1] - d.prefixSum[y1*w+x0] - d.prefixSum[y0*w+x1] + d.prefixSum[y0*w+x0]
}

func detectCenter(d detectData) (int, int) {
	centerX := -1
	centerY := -1
	nonZero := -1
	for i := 0; i < d.height; i++ {
		nnz := d.rowCounts[i]
		if nnz > nonZero {
			centerY = i
			nonZero = nnz
		}
	}

	nonZero = -1
	for j := 0; j < d.width; j++ {
		nnz := d.colCounts[j]
		if nnz > nonZero {
			centerX = j
			nonZero = nnz
		}
	}
	return centerX, centerY
}

func detectTop(d detectData, cy, space, threshold int) int {
	for endY := cy; endY > 0; endY -= space {
		startY := endY - threshold
		if startY < 0 {
			startY = 0
		}
		if d.countNonZero(0, startY, d.width, endY) == 0 {
			return endY
		}
	}
	return 0
}

func detectBottom(d detectData, cy, space, threshold int) int {
	for startY := cy; startY < d.height; startY += space {
		endY := startY + threshold
		if endY > d.height {
			endY = d.height
		}
		if d.countNonZero(0, startY, d.width, endY) == 0 {
			return startY
		}
	}
	return d.height
}

func detectLeft(d detectData, cx, top, bottom, space, threshold int) int {
	for endX := cx; endX > 0; endX -= space {
		startX := endX - threshold
		if startX < 0 {
			startX = 0
		}
		if d.countNonZero(startX, top, endX, bottom) == 0 {
			return endX
		}
	}
	return 0
}

func detectRight(d detectData, cx, top, bottom, space, threshold int) int {
	for startX := cx; startX < d.width; startX += space {
		endX := startX + threshold
		if endX > d.width {
			endX = d.width
		}
		if d.countNonZero(startX, top, endX, bottom) == 0 {
			return startX
		}
	}
	return d.width
}

func detectBorder(d detectData, space, thresholdW, thresholdH int) (int, int, int, int) {
	top, right, bottom, left := 0, 0, 0, 0
	for startY := 0; startY < d.height; startY += space {
		endY := startY + thresholdH
		if endY > d.height {
			endY = d.height
		}
		if d.countNonZero(0, startY, d.width, endY) != 0 {
			top = startY
			break
		}
	}

	for endY := d.height; endY > top; endY -= space {
		startY := endY - thresholdH
		if startY < 0 {
			startY = 0
		}
		if d.countNonZero(0, startY, d.width, endY) != 0 {
			bottom = endY
			break
		}
	}

	for startX := 0; startX < d.width; startX += space {
		endX := startX + thresholdW
		if endX > d.width {
			endX = d.width
		}
		if d.countNonZero(startX, 0, endX, d.height) != 0 {
			left = startX
			break
		}
	}

	for endX := d.width; endX > left; endX -= space {
		startX := endX - thresholdW
		if startX < 0 {
			startX = 0
		}
		if d.countNonZero(startX, 0, endX, d.height) != 0 {
			right = endX
			break
		}
	}

	return top, bottom, left, right
}

func detectFrame(img *image.RGBA, space int, threshold float64, cropFrom string) (float64, float64, float64, float64) {
	d := buildDetectData(img)
	if d.width == 0 || d.height == 0 {
		return 0, 0, 0, 0
	}

	if cropFrom == "center" {
		thresholdH := int(float64(d.height) * threshold)
		thresholdW := int(float64(d.width) * threshold)
		if thresholdH < 1 {
			thresholdH = 1
		}
		if thresholdW < 1 {
			thresholdW = 1
		}
		cx, cy := detectCenter(d)
		top := detectTop(d, cy, space, thresholdH)
		bottom := detectBottom(d, cy, space, thresholdH)
		left := detectLeft(d, cx, top, bottom, space, thresholdW)
		right := detectRight(d, cx, top, bottom, space, thresholdW)
		return float64(left) / float64(d.width), float64(top) / float64(d.height), float64(right) / float64(d.width), float64(bottom) / float64(d.height)
	}

	thresholdH := int(float64(d.height) * threshold)
	thresholdW := int(float64(d.width) * threshold)
	if thresholdH < 1 {
		thresholdH = 1
	}
	if thresholdW < 1 {
		thresholdW = 1
	}
	if thresholdH > space {
		thresholdH = space
	}
	if thresholdW > space {
		thresholdW = space
	}

	top, bottom, left, right := detectBorder(d, space, thresholdW, thresholdH)
	return float64(left) / float64(d.width), float64(top) / float64(d.height), float64(right) / float64(d.width), float64(bottom) / float64(d.height)
}
