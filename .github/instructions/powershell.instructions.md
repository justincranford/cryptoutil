---
description: "Instructions for PowerShell usage on Windows"
applyTo: "**"
---
# PowerShell Instructions

## Execution policy: preferred invocation

- ALWAYS prefer a one-shot, process-scoped bypass when running bundled helper scripts. This avoids permanently weakening machine policies and works reliably on systems where script execution is restricted.

```powershell
# Recommended (one-shot, no persistent policy change)
powershell -NoProfile -ExecutionPolicy Bypass -File script.ps1 -ScanProfile quick -Timeout 900
```

- Alternative (session-scoped, safe for interactive runs):

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
.\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 900
```

Notes:
- The first form launches a new PowerShell process with ExecutionPolicy bypassed for that process only. It's the safest and most repeatable approach for automation and CI helpers.
- Avoid changing `-Scope LocalMachine` or `-Scope CurrentUser` unless you understand the security implications.

## Scripting best-practices

- Use PowerShell syntax for Windows terminal commands (not Bash).
- Use `;` for chaining, `\` for paths, and `$env:VAR` for environment variables.
- Use `| Select-Object -First 10` for head-like behavior and `| Select-String` for grep-like searches.
- Avoid emojis or complex Unicode in here-strings — they can cause parsing or encoding issues.

## Common mistakes to avoid

- Switch parameter defaults: avoid `[switch]$All = $true`; prefer explicit logic to set defaults.
- Here-strings: avoid complex Unicode characters inside `@"..."@`.
- Variable expansion in paths: use `${variable}` for clarity, e.g. `"${PWD}\${OutputDir}"`.
- Prefer here-strings over complex backtick concatenation for multi-line text.
- Validate script parameters (test help and parameter validation) before use.
- Use proper error handling and exit codes in scripts.

## PowerShell gotchas (short, actionable)

1) Don't use `&&` / `||` chaining in PowerShell v5.1

	 - Problem: `cmd`/bash-style `&&` and `||` are not supported in older PowerShell (v5.1) and will produce a parser error.
	 - Safe alternatives:
		 - Put commands on separate lines in a script step (recommended).
		 - Use `;` as a simple separator for unrelated commands: `cmd1; cmd2`.
		 - For conditional execution mimic `&&` with `if ($LASTEXITCODE -eq 0) { cmd2 }`.

	 Example (safe):
	 ```powershell
	 git add .github/workflows/dast.yml
	 if ($LASTEXITCODE -eq 0) { git commit -m 'msg' }
	 ```

2) Avoid fragile one-liners with nested interpolation and unescaped `$` or `{}`

	 - Problem: PowerShell expands `$var` and `${}` inside double-quoted strings; complex one-liners with `"` and `${}` frequently cause parser errors.
	 - Safe alternatives:
		 - Use single-quoted strings when you don't want interpolation: `'literal $value'`.
		 - Use here-strings for multi-line commands to avoid escaping: `@" ... "@` or `@' ... '@`.
		 - Put complex logic into a small `.ps1` script and call it from the one-liner.

	 Example (safe here-string):
	 ```powershell
	 $script = @'
	 $lines = Get-Content .github/workflows/dast.yml
	 $lines[440..490] | ForEach-Object { Write-Output $_ }
	 '@
	 Invoke-Expression $script
	 ```

3) Use robust file-slicing/listing (avoid Select-Object -Index with a range string)

	 - Problem: Trying to pass a range like `440..490` incorrectly to `Select-Object -Index` or using unescaped expressions inside a quoted one-liner will fail.
	 - Safe alternatives:
		 - Read the file into a variable and slice the array: `$lines = Get-Content path; $lines[440..490]`.
		 - Use a simple `for` loop with numeric indices if you need line numbers.

	 Example (safe):
	 ```powershell
	 $lines = Get-Content .github/workflows/dast.yml
	 $start = 440; $end = 490
	 for ($i = $start; $i -le $end; $i++) { "{0}: {1}" -f $i, $lines[$i-1] }
	 ```

These short rules prevent the three common failures we saw: unsupported shell operators, quoting/expansion parser errors, and brittle file-slicing one-liners. When in doubt, prefer multi-line script files or small helper scripts invoked from PowerShell.
