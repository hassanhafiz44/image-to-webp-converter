use chrono::Local;
use clap::Parser;
use image::DynamicImage;
use rayon::prelude::*;
use std::fs;
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicU64, AtomicUsize, Ordering};
use std::thread;
use std::time::Instant;
use walkdir::WalkDir;

const SUPPORTED_EXTENSIONS: &[&str] = &["jpg", "jpeg", "png", "gif", "bmp", "tiff", "tif"];

#[derive(Parser)]
#[command(name = "image-converter")]
#[command(about = "High-performance image to WebP converter written in Rust")]
struct Args {
    /// WebP quality (1-100)
    #[arg(short, long, default_value_t = 80)]
    quality: u8,

    /// Input directory
    #[arg(short, long, default_value = "/app/images")]
    input: String,

    /// Output directory
    #[arg(short, long, default_value = "/app/output")]
    output: String,

    /// Number of parallel workers (default: CPU cores)
    #[arg(short, long, default_value_t = 0)]
    workers: usize,
}

struct ConversionResult {
    input: String,
    output: String,
    success: bool,
    message: String,
    original_size: u64,
    new_size: u64,
    savings: f64,
}

fn main() {
    let args = Args::parse();

    let quality = args.quality.clamp(1, 100);
    let workers = if args.workers > 0 {
        args.workers
    } else {
        thread::available_parallelism()
            .map(|n| n.get())
            .unwrap_or(4)
    };
    let input_dir = PathBuf::from(&args.input);
    let output_dir = PathBuf::from(&args.output);

    // Configure rayon thread pool
    rayon::ThreadPoolBuilder::new()
        .num_threads(workers)
        .build_global()
        .unwrap_or(());

    let start = Instant::now();
    let start_time = Local::now().format("%Y-%m-%d %H:%M:%S %Z").to_string();

    println!();
    println!("╔════════════════════════════════════════════╗");
    println!("║     Image to WebP Converter (Rust)         ║");
    println!("╚════════════════════════════════════════════╝");
    println!();
    println!("Started at: {}", start_time);
    println!();
    println!("Configuration:");
    println!("  • Input directory:  {}", input_dir.display());
    println!("  • Output directory: {}", output_dir.display());
    println!("  • Quality:          {}%", quality);
    println!("  • Workers:          {}", workers);
    println!();

    // Validate input directory
    if !input_dir.is_dir() {
        eprintln!("Error: Input directory does not exist: {}", input_dir.display());
        std::process::exit(1);
    }

    // Ensure output directory exists
    if let Err(e) = fs::create_dir_all(&output_dir) {
        eprintln!("Error: Failed to create output directory: {}", e);
        std::process::exit(1);
    }

    // Scan for image files
    let all_files = get_image_files(&input_dir);
    let total_found = all_files.len();

    if total_found == 0 {
        println!("⚠ No images found in {}", input_dir.display());
        println!("  Supported formats: {}", SUPPORTED_EXTENSIONS.join(", "));
        return;
    }

    println!("Found {} image(s)", total_found);

    // Filter already converted
    let (files, skipped) = filter_already_converted(&all_files, &input_dir, &output_dir);
    let total_files = files.len();

    if skipped > 0 {
        println!("  ⏭ Skipped {} file(s) (already converted)", skipped);
    }

    if total_files == 0 {
        println!("  All files already converted. Nothing to do.");
        return;
    }

    println!("  Converting {} file(s)", total_files);
    println!("{}", "-".repeat(50));

    // Parallel conversion
    let counter = AtomicUsize::new(0);
    let total_original = AtomicU64::new(0);
    let total_new = AtomicU64::new(0);
    let successful = AtomicUsize::new(0);
    let failed = AtomicUsize::new(0);

    files.par_iter().for_each(|file| {
        let result = convert_image(file, &input_dir, &output_dir, quality as f32);
        let current = counter.fetch_add(1, Ordering::Relaxed) + 1;
        let filename = Path::new(&result.input)
            .file_name()
            .unwrap_or_default()
            .to_string_lossy();

        if result.success {
            successful.fetch_add(1, Ordering::Relaxed);
            total_original.fetch_add(result.original_size, Ordering::Relaxed);
            total_new.fetch_add(result.new_size, Ordering::Relaxed);
            println!(
                "[{}/{}] ✓ {}: {} → {} ({:.2}% saved)",
                current,
                total_files,
                filename,
                format_bytes(result.original_size),
                format_bytes(result.new_size),
                result.savings
            );
        } else {
            failed.fetch_add(1, Ordering::Relaxed);
            println!(
                "[{}/{}] ✗ {}: {}",
                current, total_files, filename, result.message
            );
        }
    });

    let successful = successful.load(Ordering::Relaxed);
    let failed = failed.load(Ordering::Relaxed);
    let total_original = total_original.load(Ordering::Relaxed);
    let total_new = total_new.load(Ordering::Relaxed);

    let end_time = Local::now().format("%Y-%m-%d %H:%M:%S %Z").to_string();
    let elapsed = start.elapsed();

    println!();
    println!("{}", "=".repeat(50));
    println!("CONVERSION SUMMARY");
    println!("{}", "=".repeat(50));
    println!("  • Total files:      {}", total_files);
    println!("  • Successful:       {}", successful);
    println!("  • Failed:           {}", failed);

    if successful > 0 {
        let total_savings = if total_original > 0 {
            ((1.0 - total_new as f64 / total_original as f64) * 10000.0).round() / 100.0
        } else {
            0.0
        };
        println!("  • Original size:    {}", format_bytes(total_original));
        println!("  • New size:         {}", format_bytes(total_new));
        println!("  • Total savings:    {:.2}%", total_savings);
    }

    println!("  • Start time:       {}", start_time);
    println!("  • End time:         {}", end_time);
    println!("  • Time elapsed:     {}", format_duration(elapsed));
    println!();
    println!("Output directory: {}", output_dir.display());
    println!();
}

