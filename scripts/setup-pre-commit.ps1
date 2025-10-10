# Pre-commit setup script for PowerShell
# Ensures consistent cache location and proper installation

Write-Host "Setting up pre-commit hooks..."

# Set consistent cache location
$cachePath = "$env:USERPROFILE\.cache\pre-commit"
[Environment]::SetEnvironmentVariable("PRE_COMMIT_HOME", $cachePath, "User")
Write-Host "Set PRE_COMMIT_HOME to: $cachePath"

# Install pre-commit if not already installed
pip install pre-commit

# Install the hooks
pre-commit install

# Test the setup
pre-commit run --all-files

Write-Host "Pre-commit setup complete!"
Write-Host "Cache location: $cachePath"
