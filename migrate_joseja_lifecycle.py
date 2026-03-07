path = "internal/apps/jose/ja/server/server_integration_test.go"

with open(path, "r", encoding="utf-8") as f:
    content = f.read()

# Get Lifecycle function
lc_start = content.find("func TestJoseJAServer_Lifecycle(")
lc_end = content.find("\nfunc ", lc_start + 1)
print("=== Lifecycle test ===")
print(repr(content[lc_start:lc_end]))