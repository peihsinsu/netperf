#!/bin/bash
# Make all shell scripts executable

chmod +x build-release.sh
chmod +x verify-build.sh
chmod +x test-summary.sh
chmod +x test-build.sh
chmod +x setup-scripts.sh

echo "✓ All scripts are now executable"
ls -lh *.sh
