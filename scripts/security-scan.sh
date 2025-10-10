#!/bin/bash
set -euo pipefail

# Cryptoutil Security Scanner - Local Development Script
# This script runs the same security scans as CI/CD pipeline locally

# Default values
ALL=true
STATIC_ONLY=false
VULN_ONLY=false
CONTAINER_ONLY=false
OUTPUT_DIR="security-reports"
IMAGE_TAG="cryptoutil:latest"
SKIP_DOCKER=false
HELP=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --static-only)
            STATIC_ONLY=true
            ALL=false
            shift
            ;;
        --vuln-only)
            VULN_ONLY=true
            ALL=false
            shift
            ;;
        --container-only)
            CONTAINER_ONLY=true
            ALL=false
            shift
            ;;
        --output-dir)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --image-tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        --skip-docker)
            SKIP_DOCKER=true
            shift
            ;;
        --help|-h)
            HELP=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Show help
show_help() {
    cat << EOF
Cryptoutil Security Scanner - Local Development Script

USAGE:
    $0 [OPTIONS]

DESCRIPTION:
    Run comprehensive security scans locally for cryptoutil.
    Includes static analysis, vulnerability scanning, container security, and dependency analysis.

OPTIONS:
    --static-only        Run only static analysis tools (staticcheck, golangci-lint)
    --vuln-only         Run only vulnerability scans (govulncheck, trivy)
    --container-only    Run only container security scans (trivy image, docker scout)
    --output-dir DIR    Directory to save security reports (default: security-reports)
    --image-tag TAG     Docker image tag to scan for container security (default: cryptoutil:latest)
    --skip-docker       Skip Docker-based scans (trivy, docker scout)
    --help, -h          Show this help message

EXAMPLES:
    $0                                      # Run all security scans with default settings
    $0 --static-only --output-dir reports   # Run only static analysis, save to "reports"
    $0 --container-only --image-tag dev     # Run only container scans on specific image

SECURITY TOOLS:
    - Staticcheck: Go static analysis and lint checking
    - golangci-lint: Comprehensive Go linting with multiple analyzers
    - govulncheck: Official Go vulnerability database scanning
    - Trivy: File system and container vulnerability scanning
    - Docker Scout: Advanced container security analysis and recommendations

EOF
}

if [ "$HELP" = true ]; then
    show_help
    exit 0
fi

# Color output functions
print_header() {
    echo
    echo -e "\e[36m🛡️  $1\e[0m"
    echo -e "\e[36m$(printf '=%.0s' $(seq 1 $((${#1} + 4))))\e[0m"
}

print_success() {
    echo -e "\e[32m✅ $1\e[0m"
}

print_warning() {
    echo -e "\e[33m⚠️  $1\e[0m"
}

print_error() {
    echo -e "\e[31m❌ $1\e[0m"
}

print_info() {
    echo -e "\e[34mℹ️  $1\e[0m"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    local missing=()

    # Check Go
    if ! command -v go &> /dev/null; then
        missing+=("Go")
    fi

    # Check Docker (if not skipping Docker scans)
    if [ "$SKIP_DOCKER" = false ] && ! command -v docker &> /dev/null; then
        missing+=("Docker")
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        print_error "Missing required tools: ${missing[*]}"
        exit 1
    fi

    print_success "All prerequisites available"
}

# Create output directory
initialize_output_directory() {
    print_info "Creating output directory: $OUTPUT_DIR"

    rm -rf "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"

    print_success "Output directory ready: $OUTPUT_DIR"
}

