# CRT - Terminal Image Viewer

A high-quality command-line image viewer for displaying images directly in your terminal using Unicode block characters or Sixel graphics.

## Features

- 🖼️ **Multiple Display Modes**
  - **Block mode**: Uses Unicode half-block characters (▀) for universal compatibility
  - **Sixel mode**: High-fidelity graphics for compatible terminals
  
- 🎨 **Rich Format Support**
  - JPEG, PNG, GIF, BMP, TIFF, WebP
  - Transparent PNG support (alpha channel handling)
  
- 📐 **Smart Rendering**
  - Automatic aspect ratio preservation
  - Adaptive sizing to fit terminal dimensions
  - Optional horizontal centering
  - No distortion or cropping
  
- 🌈 **True Color**
  - 24-bit RGB color support
  - High-quality image resizing with Lanczos3 algorithm
  - Color dithering for Sixel mode

## Installation

### Prerequisites

- Go 1.22.0 or higher

### Build from Source

```bash
git clone https://github.com/rzh/crt.git
cd crt
go build -o crt
```

## Usage

### Basic Usage

```bash
# Display image using block characters (default)
./crt image.png

# Display image using Sixel graphics
./crt image.png sixel

# Center the image horizontally
./crt image.png --center

# Combine Sixel mode with centering
./crt image.png sixel --center
```

### Display Modes

#### Block Mode (Default)

Uses Unicode half-block characters (▀) to display images with 2 colors per character cell. Works in all terminals with Unicode support.

**Features:**
- Universal compatibility
- Transparent background shows terminal background
- Resolution: ~2x vertical pixel density

**Example:**
```bash
./crt photo.jpg --center
```
<img width="710" height="371" alt="image-block" src="https://github.com/user-attachments/assets/5b6c23cb-f8c5-4dd9-8ad0-10c195a9cc52" />

#### Sixel Mode

Uses Sixel graphics protocol for high-fidelity image display. Requires a Sixel-compatible terminal.

**Features:**
- Highest quality rendering
- True pixel-level graphics
- Color dithering for better gradients
- Transparent areas filled with black (for dark terminals)

**Compatible Terminals:**
- iTerm2 (macOS)
- WezTerm
- Mintty (Windows)
- mlterm
- xterm (with Sixel support)

**Example:**
```bash
./crt photo.jpg sixel --center
```
<img width="638" height="215" alt="image-sixel" src="https://github.com/user-attachments/assets/4893322d-fb3f-41e4-a857-de092bfadd49" />

### Options

| Option | Description |
|--------|-------------|
| `block` | Use Unicode block characters (default) |
| `sixel` | Use Sixel graphics protocol |
| `--center` | Center the image horizontally |

### Supported Image Formats

- **JPEG** (.jpg, .jpeg)
- **PNG** (.png) - with transparency support
- **BMP** (.bmp)
- **TIFF** (.tiff, .tif)
- **GIF** (.gif) - first frame only, no animation
- **WebP** (.webp) - static images only, no animation

**Note:** This is a static image viewer. Animated GIF and WebP files will only display their first frame.

## Examples

```bash
# View a photo
./crt vacation.jpg

# View a transparent PNG with centering
./crt logo.png --center

# High-quality Sixel rendering
./crt artwork.png sixel

# WebP image in block mode
./crt animation.webp block
```

## How It Works

### Block Mode

1. Resizes the image to fit terminal dimensions while maintaining aspect ratio
2. Converts each pair of vertical pixels into a single terminal character
3. Uses the half-block character (▀) with:
   - Foreground color = top pixel color
   - Background color = bottom pixel color
4. Handles transparency by showing terminal background for transparent pixels

### Sixel Mode

1. Resizes the image to fit terminal pixel dimensions
2. Processes transparent areas (fills with black for dark terminals)
3. Encodes the image using Sixel graphics protocol
4. Outputs directly to terminal using ANSI escape sequences

## Technical Details

### Color Handling

- **24-bit True Color**: Full RGB color support using ANSI escape sequences
- **Transparency**: 
  - Block mode: Transparent pixels show terminal background
  - Sixel mode: Transparent areas filled with black (configurable)
- **Alpha Blending**: Semi-transparent pixels are properly blended

### Image Scaling

- Uses Lanczos3 resampling for high-quality downscaling
- Maintains aspect ratio automatically
- Adapts to terminal size dynamically
- Terminal character aspect ratio: 1:2 (width:height)

### Performance

- Fast image decoding using Go's standard library
- Efficient resizing with hardware acceleration support
- Minimal memory footprint
- No external dependencies except Go libraries

## Troubleshooting

### Image appears distorted

- The tool automatically maintains aspect ratio
- Check your terminal font settings (monospaced fonts work best)

### Sixel mode not working

- Verify your terminal supports Sixel graphics
- Try block mode as a fallback: `./crt image.png block`

### Colors look incorrect

- Ensure your terminal supports 24-bit true color
- Check terminal color scheme settings

### Transparent PNG shows black background in Sixel mode

- This is expected behavior (Sixel doesn't support transparency)
- Use block mode to show terminal background: `./crt image.png block`

## Dependencies

- [github.com/nfnt/resize](https://github.com/nfnt/resize) - High-quality image resizing
- [github.com/mattn/go-sixel](https://github.com/mattn/go-sixel) - Sixel graphics encoding
- [golang.org/x/image](https://golang.org/x/image) - Extended image format support

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
