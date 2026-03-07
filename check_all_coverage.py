import subprocess

result = subprocess.run(
    ["go", "test", "-count=1", "-coverprofile=coverage_all.out",
     "./internal/apps/template/service/testing/assertions/",
     "./internal/apps/template/service/testing/fixtures/",
     "./internal/apps/template/service/testing/healthclient/",
     "./internal/apps/template/service/testing/testserver/"],
    capture_output=True, text=True
)
print("STDOUT:", result.stdout)

result2 = subprocess.run(
    ["go", "tool", "cover", "-func=coverage_all.out"],
    capture_output=True, text=True
)
print("COVERAGE:")
for line in result2.stdout.splitlines():
    print(" ", line.strip())