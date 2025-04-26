#!/bin/bash

INPUT_DIR="/input"
OUTPUT_DIR="/output"

mkdir -p "$OUTPUT_DIR"

for input_file in "$INPUT_DIR"/*; do
    filename=$(basename "$input_file")
    output_file="$OUTPUT_DIR/count-$filename"

    > "$output_file"

    tr ' ' '\n' < "$input_file" | \
    sort | uniq -c | \
    awk '{print $2 ": " $1}' > "$output_file"
done
