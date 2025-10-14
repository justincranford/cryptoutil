@echo off
REM Pre-commit setup script for Windows
REM Ensures consistent cache location and proper installation

echo Setting up pre-commit hooks...

REM Set consistent cache location
setx PRE_COMMIT_HOME "C:\Users\%USERNAME%\.cache\pre-commit" /M

REM Install pre-commit if not already installed (use python -m pip for PATH compatibility)
python -m pip install pre-commit

REM Install the hooks (use python -m pre_commit for PATH compatibility)
python -m pre_commit install

REM Test the setup
python -m pre_commit run --all-files

echo Pre-commit setup complete!
echo Cache location: C:\Users\%USERNAME%\.cache\pre-commit
