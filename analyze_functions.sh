#!/bin/bash

echo "=== Function Size Analysis ==="
echo "Finding functions exceeding 50 lines..."
echo

find . -name "*.go" -not -path "./test*" -not -path "./.trunk/*" -not -path "./vendor/*" -not -path "./.git/*" | while read file; do
    if [ -f "$file" ]; then
        awk '
        /^func / {
            func_name = $0
            func_start = NR
            in_func = 1
            brace_count = 0
        }
        in_func {
            # Count braces
            for (i = 1; i <= length($0); i++) {
                char = substr($0, i, 1)
                if (char == "{") brace_count++
                if (char == "}") brace_count--
            }

            # Function ends when braces balance
            if (brace_count == 0 && in_func && NR > func_start) {
                func_lines = NR - func_start + 1
                if (func_lines > 50) {
                    print FILENAME ":" func_start ":" func_name " (" func_lines " lines)"
                }
                in_func = 0
            }
        }
        ' "$file"
    fi
done | sort
