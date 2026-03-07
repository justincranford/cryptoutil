path = "internal/apps/sm/kms/server/testmain_integration_test.go"
with open(path, "r", encoding="utf-8") as f:
    content = f.read()
print("Length:", len(content))
imp_start = content.find("import (")
imp_end = content.find("\n)\n", imp_start)
print("=== Imports ===")
print(repr(content[imp_start:imp_end+3]))
# Find var block
var_start = content.find("\nvar (")
var_end = content.find("\n)\n", var_start)
print("=== Var block ===")
print(repr(content[var_start:var_end+3]))
# Find the DualPortBaseURLs line for init
dualport_idx = content.find("DualPortBaseURLs(")
print("=== Context around DualPortBaseURLs ===")
print(repr(content[dualport_idx-5:dualport_idx+200]))