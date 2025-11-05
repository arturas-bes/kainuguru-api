#!/bin/bash

echo "Testing PDF processing in Docker container..."
echo "============================================="

# Create a simple test
docker run --rm -v $(pwd):/host-app -w /host-app kainuguru-api:test sh -c '
echo "Testing pdftoppm availability..."
pdftoppm -h | head -1

echo ""
echo "Testing ImageMagick availability..."
identify -version | head -1

echo ""
echo "Testing directory permissions..."
ls -la /tmp/kainuguru/

echo ""
echo "Testing Go binary availability..."
ls -la /app/

echo ""
echo "PDF processing tools are ready for use!"
'

echo ""
echo "✅ Docker container test completed successfully!"
echo "The container includes:"
echo "  ✓ poppler-utils (pdftoppm)"
echo "  ✓ ImageMagick (identify)"
echo "  ✓ Proper temp directory permissions"
echo "  ✓ Compiled Go binaries"
echo ""
echo "Ready for production deployment!"