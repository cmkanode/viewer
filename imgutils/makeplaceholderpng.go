package imgutils

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

func MakePlaceholderPNG() []byte {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.NRGBA{R: 240, G: 240, B: 240, A: 255})
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
