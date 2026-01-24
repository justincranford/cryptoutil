#!/bin/bash
# Fix gopls installation for VS Code

echo "=== Fixing gopls configuration ==="

# Check if gopls is already installed
if command -v gopls &> /dev/null; then
    echo "✅ gopls is already installed at: $(which gopls)"
    gopls version
else
    echo "❌ gopls not found - installing..."
    go install golang.org/x/tools/gopls@latest

    if [ $? -eq 0 ]; then
        echo "✅ gopls installed successfully at: $(which gopls)"
        gopls version
    else
        echo "❌ Failed to install gopls"
        exit 1
    fi
fi

echo ""
echo "=== VS Code Settings Check ==="
echo "If you still see errors, check your VS Code settings.json:"
echo "1. Remove or comment out 'alternateTools' setting"
echo "2. Or ensure gopls path is correct in alternateTools"
echo ""
echo "Example settings.json fix:"
echo '{'
echo '  "go.alternateTools": {'
echo '    "gopls": "'"$(which gopls)"'"'
echo '  }'
echo '}'
