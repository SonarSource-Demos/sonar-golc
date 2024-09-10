#!/bin/bash

# Run the Docker container
cd linux-arm64 && \
docker build -t sonar-golc-linux-arm64-test -f Dockerfile.test-linux-arm64 . && \
docker run --rm sonar-golc-linux-arm64-test

# Capture the exit code
EXIT_CODE=$?

# Check the exit code
if [ $EXIT_CODE -eq 0 ]; then
  echo "Test passed."
else
  echo "Test failed with exit code $EXIT_CODE."
fi

# Exit with the same code as the Docker container
exit $EXIT_CODE