## RAW OUTPUT

PS C:\Dev\Projects\cryptoutil> go test ./... -cover
?       cryptoutil/api  [no test files]
        cryptoutil/api/authz            coverage: 0.0% of statements
?       cryptoutil/api/ca       [no test files]
        cryptoutil/api/ca/models                coverage: 0.0% of statements
        cryptoutil/api/ca/server                coverage: 0.0% of statements
        cryptoutil/api/client           coverage: 0.0% of statements
?       cryptoutil/api/identity [no test files]
        cryptoutil/api/identity/authz           coverage: 0.0% of statements
        cryptoutil/api/identity/idp             coverage: 0.0% of statements
        cryptoutil/api/identity/rs              coverage: 0.0% of statements
        cryptoutil/api/idp              coverage: 0.0% of statements
        cryptoutil/api/model            coverage: 0.0% of statements
        cryptoutil/api/server           coverage: 0.0% of statements
        cryptoutil/cmd/cicd             coverage: 0.0% of statements
        cryptoutil/cmd/cryptoutil               coverage: 0.0% of statements
        cryptoutil/cmd/demo             coverage: 0.0% of statements
        cryptoutil/cmd/identity-compose         coverage: 0.0% of statements
        cryptoutil/cmd/identity-demo            coverage: 0.0% of statements
        cryptoutil/cmd/identity-unified         coverage: 0.0% of statements
        cryptoutil/cmd/jose-server              coverage: 0.0% of statements
        cryptoutil/cmd/workflow         coverage: 0.0% of statements
ok      cryptoutil/internal/ca/api/handler      0.917s  coverage: 1.2% of statements
ok      cryptoutil/internal/ca/bootstrap        1.187s  coverage: 80.8% of statements
ok      cryptoutil/internal/ca/cli      (cached)        coverage: 79.6% of statements
ok      cryptoutil/internal/ca/compliance       (cached)        coverage: 86.4% of statements
ok      cryptoutil/internal/ca/config   (cached)        coverage: 87.2% of statements
ok      cryptoutil/internal/ca/crypto   2.301s  coverage: 92.1% of statements
?       cryptoutil/internal/ca/domain   [no test files]
ok      cryptoutil/internal/ca/intermediate     1.411s  coverage: 80.0% of statements
?       cryptoutil/internal/ca/magic    [no test files]
ok      cryptoutil/internal/ca/observability    (cached)        coverage: 96.9% of statements
ok      cryptoutil/internal/ca/profile/certificate      (cached)        coverage: 91.5% of statements
ok      cryptoutil/internal/ca/profile/subject  (cached)        coverage: 85.8% of statements
ok      cryptoutil/internal/ca/security 1.069s  coverage: 82.7% of statements
ok      cryptoutil/internal/ca/service/issuer   0.666s  coverage: 84.4% of statements
ok      cryptoutil/internal/ca/service/ra       (cached)        coverage: 88.3% of statements
ok      cryptoutil/internal/ca/service/revocation       0.568s  coverage: 78.9% of statements
ok      cryptoutil/internal/ca/service/timestamp        0.613s  coverage: 98.7% of statements
ok      cryptoutil/internal/ca/storage  (cached)        coverage: 89.9% of statements
ok      cryptoutil/internal/cmd/cicd    0.455s  coverage: 51.5% of statements
ok      cryptoutil/internal/cmd/cicd/adaptive-sim       1.533s  coverage: 63.0% of statements
ok      cryptoutil/internal/cmd/cicd/common     0.552s  coverage: 100.0% of statements
ok      cryptoutil/internal/cmd/cicd/format_go  1.271s  coverage: 52.9% of statements
ok      cryptoutil/internal/cmd/cicd/format_gotest      1.102s  coverage: 81.4% of statements
        cryptoutil/internal/cmd/cicd/identity_requirements_check                coverage: 0.0% of statements
