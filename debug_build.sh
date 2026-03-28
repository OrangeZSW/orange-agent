#!/bin/bash
set -x
chmod +x ./build.sh
./build.sh 2>&1
echo "Exit code: $?"