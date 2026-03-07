path = "internal/apps/sm/kms/server/testmain_integration_test.go"
with open(path, "r", encoding="utf-8") as f:
    content = f.read()
# Find exact context after DualPortBaseURLs line for the healthclient insertion
dualport_idx = content.find("\ttestIntegrationPublicURL, testIntegrationAdminURL = ")
line_end = content.find("\n", dualport_idx)
next_lines = content[line_end:line_end+200]
print("After DualPortBaseURLs line:")
print(repr(next_lines))