fn get_image_files(input_dir: &Path) -> Vec<PathBuf> {
    WalkDir::new(input_dir)
        .into_iter()
        .filter_map(|e| e.ok())
        .filter(|e| e.file_type().is_file())
        .filter(|e| {
            e.path()
                .extension()
                .and_then(|ext| ext.to_str())
                .map(|ext| SUPPORTED_EXTENSIONS.contains(&ext.to_lowercase().as_str()))
                .unwrap_or(false)
        })
        .map(|e| e.into_path())
        .collect()
}

fn get_output_path(input_path: &Path, input_dir: &Path, output_dir: &Path) -> PathBuf {
    let rel = input_path.strip_prefix(input_dir).unwrap_or(input_path);
    let mut out = output_dir.join(rel);
    out.set_extension("webp");
    out
}

fn filter_already_converted(
    files: &[PathBuf],
    input_dir: &Path,
    output_dir: &Path,
) -> (Vec<PathBuf>, usize) {
    let mut to_convert = Vec::with_capacity(files.len());
    let mut skipped = 0;

    for file in files {
        let output_path = get_output_path(file, input_dir, output_dir);
        if output_path.exists() {
            skipped += 1;
        } else {
            to_convert.push(file.clone());
        }
    }

    (to_convert, skipped)
}

fn convert_image(
    input_path: &Path,
    input_dir: &Path,
    output_dir: &Path,
    quality: f32,
) -> ConversionResult {
    let input_str = input_path.display().to_string();
    let output_path = get_output_path(input_path, input_dir, output_dir);
    let output_str = output_path.display().to_string();

    // Ensure output subdirectory exists
    if let Some(parent) = output_path.parent() {
        if let Err(e) = fs::create_dir_all(parent) {
            return ConversionResult {
                input: input_str,
                output: output_str,
                success: false,
                message: format!("Failed to create output dir: {}", e),
                original_size: 0,
                new_size: 0,
                savings: 0.0,
            };
        }
    }

    // Get original size
    let original_size = match fs::metadata(input_path) {
        Ok(m) => m.len(),
        Err(e) => {
            return ConversionResult {
                input: input_str,
                output: output_str,
                success: false,
                message: format!("Cannot stat input: {}", e),
                original_size: 0,
                new_size: 0,
                savings: 0.0,
            };
        }
    };

    // Load image
    let img: DynamicImage = match image::open(input_path) {
        Ok(img) => img,
        Err(e) => {
            return ConversionResult {
                input: input_str,
                output: output_str,
                success: false,
                message: format!("Failed to load image: {}", e),
                original_size,
                new_size: 0,
                savings: 0.0,
            };
        }
    };

    // Encode to WebP using the webp crate (native libwebp bindings)
    let encoder = match webp::Encoder::from_image(&img) {
        Ok(enc) => enc,
        Err(e) => {
            return ConversionResult {
                input: input_str,
                output: output_str,
                success: false,
                message: format!("Failed to create encoder: {}", e),
                original_size,
                new_size: 0,
                savings: 0.0,
            };
        }
    };

    let webp_data = encoder.encode(quality);

    // Write output
    if let Err(e) = fs::write(&output_path, &*webp_data) {
        return ConversionResult {
            input: input_str,
            output: output_str,
            success: false,
            message: format!("Failed to write output: {}", e),
            original_size,
            new_size: 0,
            savings: 0.0,
        };
    }

    let new_size = webp_data.len() as u64;
    let savings = if original_size > 0 {
        ((1.0 - new_size as f64 / original_size as f64) * 10000.0).round() / 100.0
    } else {
        0.0
    };

    ConversionResult {
        input: input_str,
        output: output_str,
        success: true,
        message: "Converted successfully".to_string(),
        original_size,
        new_size,
        savings,
    }
}

fn format_bytes(bytes: u64) -> String {
    let units = ["B", "KB", "MB", "GB"];
    let mut size = bytes as f64;
    let mut unit_index = 0;

    while size >= 1024.0 && unit_index < units.len() - 1 {
        size /= 1024.0;
        unit_index += 1;
    }

    format!("{:.2} {}", size, units[unit_index])
}

fn format_duration(d: std::time::Duration) -> String {
    let total_secs = d.as_secs_f64();

    if total_secs < 1.0 {
        return format!("{}ms", d.as_millis());
    }

    let minutes = (total_secs / 60.0) as u64;
    let secs = total_secs % 60.0;

    if minutes > 0 {
        format!("{}m {:.1}s", minutes, secs)
    } else {
        format!("{:.1}s", secs)
    }
}
