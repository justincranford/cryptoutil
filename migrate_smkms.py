path = "internal/apps/sm/kms/server/testmain_integration_test.go"
with open(path, "r", encoding="utf-8") as f:
    content = f.read()

# 1. Add healthclient import
old_imports = 'import (\n\t"context"\n\t"crypto/tls"\n\tcryptoutilSharedMagic "cryptoutil/internal/shared/magic"\n\t"fmt"\n\thttp "net/http"\n\t"os"\n\t"testing"\n\t"time"\n\n\tcryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"\n\tcryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"\n)'
new_imports = 'import (\n\t"context"\n\t"crypto/tls"\n\tcryptoutilSharedMagic "cryptoutil/internal/shared/magic"\n\t"fmt"\n\thttp "net/http"\n\t"os"\n\t"testing"\n\t"time"\n\n\tcryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"\n\tcryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"\n\tcryptoutilTestingHealthclient "cryptoutil/internal/apps/template/service/testing/healthclient"\n)'
if old_imports in content:
    content = content.replace(old_imports, new_imports)
    print("Replaced imports OK")
else:
    print("ERROR: imports not found")

# 2. Add healthclient var
old_var = 'var (\n\ttestIntegrationServer    *KMSServer\n\ttestIntegrationClient    *http.Client\n\ttestIntegrationPublicURL string\n\ttestIntegrationAdminURL  string\n)'
new_var = 'var (\n\ttestIntegrationServer       *KMSServer\n\ttestIntegrationClient       *http.Client\n\ttestIntegrationHealthClient *cryptoutilTestingHealthclient.HealthClient\n\ttestIntegrationPublicURL   string\n\ttestIntegrationAdminURL    string\n)'
if old_var in content:
    content = content.replace(old_var, new_var)
    print("Replaced var block OK")
else:
    print("ERROR: var block not found")

# 3. Add healthclient init after DualPortBaseURLs
old_init = '\ttestIntegrationPublicURL, testIntegrationAdminURL = cryptoutilAppsTemplateServiceTestingE2eHelpers.DualPortBaseURLs(testIntegrationServer)\n\n\t// Create HTTP client that accepts self-signed certificates.'
new_init = '\ttestIntegrationPublicURL, testIntegrationAdminURL = cryptoutilAppsTemplateServiceTestingE2eHelpers.DualPortBaseURLs(testIntegrationServer)\n\n\t// Create shared health client.\n\ttestIntegrationHealthClient = cryptoutilTestingHealthclient.NewHealthClient(testIntegrationPublicURL, testIntegrationAdminURL)\n\n\t// Create HTTP client that accepts self-signed certificates.'
if old_init in content:
    content = content.replace(old_init, new_init)
    print("Added healthclient init OK")
else:
    print("ERROR: init insertion point not found")

with open(path, "w", encoding="utf-8", newline="\n") as f:
    f.write(content)
print("File written, length:", len(content))