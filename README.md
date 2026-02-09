# Image to WebP Converter

A high-performance Dockerized Rust tool that batch-converts images (JPG, PNG, GIF, BMP, TIFF) to WebP format. Uses native libwebp bindings and Rayon for data-parallel conversion across all CPU cores.

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
║     Image to WebP Converter (Rust)         ║
╚════════════════════════════════════════════╝

Started at: 2026-02-09 04:41:00 PKT

Configuration:
  • Input directory:  /app/images
  • Output directory: /app/output
  • Quality:          80%
  • Workers:          8

Found 3 image(s)
  Converting 3 file(s)
--------------------------------------------------
[1/3] ✓ photo.jpg: 2.50 MB → 450.00 KB (82.00% saved)
[2/3] ✓ logo.png: 150.00 KB → 45.00 KB (70.00% saved)
[3/3] ✓ banner.gif: 1.20 MB → 300.00 KB (75.00% saved)

==================================================
CONVERSION SUMMARY
==================================================
  • Total files:      3
  • Successful:       3
  • Failed:           0
  • Original size:    3.85 MB
  • New size:         795.00 KB
  • Total savings:    79.80%
  • Start time:       2026-02-09 04:41:00 PKT
  • End time:         2026-02-09 04:41:01 PKT
  • Time elapsed:     820ms

Output directory: /app/output
```

## Supported Formats

jpg, jpeg, png, gif, bmp, tiff, tif

## Troubleshooting

### Rebuild Container (after code or Dockerfile changes)

```bash
docker compose build --no-cache
```

## Project Structure

```
├── docker/
│   └── Dockerfile       # Multi-stage Rust build + Alpine runtime
├── src/
│   └── main.rs          # Rust source
├── Cargo.toml           # Rust dependencies
├── images/              # Input images (gitignored)
├── output/              # Converted WebP files (gitignored)
├── convert.sh           # Helper script (Linux/Mac)
├── convert.bat          # Helper script (Windows)
├── docker-compose.yml
└── README.md
```
