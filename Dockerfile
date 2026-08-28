# ============================================================================
#  Portable MariaDB Backup & Restore — Dockerfile
#  Multi-stage build: produces a slim Linux image with the Wails binary.
#
#  Build:  docker build -t mariadb-restore .
#  Run:    docker run --rm -it mariadb-restore
#
#  NOTE: Wails desktop apps require a display server (X11/Wayland).
#  To run with host display:
#    docker run --rm -it \
#      -e DISPLAY=$DISPLAY \
#      -v /tmp/.X11-unix:/tmp/.X11-unix \
#      mariadb-restore
# ============================================================================

# ---------------------------------------------------------------------------
#  Stage 1 — Build frontend (React + Vite)
# ---------------------------------------------------------------------------
FROM node:20-alpine AS frontend

WORKDIR /app/frontend

# Install frontend deps (cached layer)
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci --prefer-offline 2>/dev/null || npm install

# Copy source and build
COPY frontend/ ./
RUN npm run build

# ---------------------------------------------------------------------------
#  Stage 2 — Build Go binary with Wails
# ---------------------------------------------------------------------------
FROM golang:1.25-bookworm AS builder

# System dependencies required by Wails (Linux build)
RUN apt-get update && apt-get install -y --no-install-recommends \
        libgtk-3-dev \
        libwebkit2gtk-4.1-dev \
        build-essential \
        pkg-config \
        gcc \
    && rm -rf /var/lib/apt/lists/*

# Install Wails CLI
RUN go install github.com/wailsapp/wails/v2/cmd/wails@latest

WORKDIR /app

# Cache Go modules (cached layer)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy full source
COPY . .

# Copy built frontend from Stage 1
COPY --from=frontend /app/frontend/dist ./frontend/dist

# Build the binary
#   -trimpath   : reproducible builds
#   -ldflags    : strip debug info for smaller binary
#   -tags       : Wails v2 (≤ 2.15) hardcodes webkit2gtk-4.0 in its cgo
#                 directives. Auto-pick the webkit2_41 tag when only 4.1
#                 dev headers are present (Ubuntu 24.04+, Fedora 40+,
#                 Linux Mint 22+). Older Debian/Ubuntu keep the default 4.0.
RUN WAILS_TAGS="$( \
        if pkg-config --exists webkit2gtk-4.1 2>/dev/null && \
           ! pkg-config --exists webkit2gtk-4.0 2>/dev/null; then \
            echo webkit2_41; \
        fi)"; \
    wails build \
        -platform linux/amd64 \
        -o /app/mariadb-restore-desktop-app \
        -trimpath \
        -ldflags "-s -w" \
        ${WAILS_TAGS:+-tags $WAILS_TAGS}

# ---------------------------------------------------------------------------
#  Stage 3 — Minimal runtime image
# ---------------------------------------------------------------------------
FROM debian:bookworm-slim AS runtime

# Runtime libraries required by the Wails binary (GTK + WebKit2)
RUN apt-get update && apt-get install -y --no-install-recommends \
        libgtk-3-0 \
        libwebkit2gtk-4.1-37 \
        libgl1 \
        libglib2.0-0 \
        libdbus-1-3 \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user
RUN groupadd -r appuser && useradd -r -g appuser -m appuser

# Copy binary from builder
COPY --from=builder /app/mariadb-restore-desktop-app /usr/local/bin/mariadb-restore-desktop-app

# The app stores data next to the binary (executable scope).
# Mount a volume at /data to persist profiles, keys, and catalog.
RUN mkdir -p /data && chown appuser:appuser /data
VOLUME /data

USER appuser
WORKDIR /data

ENTRYPOINT ["mariadb-restore-desktop-app"]