ok      cryptoutil/internal/cmd/cicd/lint_go    0.895s  coverage: 34.5% of statements
ok      cryptoutil/internal/cmd/cicd/lint_go_mod        0.462s  coverage: 43.2% of statements
ok      cryptoutil/internal/cmd/cicd/lint_gotest        0.379s  coverage: 86.6% of statements
ok      cryptoutil/internal/cmd/cicd/lint_text  0.475s  coverage: 97.3% of statements
ok      cryptoutil/internal/cmd/cicd/lint_workflow      0.411s  coverage: 28.5% of statements
ok      cryptoutil/internal/cmd/cryptoutil      1.003s  coverage: 18.8% of statements
        cryptoutil/internal/cmd/demo            coverage: 0.0% of statements
ok      cryptoutil/internal/cmd/workflow        0.456s  coverage: 8.1% of statements
ok      cryptoutil/internal/common/apperr       (cached)        coverage: 27.6% of statements
ok      cryptoutil/internal/common/config       0.487s  coverage: 77.2% of statements
        cryptoutil/internal/common/container            coverage: 0.0% of statements
ok      cryptoutil/internal/common/crypto/asn1  1.819s  coverage: 88.7% of statements
ok      cryptoutil/internal/common/crypto/certificate   2.015s  coverage: 77.7% of statements
ok      cryptoutil/internal/common/crypto/digests       0.571s  coverage: 97.7% of statements
ok      cryptoutil/internal/common/crypto/keygen        6.394s  coverage: 85.2% of statements
ok      cryptoutil/internal/common/crypto/keygenpooltest        2.235s  coverage: 0.0% of statements
        cryptoutil/internal/common/magic                coverage: 0.0% of statements
ok      cryptoutil/internal/common/pool 0.828s  coverage: 62.1% of statements
ok      cryptoutil/internal/common/telemetry    1.141s  coverage: 67.5% of statements
ok      cryptoutil/internal/common/testutil     1.399s  coverage: 97.3% of statements
ok      cryptoutil/internal/common/util (cached)        coverage: 95.3% of statements
ok      cryptoutil/internal/common/util/combinations    (cached)        coverage: 100.0% of statements
ok      cryptoutil/internal/common/util/datetime        (cached)        coverage: 100.0% of statements
ok      cryptoutil/internal/common/util/files   0.512s  coverage: 88.9% of statements
ok      cryptoutil/internal/common/util/network 0.419s  coverage: 22.6% of statements
ok      cryptoutil/internal/common/util/sysinfo 1.920s  coverage: 84.4% of statements
ok      cryptoutil/internal/common/util/thread  0.434s  coverage: 100.0% of statements
ok      cryptoutil/internal/crypto      3.482s  coverage: 94.4% of statements
        cryptoutil/internal/identity/apperr             coverage: 0.0% of statements
ok      cryptoutil/internal/identity/authz      19.248s coverage: 77.2% of statements
ok      cryptoutil/internal/identity/authz/clientauth   168.383s        coverage: 78.4% of statements
ok      cryptoutil/internal/identity/authz/e2e  3.774s  coverage: [no statements]
ok      cryptoutil/internal/identity/authz/pkce (cached)        coverage: 95.5% of statements
ok      cryptoutil/internal/identity/bootstrap  3.620s  coverage: 79.1% of statements
        cryptoutil/internal/identity/cmd                coverage: 0.0% of statements
        cryptoutil/internal/identity/cmd/main           coverage: 0.0% of statements
        cryptoutil/internal/identity/cmd/main/authz             coverage: 0.0% of statements
ok      cryptoutil/internal/identity/cmd/main/hardware-cred     0.558s  coverage: 19.3% of statements
        cryptoutil/internal/identity/cmd/main/idp               coverage: 0.0% of statements
        cryptoutil/internal/identity/cmd/main/rs                coverage: 0.0% of statements
        cryptoutil/internal/identity/cmd/main/spa-rp            coverage: 0.0% of statements
ok      cryptoutil/internal/identity/config     2.277s  coverage: 70.1% of statements
ok      cryptoutil/internal/identity/domain     (cached)        coverage: 87.4% of statements
ok      cryptoutil/internal/identity/healthcheck        4.073s  coverage: 85.3% of statements
ok      cryptoutil/internal/identity/idp        15.381s coverage: 54.9% of statements
        cryptoutil/internal/identity/idp/auth           coverage: 0.0% of statements
