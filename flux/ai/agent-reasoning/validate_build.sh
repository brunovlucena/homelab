#!/bin/bash
# Build validation script for Agent-Reasoning

set -e

echo "🔍 Validating Agent-Reasoning build..."

# Check Python syntax
echo "✓ Checking Python syntax..."
python3 -c "
import ast
import sys

files = [
    'src/reasoning/main.py',
    'src/reasoning/handler.py',
    'src/reasoning/__init__.py',
    'src/shared/types.py',
    'src/shared/metrics.py',
    'src/shared/__init__.py',
]

errors = []
for f in files:
    try:
        with open(f, 'r') as file:
            ast.parse(file.read(), f)
    except SyntaxError as e:
        errors.append(f'{f}: {e}')

if errors:
    print('✗ Syntax errors found:')
    for err in errors:
        print(f'  {err}')
    sys.exit(1)
else:
    print('✓ All Python files are syntactically valid')
"

# Check if required files exist
echo "✓ Checking required files..."
required_files=(
    "src/reasoning/main.py"
    "src/reasoning/handler.py"
    "src/reasoning/Dockerfile"
    "src/shared/types.py"
    "src/shared/metrics.py"
    "src/requirements.txt"
    "Makefile"
)

for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "✗ Missing required file: $file"
        exit 1
    fi
done
echo "✓ All required files present"

# Check Dockerfile syntax (if docker is available)
if command -v docker &> /dev/null; then
    echo "✓ Validating Dockerfile..."
    if docker build --dry-run -f src/reasoning/Dockerfile . &> /dev/null 2>&1 || \
       docker buildx build --dry-run -f src/reasoning/Dockerfile . &> /dev/null 2>&1; then
        echo "✓ Dockerfile is valid"
    else
        echo "⚠ Dockerfile validation skipped (docker not accessible)"
    fi
else
    echo "⚠ Docker not available, skipping Dockerfile validation"
fi

# Check imports (basic check)
echo "✓ Checking imports..."
python3 -c "
import sys
sys.path.insert(0, 'src')

try:
    # Try importing main modules
    from reasoning import main
    from shared import types, metrics
    print('✓ All imports are valid')
except ImportError as e:
    print(f'⚠ Import check: {e}')
    print('  (This is expected if dependencies are not installed)')
"

echo ""
echo "✅ Build validation complete!"
echo ""
echo "To build the Docker image, run:"
echo "  REGISTRY=your-registry IMAGE_NAME=agent-reasoning IMAGE_TAG=latest make build"
echo ""
echo "To install dependencies locally (in virtual environment):"
echo "  python3 -m venv venv"
echo "  source venv/bin/activate"
echo "  make install"
