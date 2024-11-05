#!/bin/bash

# Check if a parameter is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 {dev|prod}"
    exit 1
fi

# Get the parameter
BUILD_TAG=$1

# Run the appropriate command based on the parameter
case $BUILD_TAG in
    dev)
        go run -tags dev .
        ;;
    prod)
        go run -tags prod .
        ;;
    *)
        echo "Invalid parameter: $BUILD_TAG"
        echo "Usage: $0 {dev|prod}"
        exit 1
        ;;
esac
