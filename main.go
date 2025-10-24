package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"syscall"
	"unsafe"

	"github.com/mattn/go-sixel"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getTerminalSize() (int, int) {
	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: crt <image_file> [mode] [--center]")
		fmt.Println("\nSupported formats: JPEG, PNG, BMP, TIFF")
		fmt.Println("Note: GIF and WebP show first frame only (no animation)")
		fmt.Println("\nOptional mode:")
		fmt.Println("  block - Use block characters (default, works everywhere)")
		fmt.Println("  sixel - Use Sixel graphics (highest quality, requires compatible terminal)")
		fmt.Println("\nOptional flags:")
		fmt.Println("  --center - Center the image horizontally")
		os.Exit(1)
	}

	filePath := os.Args[1]
	mode := "block"
	center := false

	// Parse arguments
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--center" {
			center = true
		} else if arg == "block" || arg == "sixel" {
			mode = arg
		}
	}

	// Open image file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if mode == "sixel" {
		renderSixel(img, center)
	} else {
		renderBlock(img, center)
	}
}

func renderSixel(img image.Image, center bool) {
	// Get terminal size
	termWidth, termHeight := getTerminalSize()

	// Get image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Calculate image aspect ratio
	imgAspect := float64(imgWidth) / float64(imgHeight)

	// Estimate terminal pixel dimensions
	// Sixel uses 6 pixels per character height (one sixel band)
	charPixelWidth := 8
	charPixelHeight := 6 // Sixel band height

	maxPixelWidth := (termWidth - 1) * charPixelWidth
	maxPixelHeight := (termHeight - 4) * charPixelHeight // Leave space to avoid cropping

	var targetWidth, targetHeight int

	// Fit by width first
	targetWidth = maxPixelWidth
	targetHeight = int(float64(targetWidth) / imgAspect)

	// If height exceeds, fit by height instead
	if targetHeight > maxPixelHeight {
		targetHeight = maxPixelHeight
		targetWidth = int(float64(targetHeight) * imgAspect)
	}

	// Ensure non-zero dimensions
	if targetWidth < 1 {
		targetWidth = 1
	}
	if targetHeight < 1 {
		targetHeight = 1
	}

	// Resize image
	resized := resize.Resize(uint(targetWidth), uint(targetHeight), img, resize.Lanczos3)

	// Handle transparency: Sixel doesn't support alpha channel
	// Solution: fill transparent areas with terminal background color
	processedImg := handleTransparency(resized)

	// If centering is enabled, move cursor using ANSI escape sequence
	if center {
		actualCharWidth := (targetWidth + charPixelWidth - 1) / charPixelWidth
		leftPadding := (termWidth - actualCharWidth) / 2
		if leftPadding > 0 {
			// Move cursor right using ANSI escape sequence
			fmt.Printf("\033[%dC", leftPadding)
		}
	}

	// Output using go-sixel library
	enc := sixel.NewEncoder(os.Stdout)
	enc.Dither = true

	if err := enc.Encode(processedImg); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding sixel: %v\n", err)
		os.Exit(1)
	}

	// Print newline to avoid zsh showing %
	fmt.Println()
}

// handleTransparency processes transparent areas in the image
// Sixel doesn't support transparency, so we fill transparent areas with terminal background color
func handleTransparency(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	// Check if image has alpha channel
	hasAlpha := false
	for y := bounds.Min.Y; y < bounds.Max.Y && !hasAlpha; y++ {
		for x := bounds.Min.X; x < bounds.Max.X && !hasAlpha; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 65535 {
				hasAlpha = true
			}
		}
	}

	// If no alpha channel, return original image
	if !hasAlpha {
		return img
	}

	// Get terminal background color (default: black)
	bgR, bgG, bgB := uint8(0), uint8(0), uint8(0)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			if a == 0 {
				// Fully transparent, use background color
				newImg.SetRGBA(x, y, color.RGBA{R: bgR, G: bgG, B: bgB, A: 255})
			} else if a < 65535 {
				// Semi-transparent, blend with background color
				alpha := float64(a) / 65535.0

				r8 := uint8((float64(r>>8)*alpha + float64(bgR)*(1-alpha)))
				g8 := uint8((float64(g>>8)*alpha + float64(bgG)*(1-alpha)))
				b8 := uint8((float64(b>>8)*alpha + float64(bgB)*(1-alpha)))

				newImg.SetRGBA(x, y, color.RGBA{R: r8, G: g8, B: b8, A: 255})
			} else {
				// Opaque, use original color
				r8 := uint8(r >> 8)
				g8 := uint8(g >> 8)
				b8 := uint8(b >> 8)
				newImg.SetRGBA(x, y, color.RGBA{R: r8, G: g8, B: b8, A: 255})
			}
		}
	}

	return newImg
}

func renderBlock(img image.Image, center bool) {
	// Get terminal size
	termWidth, termHeight := getTerminalSize()

	// Get image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Calculate image aspect ratio
	imgAspect := float64(imgWidth) / float64(imgHeight)

	// Calculate display size in characters
	maxCharWidth := termWidth - 1
	maxCharHeight := termHeight - 2

	var charWidth, charHeight int

	// Terminal characters have roughly 1:2 aspect ratio (width:height)
	// Fit by width first
	charWidth = maxCharWidth
	charHeight = int(float64(charWidth) / (imgAspect * 2.0))

	// If height exceeds, fit by height instead
	if charHeight > maxCharHeight {
		charHeight = maxCharHeight
		charWidth = int(float64(charHeight) * imgAspect * 2.0)
	}

	// Calculate left padding for centering
	leftPadding := 0
	if center {
		leftPadding = (termWidth - charWidth) / 2
		if leftPadding < 0 {
			leftPadding = 0
		}
	}

	// Use half-block characters, each character displays 2 pixels vertically
	pixelWidth := charWidth
	pixelHeight := charHeight * 2

	// Resize image
	resized := resize.Resize(uint(pixelWidth), uint(pixelHeight), img, resize.Lanczos3)

	// Render using half-block character ▀
	for row := 0; row < charHeight; row++ {
		// Output left padding
		if leftPadding > 0 {
			for i := 0; i < leftPadding; i++ {
				fmt.Print(" ")
			}
		}

		for col := 0; col < charWidth; col++ {
			// Top half pixel
			topY := row * 2
			topR, topG, topB, topA := resized.At(col, topY).RGBA()

			// Bottom half pixel
			bottomY := row*2 + 1
			bottomR, bottomG, bottomB, bottomA := resized.At(col, bottomY).RGBA()

			// Convert to 8-bit color
			tr := uint8(topR >> 8)
			tg := uint8(topG >> 8)
			tb := uint8(topB >> 8)
			ta := uint8(topA >> 8)

			br := uint8(bottomR >> 8)
			bg := uint8(bottomG >> 8)
			bb := uint8(bottomB >> 8)
			ba := uint8(bottomA >> 8)

			// If both pixels are fully transparent, skip (show terminal background)
			if ta == 0 && ba == 0 {
				fmt.Print(" ")
				continue
			}

			// If top half is transparent, use space with background color
			if ta == 0 {
				// Only show bottom half
				fmt.Printf("\033[48;2;%d;%d;%dm \033[0m", br, bg, bb)
			} else if ba == 0 {
				// Only show top half
				fmt.Printf("\033[38;2;%d;%d;%dm▀\033[0m", tr, tg, tb)
			} else {
				// Both opaque, display normally
				fmt.Printf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm▀\033[0m",
					tr, tg, tb, br, bg, bb)
			}
		}
		fmt.Println()
	}
}
