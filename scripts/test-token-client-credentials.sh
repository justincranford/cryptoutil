#!/bin/bash
# Test Token Endpoint - Client Credentials Grant
# This script tests the /oauth2/v1/token endpoint with client_credentials grant

# CRITICAL: This is for CI/CD workflow context only (NOT Windows PowerShell local commands)
# For local testing on Windows, use: Invoke-WebRequest instead of curl

# Test client_credentials grant
curl -X POST http://127.0.0.1:8080/oauth2/v1/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=demo-client" \
  -d "client_secret=demo-secret" \
  -d "scope=openid profile email read write"

# Expected response:
# {
#   "access_token": "eyJhbG...",
#   "token_type": "Bearer",
#   "expires_in": 3600,
#   "scope": "openid profile email read write"
# }

# PowerShell equivalent (for local Windows testing):
# $body = @{
#     grant_type = "client_credentials"
#     client_id = "demo-client"
#     client_secret = "demo-secret"
#     scope = "openid profile email read write"
# }
# Invoke-WebRequest -Method POST -Uri "http://127.0.0.1:8080/oauth2/v1/token" `
#   -ContentType "application/x-www-form-urlencoded" `
#   -Body $body | Select-Object -ExpandProperty Content
