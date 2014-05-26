package util

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"gopkg.in/qml.v0"
)

// Save an image from CanvasImageData
func SaveCanvasImage(name string, width, height int, data *qml.Map) (err error) {
	tmp := make(map[string]color.RGBA)
	data.Convert(&tmp)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for sid, px := range tmp {
		i := digitsToNum(sid)
		img.SetRGBA(i%width, i/width, px)
	}

	f, err := os.Create(name)
	if err != nil {
		return
	}
	defer f.Close()
	if err = png.Encode(f, img); err != nil {
		return
	}
	return
}

func digitsToNum(s string) (n int) {
	if l := len(s); l > 0 {
		n += int(s[l-1] - 48)
		k := int(l - 1)
		for _, r := range s[:l-1] {
			n, k = n+int(r-48)*int(math.Pow10(k)), k-1
		}
	}
	return
}
