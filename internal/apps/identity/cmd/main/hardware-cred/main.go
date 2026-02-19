// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityORM "cryptoutil/internal/apps/identity/repository/orm"
)

const (
	commandEnroll    = "enroll"
	commandList      = "list"
	commandRevoke    = "revoke"
	commandRenew     = "renew"
	commandInventory = "inventory"
	commandHelp      = "help"
)

const (
	defaultDeviceName = "Unknown Device"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case commandEnroll:
		runEnroll(os.Args[2:])
	case commandList:
		runList(os.Args[2:])
	case commandRevoke:
		runRevoke(os.Args[2:])
	case commandRenew:
		runRenew(os.Args[2:])
	case commandInventory:
		runInventory(os.Args[2:])
	case commandHelp:
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// printUsage displays help text for all commands.
func printUsage() {
	fmt.Println("Hardware Credential CLI - Self-service enrollment and management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  hardware-cred <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  enroll    Enroll a new hardware credential (smart card, FIDO key)")
	fmt.Println("  list      List all enrolled hardware credentials for a user")
	fmt.Println("  revoke    Revoke a hardware credential by ID")
	fmt.Println("  renew     Renew/rotate a hardware credential with new key material")
	fmt.Println("  inventory Generate inventory report of all hardware credentials")
	fmt.Println("  help      Display this help message")
	fmt.Println()
	fmt.Println("Enroll Options:")
	fmt.Println("  -user-id <UUID>        User ID (required)")
	fmt.Println("  -device-name <string>  Device name (e.g., 'YubiKey 5 NFC', 'Smart Card Reader 1')")
	fmt.Println("  -credential-type <string>  Credential type (passkey, smart_card, security_key)")
	fmt.Println()
	fmt.Println("List Options:")
	fmt.Println("  -user-id <UUID>        User ID (required)")
	fmt.Println()
	fmt.Println("Revoke Options:")
	fmt.Println("  -credential-id <string>  Credential ID (required)")
	fmt.Println()
	fmt.Println("Renew Options:")
	fmt.Println("  -credential-id <string>  Credential ID to renew (required)")
	fmt.Println("  -device-name <string>    Updated device name (optional)")
	fmt.Println()
	fmt.Println("Inventory Options:")
	fmt.Println("  (no flags required)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  hardware-cred enroll -user-id 01930de8-c123-7890-abcd-ef1234567890 -device-name 'YubiKey 5C'")
	fmt.Println("  hardware-cred list -user-id 01930de8-c123-7890-abcd-ef1234567890")
	fmt.Println("  hardware-cred revoke -credential-id 'AQIDBAUGBwgJCgsMDQ4PEA=='")
	fmt.Println("  hardware-cred renew -credential-id 'AQIDBAUGBwgJCgsMDQ4PEA==' -device-name 'YubiKey 5C (Renewed)'")
	fmt.Println("  hardware-cred inventory")
	fmt.Println()
}

// runEnroll handles the enroll command.
func runEnroll(args []string) {
	fs := flag.NewFlagSet("enroll", flag.ExitOnError)
	userIDStr := fs.String("user-id", "", "User ID (UUID)")
	deviceName := fs.String("device-name", "", "Device name")
	credentialType := fs.String("credential-type", "passkey", "Credential type (passkey, smart_card, security_key)")

	err := fs.Parse(args)
	if err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	if *userIDStr == "" {
		fmt.Fprintln(os.Stderr, "Error: -user-id is required")
		fs.Usage()
		os.Exit(1)
	}

	userID, err := googleUuid.Parse(*userIDStr)
	if err != nil {
		log.Fatalf("Invalid user-id UUID: %v", err)
	}

	if *deviceName == "" {
		*deviceName = fmt.Sprintf("Device-%s", time.Now().UTC().Format("20060102-150405"))
	}

	ctx := context.Background()

	// Initialize database connection.
	db, err := initDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create credential repository.
	credRepo, err := cryptoutilIdentityORM.NewWebAuthnCredentialRepository(db)
	if err != nil {
		log.Fatalf("Failed to create credential repository: %v", err)
	}

	// Simulate hardware credential enrollment (mock implementation).
	fmt.Printf("Enrolling hardware credential for user %s...\n", userID)

	credentialID := generateMockCredentialID()

	credential := &cryptoutilIdentityORM.Credential{
		ID:              credentialID,
		UserID:          userID.String(),
		Type:            parseCredentialType(*credentialType),
		PublicKey:       generateMockPublicKey(),
		AttestationType: "none",
		AAGUID:          []byte{0x00, 0x00, 0x00, 0x00},
		SignCount:       0,
		Metadata: map[string]any{
			"device_name": *deviceName,
		},
	}

	err = credRepo.StoreCredential(ctx, credential)
	if err != nil {
		log.Fatalf("Failed to store credential: %v", err)
	}

	fmt.Println("✅ Enrollment successful")
	fmt.Printf("Credential ID: %s\n", credential.ID)
	fmt.Printf("Device Name: %s\n", *deviceName)
	fmt.Printf("Type: %s\n", credential.Type)
	fmt.Printf("Created At: %s\n", credential.CreatedAt.Format(time.RFC3339))

	// Audit log entry.
	logAuditEvent(ctx, "CREDENTIAL_ENROLLED", userID.String(), credential.ID, map[string]any{
		"device_name":     *deviceName,
		"credential_type": credential.Type,
		"attestation":     credential.AttestationType,
		"event_category":  "lifecycle",
		"compliance_flag": "hardware_credential_enrollment",
	})
}

// runList handles the list command.
func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	userIDStr := fs.String("user-id", "", "User ID (UUID)")

	err := fs.Parse(args)
	if err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	if *userIDStr == "" {
		fmt.Fprintln(os.Stderr, "Error: -user-id is required")
		fs.Usage()
		os.Exit(1)
	}

	userID, err := googleUuid.Parse(*userIDStr)
	if err != nil {
		log.Fatalf("Invalid user-id UUID: %v", err)
	}

	ctx := context.Background()

	// Initialize database connection.
	db, err := initDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create credential repository.
	credRepo, err := cryptoutilIdentityORM.NewWebAuthnCredentialRepository(db)
	if err != nil {
		log.Fatalf("Failed to create credential repository: %v", err)
	}

	// Retrieve user credentials.
	credentials, err := credRepo.GetUserCredentials(ctx, userID.String())
	if err != nil {
		log.Fatalf("Failed to retrieve credentials: %v", err)
	}

	if len(credentials) == 0 {
		fmt.Printf("No hardware credentials enrolled for user %s\n", userID)

		return
	}

	fmt.Printf("Hardware credentials for user %s:\n", userID)
	fmt.Println()

	for i, cred := range credentials {
		deviceName := defaultDeviceName
		if name, ok := cred.Metadata["device_name"].(string); ok {
			deviceName = name
		}

		fmt.Printf("%d. Credential ID: %s\n", i+1, cred.ID)
		fmt.Printf("   Device Name: %s\n", deviceName)
		fmt.Printf("   Type: %s\n", cred.Type)
		fmt.Printf("   Sign Count: %d\n", cred.SignCount)
		fmt.Printf("   Created At: %s\n", cred.CreatedAt.Format(time.RFC3339))
		fmt.Printf("   Last Used: %s\n", cred.LastUsedAt.Format(time.RFC3339))
		fmt.Println()
	}

	// Audit log entry.
	logAuditEvent(ctx, "CREDENTIALS_LISTED", userID.String(), "", map[string]any{
		"credential_count": len(credentials),
		"event_category":   "access",
		"compliance_flag":  "credential_inventory_access",
	})
}

