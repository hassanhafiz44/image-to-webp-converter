<?php
// src/convert.php

declare(strict_types=1);

/**
 * Image to WebP Converter
 * Converts images from the /images directory to WebP format
 */

class ImageConverter
{
    private string $inputDir;
    private string $outputDir;
    private array $supportedFormats = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'tiff', 'tif'];
    private int $quality;
    private bool $preserveStructure;
    
    public function __construct(
        string $inputDir = '/app/images',
        string $outputDir = '/app/output',
        int $quality = 80,
        bool $preserveStructure = true
    ) {
        $this->inputDir = rtrim($inputDir, '/');
        $this->outputDir = rtrim($outputDir, '/');
        $this->quality = $quality;
        $this->preserveStructure = $preserveStructure;
        
        $this->validateEnvironment();
        $this->ensureDirectoriesExist();
    }
    
    /**
     * Validate that ImageMagick is available and supports WebP
     */
    private function validateEnvironment(): void
    {
        if (!extension_loaded('imagick')) {
            throw new RuntimeException('ImageMagick extension is not loaded!');
        }
        
        $formats = \Imagick::queryFormats('WEBP');
        if (empty($formats)) {
            throw new RuntimeException('ImageMagick does not support WebP format!');
        }
        
        $this->log("✓ ImageMagick version: " . \Imagick::getVersion()['versionString']);
        $this->log("✓ WebP support: Available");
    }
    
    /**
     * Ensure input and output directories exist
     */
    private function ensureDirectoriesExist(): void
    {
        if (!is_dir($this->inputDir)) {
            throw new RuntimeException("Input directory does not exist: {$this->inputDir}");
        }
        
        if (!is_dir($this->outputDir)) {
            if (!mkdir($this->outputDir, 0755, true)) {
                throw new RuntimeException("Failed to create output directory: {$this->outputDir}");
            }
        }
        
        if (!is_writable($this->outputDir)) {
            throw new RuntimeException("Output directory is not writable: {$this->outputDir}");
        }
    }
    
    /**
     * Get all image files from input directory
     */
    private function getImageFiles(): array
    {
        $files = [];
        $iterator = new RecursiveIteratorIterator(
            new RecursiveDirectoryIterator($this->inputDir, RecursiveDirectoryIterator::SKIP_DOTS)
        );
        
        foreach ($iterator as $file) {
            if ($file->isFile()) {
                $extension = strtolower($file->getExtension());
                if (in_array($extension, $this->supportedFormats)) {
                    $files[] = $file->getPathname();
                }
            }
        }
        
        return $files;
    }
    
    /**
     * Convert a single image to WebP
     */
    private function convertImage(string $inputPath): array
    {
        $result = [
            'input' => $inputPath,
            'output' => '',
            'success' => false,
            'message' => '',
            'original_size' => 0,
            'new_size' => 0,
            'savings' => 0
        ];
        
        try {
            // Calculate output path
            $relativePath = str_replace($this->inputDir, '', $inputPath);
            $pathInfo = pathinfo($relativePath);
            
            if ($this->preserveStructure && !empty($pathInfo['dirname']) && $pathInfo['dirname'] !== '.') {
                $outputSubDir = $this->outputDir . $pathInfo['dirname'];
                if (!is_dir($outputSubDir)) {
                    mkdir($outputSubDir, 0755, true);
                }
                $outputPath = $outputSubDir . '/' . $pathInfo['filename'] . '.webp';
            } else {
                $outputPath = $this->outputDir . '/' . $pathInfo['filename'] . '.webp';
            }
            
            $result['output'] = $outputPath;
            $result['original_size'] = filesize($inputPath);
            
            // Create Imagick instance and convert
            $imagick = new \Imagick($inputPath);
            
            // Handle animated GIFs
            if ($imagick->getNumberImages() > 1) {
                $imagick = $imagick->coalesceImages();
                foreach ($imagick as $frame) {
                    $frame->setImageFormat('webp');
                    $frame->setImageCompressionQuality($this->quality);
                }
                $imagick->writeImages($outputPath, true);
            } else {
                // Standard image conversion
                $imagick->setImageFormat('webp');
                $imagick->setImageCompressionQuality($this->quality);
                
                // Strip metadata to reduce file size
                $imagick->stripImage();
                
                // Set WebP specific options
                $imagick->setOption('webp:lossless', 'false');
                $imagick->setOption('webp:method', '6'); // Compression method (0-6, 6 is slowest but best)
                
                $imagick->writeImage($outputPath);
            }
            
            $imagick->clear();
            $imagick->destroy();
            
            $result['new_size'] = filesize($outputPath);
            $result['savings'] = $result['original_size'] > 0 
                ? round((1 - $result['new_size'] / $result['original_size']) * 100, 2)
                : 0;
            $result['success'] = true;
            $result['message'] = 'Converted successfully';
            
        } catch (\ImagickException $e) {
            $result['message'] = 'ImageMagick error: ' . $e->getMessage();
        } catch (\Exception $e) {
            $result['message'] = 'Error: ' . $e->getMessage();
        }
        
        return $result;
    }
    
    /**
     * Format bytes to human readable format
     */
    private function formatBytes(int $bytes): string
    {
        $units = ['B', 'KB', 'MB', 'GB'];
        $unitIndex = 0;
        $size = (float) $bytes;
        
        while ($size >= 1024 && $unitIndex < count($units) - 1) {
            $size /= 1024;
            $unitIndex++;
        }
        
        return round($size, 2) . ' ' . $units[$unitIndex];
    }
    
    /**
     * Log a message to console
     */
    private function log(string $message): void
    {
        echo $message . PHP_EOL;
    }
    
    /**
     * Run the conversion process
     */
    public function run(): void
    {
        $this->log("");
        $this->log("╔════════════════════════════════════════════╗");
        $this->log("║     Image to WebP Converter                ║");
        $this->log("╚════════════════════════════════════════════╝");
        $this->log("");
        $this->log("Configuration:");
        $this->log("  • Input directory:  {$this->inputDir}");
        $this->log("  • Output directory: {$this->outputDir}");
        $this->log("  • Quality:          {$this->quality}%");
        $this->log("");
        
        $files = $this->getImageFiles();
        $totalFiles = count($files);
        
        if ($totalFiles === 0) {
            $this->log("⚠ No images found in {$this->inputDir}");
            $this->log("  Supported formats: " . implode(', ', $this->supportedFormats));
            return;
        }
        
        $this->log("Found {$totalFiles} image(s) to convert");
        $this->log(str_repeat("-", 50));
        
        $successful = 0;
        $failed = 0;
        $totalOriginalSize = 0;
        $totalNewSize = 0;
        
        foreach ($files as $index => $file) {
            $current = $index + 1;
            $filename = basename($file);
            
            $this->log("[{$current}/{$totalFiles}] Converting: {$filename}");
            
            $result = $this->convertImage($file);
            
            if ($result['success']) {
                $successful++;
                $totalOriginalSize += $result['original_size'];
                $totalNewSize += $result['new_size'];
                
                $this->log("    ✓ Saved: " . basename($result['output']));
                $this->log("    ✓ Size: {$this->formatBytes($result['original_size'])} → {$this->formatBytes($result['new_size'])} ({$result['savings']}% saved)");
            } else {
                $failed++;
                $this->log("    ✗ Failed: {$result['message']}");
            }
        }
        
        // Summary
        $this->log("");
        $this->log(str_repeat("=", 50));
        $this->log("CONVERSION SUMMARY");
        $this->log(str_repeat("=", 50));
        $this->log("  • Total files:      {$totalFiles}");
        $this->log("  • Successful:       {$successful}");
        $this->log("  • Failed:           {$failed}");
        
        if ($successful > 0) {
            $totalSavings = $totalOriginalSize > 0 
                ? round((1 - $totalNewSize / $totalOriginalSize) * 100, 2)
                : 0;
            $this->log("  • Original size:    {$this->formatBytes($totalOriginalSize)}");
            $this->log("  • New size:         {$this->formatBytes($totalNewSize)}");
            $this->log("  • Total savings:    {$totalSavings}%");
        }
        
        $this->log("");
        $this->log("Output directory: {$this->outputDir}");
        $this->log("");
    }
}

// Parse command line options
$options = getopt('q:i:o:h', ['quality:', 'input:', 'output:', 'help']);

if (isset($options['h']) || isset($options['help'])) {
    echo <<<HELP
    
Image to WebP Converter
Usage: php convert.php [options]

Options:
  -q, --quality <1-100>   Set WebP quality (default: 80)
  -i, --input <path>      Input directory (default: /app/images)
  -o, --output <path>     Output directory (default: /app/output)
  -h, --help              Show this help message

Examples:
  php convert.php
  php convert.php -q 90
  php convert.php --quality 75 --input /custom/input --output /custom/output

HELP;
    exit(0);
}

$quality = (int)($options['q'] ?? $options['quality'] ?? 80);
$inputDir = $options['i'] ?? $options['input'] ?? '/app/images';
$outputDir = $options['o'] ?? $options['output'] ?? '/app/output';

// Validate quality
$quality = max(1, min(100, $quality));

try {
    $converter = new ImageConverter($inputDir, $outputDir, $quality);
    $converter->run();
} catch (Exception $e) {
    echo "Error: " . $e->getMessage() . PHP_EOL;
    exit(1);
}
