package utils

import (
	"image"
	"image/color"
)

// Used GPT-4o
func BilinearResize(img image.Image, newWidth, newHeight int) image.Image {
	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	scaleX := float64(srcWidth) / float64(newWidth)
	scaleY := float64(srcHeight) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := float64(x) * scaleX
			srcY := float64(y) * scaleY

			x1 := int(srcX)
			y1 := int(srcY)
			x2 := min(x1+1, srcWidth-1)
			y2 := min(y1+1, srcHeight-1)

			r1, g1, b1, a1 := getInterpolatedColor(img, x1, y1, x2, y1, srcX-float64(x1))
			r2, g2, b2, a2 := getInterpolatedColor(img, x1, y2, x2, y2, srcX-float64(x1))

			r := r1*(1-(srcY-float64(y1))) + r2*(srcY-float64(y1))
			g := g1*(1-(srcY-float64(y1))) + g2*(srcY-float64(y1))
			b := b1*(1-(srcY-float64(y1))) + b2*(srcY-float64(y1))
			a := a1*(1-(srcY-float64(y1))) + a2*(srcY-float64(y1))

			dst.Set(x, y, color.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: uint8(a),
			})
		}
	}
	return dst
}

func getInterpolatedColor(img image.Image, x1, y1, x2, y2 int, weight float64) (float64, float64, float64, float64) {
	r1, g1, b1, a1 := img.At(x1, y1).RGBA()
	r2, g2, b2, a2 := img.At(x2, y2).RGBA()

	r := float64(r1>>8)*(1-weight) + float64(r2>>8)*weight
	g := float64(g1>>8)*(1-weight) + float64(g2>>8)*weight
	b := float64(b1>>8)*(1-weight) + float64(b2>>8)*weight
	a := float64(a1>>8)*(1-weight) + float64(a2>>8)*weight

	return r, g, b, a
}
