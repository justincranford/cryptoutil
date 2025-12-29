#!/usr/bin/env python3
"""Replace all hardcoded passwords in test files with random generation."""

import re
from pathlib import Path

def fix_register_test():
    """Fix register_test.go"""
    file_path = Path("internal/learn/server/register_test.go")
    content = file_path.read_text(encoding='utf-8')

    # Already has password variable in first test, just need to fix DuplicateUsername test
    content = re.sub(
        r'reqBody := map\[string\]string\{\s*"username": "duplicate",\s*"password": "password123",\s*\}',
        '''password1, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	reqBody := map[string]string{
		"username": "duplicate",
		"password": password1,
	}''',
        content,
        flags=re.DOTALL
    )

    file_path.write_text(content, encoding='utf-8')
    print(f"✓ Fixed {file_path}")

def fix_login_test():
    """Fix login_test.go"""
    file_path = Path("internal/learn/server/login_test.go")
    content = file_path.read_text(encoding='utf-8')

    # Add import
    if 'cryptoutilRandom' not in content:
        content = content.replace(
            'import (\n\t"bytes"\n\t"context"\n\t"encoding/hex"\n\t"encoding/json"\n\t"fmt"\n\t"net/http"\n\t"testing"\n\t"time"\n\n\tgoogleUuid "github.com/google/uuid"\n\t"github.com/stretchr/testify/require"\n\n\tcryptoutilLearnCrypto "cryptoutil/internal/learn/crypto"\n\tcryptoutilLearnDomain "cryptoutil/internal/learn/domain"\n\t"cryptoutil/internal/learn/repository"\n)',
            'import (\n\t"bytes"\n\t"context"\n\t"encoding/hex"\n\t"encoding/json"\n\t"fmt"\n\t"net/http"\n\t"testing"\n\t"time"\n\n\tgoogleUuid "github.com/google/uuid"\n\t"github.com/stretchr/testify/require"\n\n\tcryptoutilLearnCrypto "cryptoutil/internal/learn/crypto"\n\tcryptoutilLearnDomain "cryptoutil/internal/learn/domain"\n\t"cryptoutil/internal/learn/repository"\n\tcryptoutilRandom "cryptoutil/internal/shared/util/random"\n)'
        )

    # Replace all "password123" literals
    content = content.replace('"password123"', 'password')
    content = content.replace('"wrongpassword123"', 'wrongPassword')
    content = content.replace('"CorrectPassword123"', 'password')

    file_path.write_text(content, encoding='utf-8')
    print(f"✓ Fixed {file_path}")

def fix_send_test():
    """Fix send_test.go"""
    file_path = Path("internal/learn/server/send_test.go")
    content = file_path.read_text(encoding='utf-8')

    # Already uses GenerateUsernameSimple/PasswordSimple
    # Just replace the one "password123" literal for receiver registration
    content = content.replace(
        'receiver := registerTestUser(t, client, baseURL, "receiver", "password123")',
        '''receiverPassword, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)
	receiver := registerTestUser(t, client, baseURL, "receiver", receiverPassword)'''
    )

    file_path.write_text(content, encoding='utf-8')
    print(f"✓ Fixed {file_path}")

def fix_receive_delete_test():
    """Fix receive_delete_test.go"""
    file_path = Path("internal/learn/server/receive_delete_test.go")
    content = file_path.read_text(encoding='utf-8')

    # Replace "password123" for receiver
    content = content.replace(
        'receiver := registerTestUser(t, client, baseURL, "receiver", "password123")',
        '''receiverPassword, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)
	receiver := registerTestUser(t, client, baseURL, "receiver", receiverPassword)'''
    )

    file_path.write_text(content, encoding='utf-8')
    print(f"✓ Fixed {file_path}")

def fix_crypto_password_test():
    """Fix crypto/password_test.go"""
    file_path = Path("internal/learn/crypto/password_test.go")
    content = file_path.read_text(encoding='utf-8')

    # Add import if needed
    if 'cryptoutilRandom' not in content:
        content = content.replace(
            'import (\n\t"testing"\n\n\t"github.com/stretchr/testify/require"\n)',
            'import (\n\t"testing"\n\n\t"github.com/stretchr/testify/require"\n\n\tcryptoutilRandom "cryptoutil/internal/shared/util/random"\n)'
        )

    # These passwords are test data for crypto operations - generate them dynamically
    content = re.sub(
        r'password := "MySecurePassword123!"',
        '''password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)''',
        content
    )

    content = re.sub(
        r'password := "CorrectPassword123"',
        '''password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)''',
        content
    )

    content = re.sub(
        r'password := "TestPassword"',
        '''password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)''',
        content
    )

    file_path.write_text(content, encoding='utf-8')
    print(f"✓ Fixed {file_path}")

if __name__ == "__main__":
    print("Fixing hardcoded passwords in test files...")
    fix_register_test()
    fix_login_test()
    fix_send_test()
    fix_receive_delete_test()
    fix_crypto_password_test()
    print("\n✅ All test files fixed!")
