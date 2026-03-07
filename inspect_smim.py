path = "internal/apps/sm/im/server/server_test.go"
with open(path, "r", encoding="utf-8") as f:
    content = f.read()
print("Length:", len(content))
# Print imports
imp_start = content.find("import (")
imp_end = content.find("\n)\n", imp_start)
print("=== Imports ===")
print(repr(content[imp_start:imp_end+3]))
# Print TestServer_Shutdown
ts_start = content.find("func TestServer_Shutdown(")
ts_end = content.find("\nfunc ", ts_start + 1)
if ts_end < 0:
    ts_end = len(content)
print("=== TestServer_Shutdown ===")
print(repr(content[ts_start:ts_end]))