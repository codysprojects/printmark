package pdfrender

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"
	_ "golang.org/x/image/webp"
)

// assumedImageDPI converts an image's pixel dimensions to a physical
// size on the page. Go's stdlib image decoders don't expose embedded
// DPI metadata, so we assume a common default rather than guessing.
const assumedImageDPI = 96.0

const mmPerInch = 25.4

// renderImage embeds a local image referenced by markdown's
// ![alt](path) syntax, scaled to fit the page width and treated as its
// own block (text around it breaks to a new line rather than flowing
// beside it). Falls back to the alt text if the image can't be used: a
// remote http(s) URL, a missing file, or data that isn't
// PNG/JPEG/GIF/WebP.
func (r *renderer) renderImage(node *ast.Image) {
	dest := string(node.Destination)
	alt := nodeText(node, r.source)

	data, ok := r.loadLocalImage(dest)
	if !ok {
		if alt != "" {
			r.write(alt)
		}
		return
	}

	kind, embedData, widthPx, heightPx, ok := decodeImageInfo(data)
	if !ok {
		if alt != "" {
			r.write(alt)
		}
		return
	}

	pdf := r.pdf
	widthMM := float64(widthPx) / assumedImageDPI * mmPerInch
	heightMM := float64(heightPx) / assumedImageDPI * mmPerInch

	left, _, right, bottom := pdf.GetMargins()
	pageW, pageH := pdf.GetPageSize()
	availWidth := pageW - left - right

	if widthMM > availWidth {
		heightMM *= availWidth / widthMM
		widthMM = availWidth
	}

	x, y := pdf.GetXY()
	if x > left {
		// Mid-line: drop to a fresh line before placing the image.
		pdf.Ln(r.cfg.LineHeight)
		x, y = pdf.GetXY()
	}
	if y+heightMM > pageH-bottom {
		pdf.AddPage()
		x, y = pdf.GetXY()
	}

	opts := fpdf.ImageOptions{ImageType: kind}
	pdf.RegisterImageOptionsReader(dest, opts, bytes.NewReader(embedData))
	pdf.ImageOptions(dest, x, y, widthMM, heightMM, false, opts, 0, "")
	pdf.SetXY(x, y+heightMM)
	pdf.Ln(r.cfg.LineHeight)
}

// loadLocalImage reads a local image file, resolved relative to the
// markdown file's directory if it isn't absolute. Remote http(s) images
// are out of scope for now.
func (r *renderer) loadLocalImage(dest string) ([]byte, bool) {
	if strings.HasPrefix(dest, "http://") || strings.HasPrefix(dest, "https://") {
		return nil, false
	}
	path := dest
	if !filepath.IsAbs(path) {
		path = filepath.Join(r.baseDir, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

// decodeImageInfo validates that data decodes as PNG/JPEG/GIF/WebP,
// using Go's stdlib decoders directly (rather than fpdf's) so a corrupt
// image can't leave fpdf's own internal error state set for the rest of
// the document. It returns the pixel dimensions and the bytes to
// actually hand to fpdf for embedding: fpdf has no native WebP support,
// so a WebP image is fully decoded and re-encoded as PNG here first.
func decodeImageInfo(data []byte) (kind string, embedData []byte, widthPx, heightPx int, ok bool) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", nil, 0, 0, false
	}
	switch format {
	case "png":
		return "PNG", data, cfg.Width, cfg.Height, true
	case "jpeg":
		return "JPG", data, cfg.Width, cfg.Height, true
	case "gif":
		return "GIF", data, cfg.Width, cfg.Height, true
	case "webp":
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return "", nil, 0, 0, false
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return "", nil, 0, 0, false
		}
		return "PNG", buf.Bytes(), cfg.Width, cfg.Height, true
	default:
		return "", nil, 0, 0, false
	}
}
