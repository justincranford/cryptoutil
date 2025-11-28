# Test Token Endpoint - Client Credentials (PowerShell)
# This script tests the /oauth2/v1/token endpoint with client_credentials grant

$body = @{
    grant_type = "client_credentials"
    client_id = "demo-client"
    client_secret = "demo-secret"
    scope = "openid profile email read write"
}

try {
    $response = Invoke-WebRequest -Method POST -Uri "http://127.0.0.1:8090/oauth2/v1/token" `
        -ContentType "application/x-www-form-urlencoded" `
        -Body $body

    Write-Host "Status Code: $($response.StatusCode)" -ForegroundColor Green
    Write-Host ""
    Write-Host "Response:" -ForegroundColor Cyan
    $response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 5
} catch {
    Write-Host "Request failed!" -ForegroundColor Red
    Write-Host "Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Yellow
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow

    if ($_.ErrorDetails.Message) {
        Write-Host ""
        Write-Host "Error Details:" -ForegroundColor Cyan
        Write-Host $_.ErrorDetails.Message
    }
}

# Expected successful response:
# {
#   "access_token": "eyJhbG...",
#   "token_type": "Bearer",
#   "expires_in": 3600,
#   "scope": "openid profile email read write"
# }
