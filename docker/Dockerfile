# docker/Dockerfile
FROM php:8.4-cli

LABEL maintainer="Avas" \
      description="PHP Image to WebP Converter"

# Install system dependencies, PHP extensions, and clean up in a single layer
RUN apt-get update && apt-get install -y --no-install-recommends \
    libmagickwand-dev \
    libmagickcore-dev \
    imagemagick \
    libwebp-dev \
    libpng-dev \
    libjpeg-dev \
    libgif-dev \
    && docker-php-ext-install -j$(nproc) pcntl opcache \
    && pecl install imagick && docker-php-ext-enable imagick \
    && docker-php-ext-configure gd --with-jpeg --with-webp \
    && docker-php-ext-install -j$(nproc) gd \
    && apt-get purge -y --auto-remove -o APT::AutoRemove::RecommendsImportant=false \
    && rm -rf /var/lib/apt/lists/* /tmp/pear

# Verify ImageMagick supports WebP
RUN convert -list format | grep -i webp

# Set working directory
WORKDIR /app

# Set recommended PHP settings
RUN echo "memory_limit=512M" > /usr/local/etc/php/conf.d/memory.ini

# Copy source files
COPY src/ /app/src/

# Default command runs the converter
CMD ["php", "src/convert.php"]
