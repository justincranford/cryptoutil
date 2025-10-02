#!/bin/bash

# DAST (Dynamic Application Security Testing) Script for cryptoutil
# Runs OWASP ZAP and Nuclei security scans locally

set -e

# Default values
CONFIG="configs/local/config.yml"
PORT=8080
TARGET_URL=""
SKIP_ZAP=false
SKIP_NUCLEI=false
OUTPUT_DIR="dast-reports"
SHOW_HELP=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

# Status functions
write_status() {
    echo -e "${BLUE}🔒 DAST: $1${NC}"
}

write_error() {
    echo -e "${RED}❌ ERROR: $1${NC}"
}

write_success() {
    echo -e "${GREEN}✅ SUCCESS: $1${NC}"
}

write_warning() {
    echo -e "${YELLOW}⚠️  WARNING: $1${NC}"
}

# Help function
show_help() {
    cat << EOF
DAST Security Testing Script

USAGE:
    $0 [OPTIONS]

DESCRIPTION:
    Starts cryptoutil application and runs OWASP ZAP and Nuclei security scans.
    Generates comprehensive security reports for local development testing.

OPTIONS:
    -c, --config CONFIG      Configuration file to use (default: configs/local/config.yml)
    -p, --port PORT          Port to run the application on (default: 8080)
    -t, --target-url URL     Target URL for scanning (default: http://localhost:PORT)
    --skip-zap               Skip OWASP ZAP scanning
    --skip-nuclei            Skip Nuclei scanning
    -o, --output-dir DIR     Output directory for reports (default: dast-reports)
    -h, --help               Show this help message

EXAMPLES:
    $0                                           # Run DAST with default settings
    $0 -c configs/test/config.yml -p 9090       # Custom configuration and port
    $0 --skip-zap                                # Run only Nuclei scan, skip ZAP
    $0 -o reports                                # Save reports to 'reports' directory

PREREQUISITES:
    - Docker (for OWASP ZAP)
    - Nuclei (will be installed if missing)
    - Go (for building cryptoutil)

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--config)
            CONFIG="$2"
            shift 2
            ;;
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -t|--target-url)
            TARGET_URL="$2"
            shift 2
            ;;
        --skip-zap)
            SKIP_ZAP=true
            shift
            ;;
        --skip-nuclei)
            SKIP_NUCLEI=true
            shift
            ;;
        -o|--output-dir)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            write_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Set target URL if not provided
if [[ -z "$TARGET_URL" ]]; then
    TARGET_URL="http://localhost:$PORT"
fi

# Cleanup function
cleanup() {
    write_status "Cleaning up..."
    if [[ -n "$APP_PID" ]]; then
        kill $APP_PID 2>/dev/null || true
        sleep 2
        kill -9 $APP_PID 2>/dev/null || true
        write_status "Application stopped"
    fi
}

# Set trap for cleanup
trap cleanup EXIT INT TERM

# Check prerequisites
write_status "Checking prerequisites..."

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    write_error "Docker is not available. Please install Docker."
    exit 1
fi
write_status "Docker is available"

# Pull ZAP image if not skipping ZAP
if [[ "$SKIP_ZAP" != true ]]; then
    write_status "Pulling OWASP ZAP Docker image..."
    if ! docker pull zaproxy/zap-stable:latest; then
        write_error "Failed to pull ZAP Docker image"
        exit 1
    fi
fi

# Check if Nuclei is available
if [[ "$SKIP_NUCLEI" != true ]]; then
    if ! command -v nuclei &> /dev/null; then
        write_warning "Nuclei not found. Installing..."
        if ! go install -v github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest; then
            write_error "Failed to install Nuclei"
            exit 1
        fi
        # Update nuclei templates
        nuclei -update-templates
    fi
    write_status "Nuclei is available"
fi

# Create output directory
write_status "Creating output directory: $OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Build application
write_status "Building cryptoutil application..."
if ! go build -o cryptoutil ./cmd/cryptoutil; then
    write_error "Failed to build application"
    exit 1
fi

# Start application
write_status "Starting cryptoutil on port $PORT..."
./cryptoutil server start --dev --config "$CONFIG" &
APP_PID=$!

# Wait for application to be ready
write_status "Waiting for application to be ready..."
MAX_ATTEMPTS=30
ATTEMPT=0
READY=false

# Health check on HTTP port 9090 (both endpoints)
HEALTH_URLS=(
    "http://localhost:9090/readyz"
    "http://localhost:9090/livez"
)

while [[ $ATTEMPT -lt $MAX_ATTEMPTS && "$READY" != true ]]; do
    sleep 2
    ((ATTEMPT++))
    for HEALTH_URL in "${HEALTH_URLS[@]}"; do
        if curl -f -k "$HEALTH_URL" >/dev/null 2>&1; then
            READY=true
            write_success "Application is ready on $HEALTH_URL"
            break
        fi
    done
    if [[ "$READY" != true ]]; then
        write_status "Attempt $ATTEMPT/$MAX_ATTEMPTS - waiting for application..."
    fi
done

if [[ "$READY" != true ]]; then
    write_error "Application failed to start within timeout"
    exit 1
fi

# Verify OpenAPI spec is available on HTTPS port 8080
OPENAPI_URLS=(
    "https://localhost:$PORT/ui/swagger/doc.json"
    "$TARGET_URL/ui/swagger/doc.json"
)

