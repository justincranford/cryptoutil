# Archive Consolidation Script
# Purpose: Delete archive files after verifying content is covered in copilot instructions

$archiveRoot = "c:\Dev\Projects\cryptoutil\docs\archive"

# Files to delete (content fully covered)
$filesToDelete = @(
    "$archiveRoot\CGO-BAN-ENFORCEMENT.md"
)

foreach ($file in $filesToDelete) {
    if (Test-Path $file) {
        Remove-Item $file -Force
        Write-Host "✅ Deleted: $file"
    } else {
        Write-Host "⚠️  File not found: $file"
    }
}

# Verify deletions
Write-Host "`n=== Verification ==="
foreach ($file in $filesToDelete) {
    if (Test-Path $file) {
        Write-Host "❌ FAILED to delete: $file"
    } else {
        Write-Host "✅ Confirmed deleted: $file"
    }
}
