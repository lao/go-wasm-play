package ops

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"

	// Register the standard decoders so image.Decode recognises these formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// ImageInfo describes a decoded image.
type ImageInfo struct {
	Format string `json:"format"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Info decodes just enough of data to report the image's format and size.
func Info(data []byte) (ImageInfo, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return ImageInfo{}, fmt.Errorf("decode image: %w", err)
	}
	return ImageInfo{Format: format, Width: cfg.Width, Height: cfg.Height}, nil
}

// encode serialises img using the requested format ("png" or "jpeg"). An empty
// format defaults to png.
func encode(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer
	switch format {
	case "", "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported image format %q", format)
	}
	return buf.Bytes(), nil
}

// Resize decodes data, scales it to width x height using bilinear
// interpolation and re-encodes it in the requested output format. If either
// dimension is <= 0 it is derived from the other to preserve the aspect ratio.
func Resize(data []byte, width, height int, format string) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	width, height, err = targetSize(src.Bounds(), width, height)
	if err != nil {
		return nil, err
	}
	return encode(bilinearResize(src, width, height), format)
}

// Grayscale decodes data, converts it to grayscale and re-encodes it.
func Grayscale(data []byte, format string) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	bounds := src.Bounds()
	gray := image.NewGray(bounds)
	// Drawing onto an *image.Gray converts every pixel through the Gray colour
	// model, which applies the standard luminance weighting.
	draw.Draw(gray, bounds, src, bounds.Min, draw.Src)
	return encode(gray, format)
}

// targetSize resolves the requested dimensions, filling in a missing one from
// the source aspect ratio.
func targetSize(bounds image.Rectangle, width, height int) (int, int, error) {
	sw, sh := bounds.Dx(), bounds.Dy()
	if sw <= 0 || sh <= 0 {
		return 0, 0, fmt.Errorf("source image has no pixels")
	}
	switch {
	case width <= 0 && height <= 0:
		return 0, 0, fmt.Errorf("at least one of width or height must be positive")
	case width <= 0:
		width = int(float64(height) * float64(sw) / float64(sh))
	case height <= 0:
		height = int(float64(width) * float64(sh) / float64(sw))
	}
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	return width, height, nil
}

// bilinearResize scales src to dstW x dstH using bilinear interpolation. It
// works for any image.Image and produces an *image.RGBA.
func bilinearResize(src image.Image, dstW, dstH int) *image.RGBA {
	bounds := src.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	// Ratio between source and destination pixel grids.
	xRatio := float64(srcW) / float64(dstW)
	yRatio := float64(srcH) / float64(dstH)

	for y := 0; y < dstH; y++ {
		// Map the destination pixel centre back into source space.
		sy := (float64(y)+0.5)*yRatio - 0.5
		y0 := int(sy)
		fy := sy - float64(y0)
		for x := 0; x < dstW; x++ {
			sx := (float64(x)+0.5)*xRatio - 0.5
			x0 := int(sx)
			fx := sx - float64(x0)

			c00 := sampleRGBA(src, bounds, x0, y0)
			c10 := sampleRGBA(src, bounds, x0+1, y0)
			c01 := sampleRGBA(src, bounds, x0, y0+1)
			c11 := sampleRGBA(src, bounds, x0+1, y0+1)

			dst.Set(x, y, color.RGBA{
				R: lerp2(c00.R, c10.R, c01.R, c11.R, fx, fy),
				G: lerp2(c00.G, c10.G, c01.G, c11.G, fx, fy),
				B: lerp2(c00.B, c10.B, c01.B, c11.B, fx, fy),
				A: lerp2(c00.A, c10.A, c01.A, c11.A, fx, fy),
			})
		}
	}
	return dst
}

// sampleRGBA returns the 8-bit RGBA colour at (x, y), clamping coordinates to
// the image bounds so edge pixels are repeated instead of wrapping.
func sampleRGBA(src image.Image, bounds image.Rectangle, x, y int) color.RGBA {
	if x < 0 {
		x = 0
	} else if x >= bounds.Dx() {
		x = bounds.Dx() - 1
	}
	if y < 0 {
		y = 0
	} else if y >= bounds.Dy() {
		y = bounds.Dy() - 1
	}
	r, g, b, a := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
	// RGBA() returns 16-bit values; shift down to 8 bits.
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

// lerp2 performs bilinear interpolation of four corner samples.
func lerp2(c00, c10, c01, c11 uint8, fx, fy float64) uint8 {
	top := float64(c00)*(1-fx) + float64(c10)*fx
	bottom := float64(c01)*(1-fx) + float64(c11)*fx
	v := top*(1-fy) + bottom*fy
	if v < 0 {
		v = 0
	} else if v > 255 {
		v = 255
	}
	return uint8(v + 0.5)
}
