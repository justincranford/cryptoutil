#!/bin/bash
# Run mutation testing with Gremlins for the cryptoutil project
# Usage: ./scripts/mutation-test.sh [target] [options]

set -e

# Default values
TARGET=""
DRY_RUN=false
WORKERS=2
TIMEOUT_COEFF=3

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --target)
            TARGET="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --workers)
            WORKERS="$2"
            shift 2
            ;;
        --timeout-coefficient)
            TIMEOUT_COEFF="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --target TARGET             Target package to test (default: high-coverage packages)"
            echo "  --dry-run                   Run in dry-run mode without executing tests"
            echo "  --workers N                 Number of parallel workers (default: 2)"
            echo "  --timeout-coefficient N     Timeout coefficient multiplier (default: 3)"
            echo "  -h, --help                  Show this help message"
            exit 0
            ;;
        *)
            if [[ -z "$TARGET" ]]; then
                TARGET="$1"
            fi
            shift
            ;;
    esac
done

echo "🧪 Starting Mutation Testing with Gremlins"
echo "=========================================="

# Check if gremlins is installed
if ! command -v gremlins &> /dev/null; then
    echo "❌ Gremlins not found. Installing..."
    go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
    if [[ $? -ne 0 ]]; then
        echo "❌ Failed to install Gremlins"
        exit 1
    fi
    echo "✅ Gremlins installed successfully"
fi

# High-coverage packages to test
HIGH_COVERAGE_PACKAGES=(
    "./internal/common/util/datetime/"
    "./internal/common/util/thread/"
    "./internal/common/util/sysinfo/"
    "./internal/common/util/combinations/"
    "./internal/common/crypto/certificate/"
    "./internal/common/crypto/digests/"
)

# If no target specified, use high-coverage packages
if [[ -z "$TARGET" ]]; then
    TARGETS_TO_TEST=("${HIGH_COVERAGE_PACKAGES[@]}")
else
    TARGETS_TO_TEST=("$TARGET")
fi

TOTAL_KILLED=0
TOTAL_LIVED=0
TOTAL_NOT_COVERED=0
TOTAL_TIMED_OUT=0
TOTAL_NOT_VIABLE=0

echo "📊 Target packages:"
for package in "${TARGETS_TO_TEST[@]}"; do
    echo "  - $package"
done
echo ""

for package in "${TARGETS_TO_TEST[@]}"; do
    echo "🎯 Testing package: $package"
    echo "----------------------------------------"

    # Build the gremlins command
    CMD_ARGS=(
        "unleash"
        "$package"
        "--workers" "$WORKERS"
        "--timeout-coefficient" "$TIMEOUT_COEFF"
        "--output" "gremlins-$(basename "$package").json"
    )

    if [[ "$DRY_RUN" == "true" ]]; then
        CMD_ARGS+=("--dry-run")
    fi

    # Run gremlins and capture output
    if OUTPUT=$(gremlins "${CMD_ARGS[@]}" 2>&1); then
        echo "✅ Mutation testing completed for $package"
    else
        echo "⚠️  Mutation testing completed with warnings for $package"
    fi

    # Parse results from output
    if [[ $OUTPUT =~ Killed:\ ([0-9]+),\ Lived:\ ([0-9]+),\ Not\ covered:\ ([0-9]+) ]]; then
        KILLED=${BASH_REMATCH[1]}
        LIVED=${BASH_REMATCH[2]}
        NOT_COVERED=${BASH_REMATCH[3]}

        TOTAL_KILLED=$((TOTAL_KILLED + KILLED))
        TOTAL_LIVED=$((TOTAL_LIVED + LIVED))
        TOTAL_NOT_COVERED=$((TOTAL_NOT_COVERED + NOT_COVERED))
    fi

    if [[ $OUTPUT =~ Timed\ out:\ ([0-9]+),\ Not\ viable:\ ([0-9]+) ]]; then
        TIMED_OUT=${BASH_REMATCH[1]}
        NOT_VIABLE=${BASH_REMATCH[2]}

        TOTAL_TIMED_OUT=$((TOTAL_TIMED_OUT + TIMED_OUT))
        TOTAL_NOT_VIABLE=$((TOTAL_NOT_VIABLE + NOT_VIABLE))
    fi

    echo "$OUTPUT"
    echo ""
done

# Summary
echo "📈 MUTATION TESTING SUMMARY"
echo "==========================="
echo "Total Killed: $TOTAL_KILLED"
echo "Total Lived: $TOTAL_LIVED"
echo "Total Not Covered: $TOTAL_NOT_COVERED"
echo "Total Timed Out: $TOTAL_TIMED_OUT"
echo "Total Not Viable: $TOTAL_NOT_VIABLE"

TOTAL_TESTED=$((TOTAL_KILLED + TOTAL_LIVED))
if [[ $TOTAL_TESTED -gt 0 ]]; then
    EFFICACY=$(( (TOTAL_KILLED * 100) / TOTAL_TESTED ))
    if [[ $EFFICACY -ge 75 ]]; then
        echo "Test Efficacy: ${EFFICACY}% ✅"
    else
        echo "Test Efficacy: ${EFFICACY}% ❌"
    fi
else
    echo "Test Efficacy: N/A (no mutations tested)"
fi

# Set exit code based on results
if [[ $TOTAL_LIVED -gt 0 && "$DRY_RUN" != "true" ]]; then
    echo "❌ Mutation testing found $TOTAL_LIVED survived mutations"
    exit 1
else
    echo "✅ Mutation testing completed successfully"
    exit 0
fi
