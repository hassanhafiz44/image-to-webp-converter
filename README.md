# Image to WebP Converter

A Dockerized PHP tool that batch-converts images (JPG, PNG, GIF, BMP, TIFF) to WebP format using ImageMagick.

## Quick Start

1. Place your images in the `images/` directory.
2. Run the converter:

```bash
docker compose up --build
```

3. Converted WebP files will appear in the `output/` directory.

## Helper Scripts

Helper scripts are provided so you don't need to remember the full Docker command.

### Linux / Mac

```bash
./convert.sh
```

### Windows

```batch
convert.bat
```

### Usage Examples

```bash
# Basic usage - converts all images in /images folder
./convert.sh

# With quality setting
./convert.sh -q 85

# Custom input/output directories
./convert.sh -q 90 -i /app/images -o /app/output

# Show help
./convert.sh -h

# Windows
convert.bat -q 85
```

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `-q, --quality` | WebP quality (1-100) | 80 |
| `-i, --input` | Input directory | /app/images |
| `-o, --output` | Output directory | /app/output |
| `-h, --help` | Show help | — |

## Expected Output

```
╔════════════════════════════════════════════╗
║     Image to WebP Converter                ║
╚════════════════════════════════════════════╝

✓ ImageMagick version: ImageMagick 7.1.1-26
✓ WebP support: Available

Configuration:
  • Input directory:  /app/images
  • Output directory: /app/output
  • Quality:          80%

Found 3 image(s) to convert
--------------------------------------------------
[1/3] Converting: photo.jpg
    ✓ Saved: photo.webp
    ✓ Size: 2.5 MB → 450 KB (82% saved)
[2/3] Converting: logo.png
    ✓ Saved: logo.webp
    ✓ Size: 150 KB → 45 KB (70% saved)
[3/3] Converting: banner.gif
    ✓ Saved: banner.webp
    ✓ Size: 1.2 MB → 300 KB (75% saved)

==================================================
CONVERSION SUMMARY
==================================================
  • Total files:      3
  • Successful:       3
  • Failed:           0
  • Original size:    3.85 MB
  • New size:         795 KB
  • Total savings:    79.8%

Output directory: /app/output
```

## Supported Formats

jpg, jpeg, png, gif, bmp, tiff, tif

## Troubleshooting

### Check WebP Support in ImageMagick

```bash
docker compose run --rm php convert -list format | grep -i webp
```

### Check Imagick Extension

```bash
docker compose run --rm php php -i | grep -i imagick
```

### Rebuild Container (after Dockerfile changes)

```bash
docker compose build --no-cache
```

## Project Structure

```
├── docker/
│   └── Dockerfile
├── src/
│   └── convert.php
├── images/          # Input images (gitignored)
├── output/          # Converted WebP files (gitignored)
├── convert.sh       # Helper script (Linux/Mac)
├── convert.bat      # Helper script (Windows)
├── docker-compose.yml
└── README.md
```
