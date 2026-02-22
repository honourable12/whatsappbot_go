package utils

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"

	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
)

// ConvertToSticker converts an image (JPEG/PNG/GIF) to a 512x512 WebP sticker.
func ConvertToSticker(r io.Reader) ([]byte, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	// WhatsApp stickers are 512x512
	newImg := resize.Thumbnail(512, 512, img, resize.Lanczos3)

	var buf bytes.Buffer
	err = webp.Encode(&buf, newImg, &webp.Options{Lossless: false, Quality: 80})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ConvertToImage converts a WebP sticker back to a PNG.
func ConvertToImage(r io.Reader) ([]byte, error) {
	img, err := webp.Decode(r)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
