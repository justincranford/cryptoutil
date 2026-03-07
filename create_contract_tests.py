import os

def write(path, lines):
    with open(path, 'w', encoding='utf-8', newline='\n') as f:
        f.write('\n'.join(lines) + '\n')
    print('Created ' + path)

base = 'internal/apps/template/service/testing/contract/'

# contracts_test.go
write(base + 'contracts_test.go', [
    '// Copyright (c) 2025 Justin Cranford',
    '//',
    '// Tests for RunContractTests - verifies the full contract suite against the test server.',
    'package contract',
    '',
    'import "testing"',
    '',
    'func TestRunContractTests(t *testing.T) {',
    '\tt.Parallel()',
    '',
    '\tRunContractTests(t, testContractServer)',
    '}',
])

# health_contracts_test.go
write(base + 'health_contracts_test.go', [
    '// Copyright (c) 2025 Justin Cranford',
    '//',
    '// Tests for RunHealthContracts and RunReadyzNotReadyContract.',
    'package contract',
    '',
    'import "testing"',
    '',
    'func TestRunHealthContracts(t *testing.T) {',
    '\tt.Parallel()',
    '',
    '\tRunHealthContracts(t, testContractServer)',
    '}',
    '',
    '// TestRunReadyzNotReadyContract tests that readyz returns 503 when server is not ready.',
    '// Not parallel: temporarily modifies server ready state.',
    'func TestRunReadyzNotReadyContract(t *testing.T) {',
    '\tRunReadyzNotReadyContract(t, testContractServer)',
    '}',
])

# server_contracts_test.go
write(base + 'server_contracts_test.go', [
    '// Copyright (c) 2025 Justin Cranford',
    '//',
    '// Tests for RunServerContracts.',
    'package contract',
    '',
    'import "testing"',
    '',
    'func TestRunServerContracts(t *testing.T) {',
    '\tt.Parallel()',
    '',
    '\tRunServerContracts(t, testContractServer)',
    '}',
])

# response_contracts_test.go
write(base + 'response_contracts_test.go', [
    '// Copyright (c) 2025 Justin Cranford',
    '//',
    '// Tests for RunResponseFormatContracts.',
    'package contract',
    '',
    'import "testing"',
    '',
    'func TestRunResponseFormatContracts(t *testing.T) {',
    '\tt.Parallel()',
    '',
    '\tRunResponseFormatContracts(t, testContractServer)',
    '}',
])