# Install Go security tools
install_go_tools() {
    print_header "Installing Go Security Tools"

    local tools=(
        "staticcheck:honnef.co/go/tools/cmd/staticcheck@latest"
        "govulncheck:golang.org/x/vuln/cmd/govulncheck@latest"
        "golangci-lint:github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    )

    for tool_entry in "${tools[@]}"; do
        local tool_name="${tool_entry%%:*}"
        local tool_package="${tool_entry#*:}"

        print_info "Installing $tool_name..."
        if go install "$tool_package"; then
            print_success "$tool_name installed successfully"
        else
            print_warning "Failed to install $tool_name"
        fi
    done
}

# Run static analysis
run_static_analysis() {
    if [ "$ALL" = false ] && [ "$STATIC_ONLY" = false ]; then
        return
    fi

    print_header "Static Code Analysis"

    # Staticcheck
    print_info "Running Staticcheck..."
    if staticcheck -f sarif ./... > "$OUTPUT_DIR/staticcheck.sarif" 2>/dev/null && staticcheck ./...; then
        print_success "Staticcheck completed - no issues found"
    else
        print_warning "Staticcheck found potential issues - check output above"
    fi

    # golangci-lint
    print_info "Running golangci-lint..."
    if golangci-lint run --timeout=10m --config=.golangci.yml --out-format=sarif > "$OUTPUT_DIR/golangci-lint.sarif" && \
       golangci-lint run --timeout=10m --config=.golangci.yml; then
        print_success "golangci-lint completed - no issues found"
    else
        print_warning "golangci-lint found potential issues - check output above"
    fi
}

# Run vulnerability scans
run_vulnerability_scans() {
    if [ "$ALL" = false ] && [ "$VULN_ONLY" = false ]; then
        return
    fi

    print_header "Vulnerability Scanning"

    # govulncheck
    print_info "Running Go vulnerability check..."
    if govulncheck ./... > "$OUTPUT_DIR/govulncheck.txt" 2>&1 && govulncheck ./...; then
        print_success "No known Go vulnerabilities found"
    else
        print_warning "Go vulnerabilities detected - check output above"
    fi

    # Trivy file system scan (if not skipping Docker)
    if [ "$SKIP_DOCKER" = false ]; then
        print_info "Running Trivy file system scan..."
        if docker run --rm -v "$(pwd):/workspace" aquasec/trivy:latest fs --format sarif --output "/workspace/$OUTPUT_DIR/trivy-fs.sarif" /workspace && \
           docker run --rm -v "$(pwd):/workspace" aquasec/trivy:latest fs /workspace; then
            print_success "Trivy file system scan completed"
        else
            print_warning "Trivy found potential vulnerabilities - check output above"
        fi
    fi
}

# Run container security scans
run_container_scans() {
    if [ "$ALL" = false ] && [ "$CONTAINER_ONLY" = false ]; then
        return
    fi

    if [ "$SKIP_DOCKER" = true ]; then
        print_warning "Skipping container scans (Docker disabled)"
        return
    fi

    print_header "Container Security Scanning"

    # Check if image exists
    print_info "Checking for Docker image: $IMAGE_TAG"
    if ! docker images "$IMAGE_TAG" --format "table" &> /dev/null; then
        print_warning "Docker image '$IMAGE_TAG' not found. Building..."
        if ! docker build -t "$IMAGE_TAG" -f deployments/Dockerfile .; then
            print_error "Failed to build Docker image"
            return
        fi
        print_success "Docker image built successfully"
    fi

    # Trivy image scan
    print_info "Running Trivy container image scan..."
    if docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$(pwd)/$OUTPUT_DIR:/output" \
       aquasec/trivy:latest image --format sarif --output "/output/trivy-image.sarif" "$IMAGE_TAG" && \
       docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy:latest image "$IMAGE_TAG"; then
        print_success "Trivy image scan completed"
    else
        print_warning "Trivy found container vulnerabilities - check output above"
    fi

    # Docker Scout (if available)
    print_info "Running Docker Scout scans..."

    # Quick overview
    docker scout quickview "$IMAGE_TAG" > "$OUTPUT_DIR/docker-scout-quickview.txt" 2>&1 || true
    docker scout quickview "$IMAGE_TAG" || true

    # CVE analysis
    docker scout cves --format sarif --output "$OUTPUT_DIR/docker-scout-cves.sarif" "$IMAGE_TAG" 2>/dev/null || true
    docker scout cves "$IMAGE_TAG" || true

    # Recommendations
    docker scout recommendations "$IMAGE_TAG" > "$OUTPUT_DIR/docker-scout-recommendations.txt" 2>&1 || true
    docker scout recommendations "$IMAGE_TAG" || true

    print_success "Docker Scout scans completed"
}

# Generate summary report
generate_security_report() {
    print_header "Generating Security Summary Report"

    local report_path="$OUTPUT_DIR/security-summary.md"
    local timestamp=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

    cat > "$report_path" << EOF
# Security Scan Results

**Scan Date:** $timestamp
**Target:** cryptoutil project
**Output Directory:** $OUTPUT_DIR

## Scan Coverage

EOF

    if [ "$ALL" = true ] || [ "$STATIC_ONLY" = true ]; then
        cat >> "$report_path" << EOF
### Static Analysis
- ✅ **Staticcheck**: Go static analysis and lint checking
- ✅ **golangci-lint**: Comprehensive Go linting with multiple analyzers

EOF
    fi

    if [ "$ALL" = true ] || [ "$VULN_ONLY" = true ]; then
        cat >> "$report_path" << EOF
### Vulnerability Scanning
- ✅ **govulncheck**: Official Go vulnerability database scanning
- ✅ **Trivy FS**: File system and dependency vulnerability scanning

EOF
    fi

    if ([ "$ALL" = true ] || [ "$CONTAINER_ONLY" = true ]) && [ "$SKIP_DOCKER" = false ]; then
        cat >> "$report_path" << EOF
### Container Security
- ✅ **Trivy Image**: Container image vulnerability scanning
- ✅ **Docker Scout**: Advanced container security analysis and recommendations

EOF
    fi

    cat >> "$report_path" << EOF

## Report Files Generated

EOF

    # List generated report files
    for file in "$OUTPUT_DIR"/*; do
        if [ -f "$file" ]; then
            echo "- $(basename "$file")" >> "$report_path"
        fi
    done

    cat >> "$report_path" << EOF

## Next Steps

1. Review detailed reports in the $OUTPUT_DIR directory
2. Address HIGH and CRITICAL findings immediately
3. Update dependencies for known vulnerabilities
4. Consider security recommendations from Docker Scout
5. Run security scans regularly as part of development workflow

## Report Locations

All security reports are saved in: **$OUTPUT_DIR**

EOF

    print_success "Security summary report generated: $report_path"
}

# Main execution
main() {
    print_header "Cryptoutil Security Scanner"
    print_info "Starting comprehensive security analysis..."

    check_prerequisites
    initialize_output_directory
    install_go_tools

    run_static_analysis
    run_vulnerability_scans
    run_container_scans

    generate_security_report

    print_header "Security Scan Complete"
    print_success "All security scans completed successfully!"
    print_info "Reports saved to: $OUTPUT_DIR"
    print_info "Review the security-summary.md file for an overview of all findings"
}

# Execute main function
main "$@"
