import subprocess

result = subprocess.run(
    ["go", "test", "-count=1", "-coverprofile=coverage_e2einfra.out",
     "./internal/apps/template/service/testing/e2e_infra/"],
    capture_output=True, text=True
)
print("STDOUT:", result.stdout)
print("STDERR:", result.stderr[:200] if result.stderr else "")

result2 = subprocess.run(
    ["go", "tool", "cover", "-func=coverage_e2einfra.out"],
    capture_output=True, text=True
)
print("COVERAGE:")
print(result2.stdout)