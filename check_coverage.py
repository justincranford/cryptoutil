import subprocess, sys

# Run testdb tests with coverage
result = subprocess.run(
    ["go", "test", "-count=1", "-coverprofile=coverage_testdb.out",
     "./internal/apps/template/service/testing/testdb/"],
    capture_output=True, text=True
)
print("STDOUT:", result.stdout)
print("STDERR:", result.stderr[:500] if result.stderr else "")

# Get function coverage
result2 = subprocess.run(
    ["go", "tool", "cover", "-func=coverage_testdb.out"],
    capture_output=True, text=True
)
print("COVERAGE:")
for line in result2.stdout.splitlines():
    if "0.0%" in line or "total" in line.lower():
        print(" ", line)