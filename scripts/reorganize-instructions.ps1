# Reorganize instruction files
# This script backs up original files and creates refactored versions

$instructionsDir = "c:\Dev\Projects\cryptoutil\.github\instructions"
$backupDir = "c:\Dev\Projects\cryptoutil\.github\instructions_backup_$(Get-Date -Format 'yyyyMMdd_HHmmss')"

# Create backup
Write-Host "Creating backup at: $backupDir"
Copy-Item -Path $instructionsDir -Destination $backupDir -Recurse

Write-Host "Backup completed. Original files preserved."
Write-Host "Proceeding with reorganization..."
