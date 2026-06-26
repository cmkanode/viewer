package imgutils

import (
	"bytes"
	"image"
)

func IsValidImage(data []byte) bool {
	_, _, err := image.Decode(bytes.NewReader(data))
	return err == nil
}