// runRevoke handles the revoke command.
func runRevoke(args []string) {
	fs := flag.NewFlagSet("revoke", flag.ExitOnError)
	credentialID := fs.String("credential-id", "", "Credential ID (base64 URL-encoded)")

	err := fs.Parse(args)
	if err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	if *credentialID == "" {
		fmt.Fprintln(os.Stderr, "Error: -credential-id is required")
		fs.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	// Initialize database connection.
	db, err := initDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create credential repository.
	credRepo, err := cryptoutilIdentityORM.NewWebAuthnCredentialRepository(db)
	if err != nil {
		log.Fatalf("Failed to create credential repository: %v", err)
	}

	// Retrieve credential before deletion (for audit logging).
	credential, err := credRepo.GetCredential(ctx, *credentialID)
	if err != nil {
		if errors.Is(err, cryptoutilIdentityAppErr.ErrCredentialNotFound) {
			fmt.Fprintf(os.Stderr, "Error: Credential not found: %s\n", *credentialID)
			os.Exit(1)
		}

		log.Fatalf("Failed to retrieve credential: %v", err)
	}

	deviceName := defaultDeviceName
	if name, ok := credential.Metadata["device_name"].(string); ok {
		deviceName = name
	}

	// Delete credential.
	err = credRepo.DeleteCredential(ctx, *credentialID)
	if err != nil {
		log.Fatalf("Failed to revoke credential: %v", err)
	}

	fmt.Println("✅ Credential revoked successfully")
	fmt.Printf("Credential ID: %s\n", credential.ID)
	fmt.Printf("Device Name: %s\n", deviceName)
	fmt.Printf("Type: %s\n", credential.Type)

	// Audit log entry.
	logAuditEvent(ctx, "CREDENTIAL_REVOKED", credential.UserID, credential.ID, map[string]any{
		"device_name":     deviceName,
		"credential_type": credential.Type,
		"sign_count":      credential.SignCount,
		"last_used_at":    credential.LastUsedAt.Format(time.RFC3339),
		"event_category":  "lifecycle",
		"compliance_flag": "hardware_credential_revocation",
	})
}

// runRenew handles the renew command for credential rotation.
func runRenew(args []string) {
	fs := flag.NewFlagSet("renew", flag.ExitOnError)
	credentialID := fs.String("credential-id", "", "Credential ID to renew")
	deviceName := fs.String("device-name", "", "Updated device name (optional)")

	err := fs.Parse(args)
	if err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	if *credentialID == "" {
		fmt.Fprintln(os.Stderr, "Error: -credential-id is required")
		fs.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	// Initialize database connection.
	db, err := initDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create credential repository.
	credRepo, err := cryptoutilIdentityORM.NewWebAuthnCredentialRepository(db)
	if err != nil {
		log.Fatalf("Failed to create credential repository: %v", err)
	}

	// Retrieve existing credential.
	credential, err := credRepo.GetCredential(ctx, *credentialID)
	if err != nil {
		if errors.Is(err, cryptoutilIdentityAppErr.ErrCredentialNotFound) {
			fmt.Fprintf(os.Stderr, "Error: Credential not found: %s\n", *credentialID)
			os.Exit(1)
		}

		log.Fatalf("Failed to retrieve credential: %v", err)
	}

	// Generate new credential with rotated key material.
	newCredentialID := generateMockCredentialID()
	newPublicKey := generateMockPublicKey()

	newCredential := &cryptoutilIdentityORM.Credential{
		ID:              newCredentialID,
		UserID:          credential.UserID,
		Type:            credential.Type,
		PublicKey:       newPublicKey,
		AttestationType: credential.AttestationType,
		AAGUID:          credential.AAGUID,
		SignCount:       0,
		CreatedAt:       time.Now().UTC(),
		LastUsedAt:      time.Now().UTC(),
		Metadata:        credential.Metadata,
	}

	if *deviceName != "" {
		newCredential.Metadata["device_name"] = *deviceName
	}

	// Store new credential.
	err = credRepo.StoreCredential(ctx, newCredential)
	if err != nil {
		log.Fatalf("Failed to store renewed credential: %v", err)
	}

	// Delete old credential (rotation completes).
	err = credRepo.DeleteCredential(ctx, *credentialID)
	if err != nil {
		log.Printf("Warning: Failed to delete old credential: %v", err)
	}

	oldDeviceName := defaultDeviceName
	if name, ok := credential.Metadata["device_name"].(string); ok {
		oldDeviceName = name
	}

	newDeviceName := oldDeviceName
	if *deviceName != "" {
		newDeviceName = *deviceName
	}

	fmt.Println("✅ Credential renewed successfully")
	fmt.Printf("Old Credential ID: %s\n", credential.ID)
	fmt.Printf("New Credential ID: %s\n", newCredential.ID)
	fmt.Printf("User ID: %s\n", credential.UserID)
	fmt.Printf("Device Name: %s → %s\n", oldDeviceName, newDeviceName)
	fmt.Printf("Type: %s\n", newCredential.Type)

	// Audit log entry.
	logAuditEvent(ctx, "CREDENTIAL_RENEWED", credential.UserID, newCredential.ID, map[string]any{
		"old_credential_id": credential.ID,
		"new_credential_id": newCredential.ID,
		"old_device_name":   oldDeviceName,
		"new_device_name":   newDeviceName,
		"credential_type":   credential.Type,
		"event_category":    "lifecycle",
		"compliance_flag":   "hardware_credential_renewal",
	})
}

// runInventory generates an inventory report of all hardware credentials.
