package ops

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// makePNG builds a simple w x h gradient image encoded as PNG.
func makePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / max(w-1, 1)),
				G: uint8(y * 255 / max(h-1, 1)),
				B: 128,
				A: 255,
			})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test png: %v", err)
	}
	return buf.Bytes()
}

func decode(t *testing.T, data []byte) image.Image {
	t.Helper()
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode result: %v", err)
	}
	return img
}

func TestInfo(t *testing.T) {
	data := makePNG(t, 40, 25)
	info, err := Info(data)
	if err != nil {
		t.Fatalf("Info error: %v", err)
	}
	if info.Format != "png" || info.Width != 40 || info.Height != 25 {
		t.Errorf("Info = %+v, want png 40x25", info)
	}
}

func TestResizeDimensions(t *testing.T) {
	data := makePNG(t, 100, 50)
	out, err := Resize(data, 20, 10, "png")
	if err != nil {
		t.Fatalf("Resize error: %v", err)
	}
	img := decode(t, out)
	if img.Bounds().Dx() != 20 || img.Bounds().Dy() != 10 {
		t.Errorf("resized bounds = %v, want 20x10", img.Bounds())
	}
}

func TestResizeAspectRatio(t *testing.T) {
	data := makePNG(t, 200, 100)
	// Only width given: height should follow the 2:1 aspect ratio.
	out, err := Resize(data, 50, 0, "png")
	if err != nil {
		t.Fatalf("Resize error: %v", err)
	}
	img := decode(t, out)
	if img.Bounds().Dx() != 50 || img.Bounds().Dy() != 25 {
		t.Errorf("resized bounds = %v, want 50x25", img.Bounds())
	}
}

func TestResizeJPEGOutput(t *testing.T) {
	data := makePNG(t, 32, 32)
	out, err := Resize(data, 16, 16, "jpeg")
	if err != nil {
		t.Fatalf("Resize jpeg error: %v", err)
	}
	info, err := Info(out)
	if err != nil {
		t.Fatalf("Info on jpeg error: %v", err)
	}
	if info.Format != "jpeg" {
		t.Errorf("output format = %q, want jpeg", info.Format)
	}
}

func TestResizeInvalidDimensions(t *testing.T) {
	data := makePNG(t, 10, 10)
	if _, err := Resize(data, 0, 0, "png"); err == nil {
		t.Fatal("expected error when both dimensions are zero")
	}
}

func TestGrayscale(t *testing.T) {
	data := makePNG(t, 30, 30)
	out, err := Grayscale(data, "png")
	if err != nil {
		t.Fatalf("Grayscale error: %v", err)
	}
	img := decode(t, out)
	// Every pixel in a grayscale image has equal R, G and B channels.
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r != g || g != b {
				t.Fatalf("pixel (%d,%d) not gray: r=%d g=%d b=%d", x, y, r, g, b)
			}
		}
	}
}

func TestDecodeFailure(t *testing.T) {
	if _, err := Info([]byte("not an image")); err == nil {
		t.Fatal("expected error decoding invalid image data")
	}
}

func BenchmarkResize(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		b.Fatal(err)
	}
	data := buf.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Resize(data, 128, 128, "png"); err != nil {
			b.Fatal(err)
		}
	}
}
