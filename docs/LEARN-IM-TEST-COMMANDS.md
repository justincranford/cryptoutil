
# ========================================

# learn-im Test Commands

# ========================================

# 1. Unit Tests with Coverage

go test ./internal/learn/... -short -coverprofile=./test-output/coverage_learn_unit.out
go tool cover -html=./test-output/coverage_learn_unit.out -o ./test-output/coverage_learn_unit.html

# 2. Integration Tests with Coverage (if internal/learn/integration exists)

go test ./internal/learn/integration/... -coverprofile=./test-output/coverage_learn_integration.out
go tool cover -html=./test-output/coverage_learn_integration.out -o ./test-output/coverage_learn_integration.html

# 3. E2E Tests with Coverage

go test ./internal/learn/e2e/... -coverprofile=./test-output/coverage_learn_e2e.out
go tool cover -html=./test-output/coverage_learn_e2e.out -o ./test-output/coverage_learn_e2e.html

# 4. Docker Compose (Start/Use/Stop)

docker compose -f deployments/learn/compose.yml up -d
docker compose -f deployments/learn/compose.yml ps
docker compose -f deployments/learn/compose.yml logs -f learn-im
curl -k <https://localhost:8888/service/api/v1/users/register> -H 'Content-Type: application/json' -d '{\"username\":\"alice\",\"password\":\"SecurePass123!\"}'
docker compose -f deployments/learn/compose.yml down

# 5. Demo Application (Start/Use/Stop)

go run ./cmd/learn-im -d

# In separate terminal

curl -k <https://localhost:8888/service/api/v1/users/register> -H 'Content-Type: application/json' -d '{\"username\":\"bob\",\"password\":\"SecurePass456!\"}'

# Stop with Ctrl+C
