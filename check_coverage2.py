import subprocess

result = subprocess.run(
    ["go", "tool", "cover", "-func=coverage_testdb.out"],
    capture_output=True, text=True
)
print(result.stdout)