#!/bin/bash

# Script to fix all domain.New*Variable calls that now return (Variable, error)

# Find all files with the pattern and fix them
for file in test/*.go; do
    echo "Processing $file..."
    
    # Create a temporary file
    temp_file="${file}.tmp"
    
    # Process the file to fix domain.New*Variable calls
    sed -E 's/domain\.New([A-Za-z]+)Variable\(([^)]+)\)/var_\2, _ := domain.New\1Variable(\2)/g' "$file" > "$temp_file"
    
    # Check if the file was modified
    if ! cmp -s "$file" "$temp_file"; then
        echo "Fixed $file"
        mv "$temp_file" "$file"
    else
        rm "$temp_file"
    fi
done

echo "Done!"
