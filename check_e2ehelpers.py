import subprocess
pkg = "./internal/apps/template/service/testing/e2e_helpers/"
result = subprocess.run(
    ["go", "test", "-count=1", "-coverprofile=coverage_e2ehelpers.out", pkg],
    capture_output=True, text=True
)
print("STDOUT:", result.stdout)
result2 = subprocess.run(
    ["go", "tool", "cover", "-func=coverage_e2ehelpers.out"],
    capture_output=True, text=True
)
print(result2.stdout)