ok      cryptoutil/internal/identity/idp/userauth       4.699s  coverage: 37.1% of statements
ok      cryptoutil/internal/identity/idp/userauth/mocks (cached)        coverage: 84.1% of statements
ok      cryptoutil/internal/identity/integration        0.423s  coverage: [no statements] [no tests to run]
ok      cryptoutil/internal/identity/issuer     2.955s  coverage: 59.6% of statements
ok      cryptoutil/internal/identity/jobs       7.448s  coverage: 89.0% of statements
ok      cryptoutil/internal/identity/jwks       1.373s  coverage: 77.5% of statements
?       cryptoutil/internal/identity/magic      [no test files]
ok      cryptoutil/internal/identity/notifications      0.966s  coverage: 87.8% of statements
        cryptoutil/internal/identity/process            coverage: 0.0% of statements
        cryptoutil/internal/identity/repository         coverage: 0.0% of statements
ok      cryptoutil/internal/identity/repository/orm     3.552s  coverage: 67.5% of statements
ok      cryptoutil/internal/identity/rotation   7.674s  coverage: 83.7% of statements
ok      cryptoutil/internal/identity/rs 0.817s  coverage: 76.4% of statements
ok      cryptoutil/internal/identity/security   (cached)        coverage: 100.0% of statements
        cryptoutil/internal/identity/server             coverage: 0.0% of statements
        cryptoutil/internal/identity/storage/fixtures           coverage: 0.0% of statements
ok      cryptoutil/internal/identity/storage/tests      2.524s  coverage: [no statements]
ok      cryptoutil/internal/identity/test/contract      0.743s  coverage: [no statements]
ok      cryptoutil/internal/identity/test/integration   16.370s coverage: [no statements]
ok      cryptoutil/internal/identity/test/load  (cached)        coverage: [no statements]
        cryptoutil/internal/identity/test/testutils             coverage: 0.0% of statements
ok      cryptoutil/internal/identity/test/unit  17.896s coverage: [no statements]
ok      cryptoutil/internal/infra/demo  1.128s  coverage: 81.8% of statements
ok      cryptoutil/internal/infra/realm 13.787s coverage: 85.6% of statements
ok      cryptoutil/internal/infra/tenant        4.174s  coverage: 66.1% of statements
ok      cryptoutil/internal/infra/tls   1.285s  coverage: 85.1% of statements
ok      cryptoutil/internal/infra/tls/hsm       (cached)        coverage: 100.0% of statements
ok      cryptoutil/internal/jose        67.003s coverage: 48.8% of statements
ok      cryptoutil/internal/jose/example        1.588s  coverage: [no statements]
ok      cryptoutil/internal/jose/server 94.342s coverage: 56.1% of statements
        cryptoutil/internal/jose/server/cmd             coverage: 0.0% of statements
ok      cryptoutil/internal/kms/client  73.859s coverage: 76.2% of statements
        cryptoutil/internal/kms/cmd             coverage: 0.0% of statements
ok      cryptoutil/internal/kms/server/application      27.596s coverage: 64.7% of statements
ok      cryptoutil/internal/kms/server/barrier  12.559s coverage: 75.5% of statements
ok      cryptoutil/internal/kms/server/barrier/contentkeysservice       3.829s  coverage: 81.2% of statements
ok      cryptoutil/internal/kms/server/barrier/intermediatekeysservice  4.046s  coverage: 77.7% of statements
ok      cryptoutil/internal/kms/server/barrier/rootkeysservice  4.093s  coverage: 79.0% of statements
ok      cryptoutil/internal/kms/server/barrier/unsealkeysservice        1.855s  coverage: 49.4% of statements
ok      cryptoutil/internal/kms/server/businesslogic    5.407s  coverage: 39.4% of statements
ok      cryptoutil/internal/kms/server/demo     0.849s  coverage: 7.3% of statements
ok      cryptoutil/internal/kms/server/handler  1.200s  coverage: 79.1% of statements
ok      cryptoutil/internal/kms/server/middleware       0.854s  coverage: 53.1% of statements
ok      cryptoutil/internal/kms/server/repository/orm   2.578s  coverage: 90.8% of statements