OPENAPI_DOWNLOADED=false
for OPENAPI_URL in "${OPENAPI_URLS[@]}"; do
    if curl -f -k "$OPENAPI_URL" -o "$OUTPUT_DIR/openapi.json" >/dev/null 2>&1; then
        write_status "OpenAPI specification downloaded from $OPENAPI_URL"
        OPENAPI_DOWNLOADED=true
        break
    fi
done

if [[ "$OPENAPI_DOWNLOADED" != true ]]; then
    write_warning "OpenAPI specification not available at any expected endpoint"
fi

# Run OWASP ZAP scans
if [[ "$SKIP_ZAP" != true ]]; then
    write_status "Running OWASP ZAP Full Scan..."
    if docker run --rm -t \
        -v "$(pwd)/$OUTPUT_DIR:/zap/wrk/:rw" \
        zaproxy/zap-stable:latest \
        zap-full-scan.py \
        -t "$TARGET_URL" \
        -r zap-full-report.html \
        -J zap-full-report.json \
        -m 10 \
        -T 60 \
        -z "-config rules.cookie.ignorelist=JSESSIONID,csrftoken"; then
        write_success "ZAP Full Scan completed"
    else
        write_warning "ZAP Full Scan completed with findings (exit code: $?)"
    fi

    # Run ZAP API scan if OpenAPI spec is available
    if [[ -f "$OUTPUT_DIR/openapi.json" ]]; then
        write_status "Running OWASP ZAP API Scan..."
        if docker run --rm -t \
            -v "$(pwd)/$OUTPUT_DIR:/zap/wrk/:rw" \
            zaproxy/zap-stable:latest \
            zap-api-scan.py \
            -t "$TARGET_URL/swagger/openapi.json" \
            -f openapi \
            -r zap-api-report.html \
            -J zap-api-report.json \
            -T 60; then
            write_success "ZAP API Scan completed"
        else
            write_warning "ZAP API Scan completed with findings (exit code: $?)"
        fi
    fi
fi

# Run Nuclei scan
if [[ "$SKIP_NUCLEI" != true ]]; then
    write_status "Running Nuclei Vulnerability Scan..."
    if nuclei \
        -target "$TARGET_URL" \
        -templates "cves/,vulnerabilities/,security-misconfiguration/,default-logins/,exposed-panels/,takeovers/,technologies/" \
        -json-export "$OUTPUT_DIR/nuclei-report.json" \
        -stats \
        -silent; then
        write_success "Nuclei scan completed"
    else
        write_warning "Nuclei scan completed with findings (exit code: $?)"
    fi
fi

# Generate summary report
write_status "Generating summary report..."
SUMMARY_FILE="$OUTPUT_DIR/dast-summary.md"
TIMESTAMP=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

cat > "$SUMMARY_FILE" << EOF
# DAST Security Scan Results

**Scan Date:** $TIMESTAMP
**Target URL:** $TARGET_URL
**Configuration:** $CONFIG

## Scan Coverage

EOF

if [[ "$SKIP_ZAP" != true ]]; then
    cat >> "$SUMMARY_FILE" << EOF
- ✅ **OWASP ZAP Full Scan:** Comprehensive web application security testing
- ✅ **OWASP ZAP API Scan:** OpenAPI specification-driven API security testing

EOF
fi

if [[ "$SKIP_NUCLEI" != true ]]; then
    cat >> "$SUMMARY_FILE" << EOF
- ✅ **Nuclei Scan:** CVE and vulnerability template-based testing

EOF
fi

cat >> "$SUMMARY_FILE" << EOF
## Reports Generated

EOF

# List generated files
for file in "$OUTPUT_DIR"/*; do
    if [[ -f "$file" ]]; then
        echo "- $(basename "$file")" >> "$SUMMARY_FILE"
    fi
done

cat >> "$SUMMARY_FILE" << EOF

## Next Steps

1. Review scan reports for HIGH or CRITICAL findings
2. Open HTML reports in your browser for detailed analysis
3. Address any security vulnerabilities found
4. Consider adding custom ZAP rules for cryptographic endpoints

## Report Locations

EOF

if [[ -f "$OUTPUT_DIR/zap-full-report.html" ]]; then
    echo "- **ZAP Full Report:** $OUTPUT_DIR/zap-full-report.html" >> "$SUMMARY_FILE"
fi
if [[ -f "$OUTPUT_DIR/zap-api-report.html" ]]; then
    echo "- **ZAP API Report:** $OUTPUT_DIR/zap-api-report.html" >> "$SUMMARY_FILE"
fi
if [[ -f "$OUTPUT_DIR/nuclei-report.json" ]]; then
    echo "- **Nuclei Report:** $OUTPUT_DIR/nuclei-report.json" >> "$SUMMARY_FILE"
fi

# Show results
write_success "DAST scanning completed!"
write_status "Summary report: $SUMMARY_FILE"
write_status "All reports saved to: $OUTPUT_DIR"

if [[ -f "$OUTPUT_DIR/zap-full-report.html" ]]; then
    write_success "Open ZAP Full Report: $OUTPUT_DIR/zap-full-report.html"
fi
if [[ -f "$OUTPUT_DIR/zap-api-report.html" ]]; then
    write_success "Open ZAP API Report: $OUTPUT_DIR/zap-api-report.html"
fi

exit 0
