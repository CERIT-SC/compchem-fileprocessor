#!/bin/bash


INPUT_DIR="/input"
OUTPUT_DIR="/output"

mkdir -p "$OUTPUT_DIR"

for input_file in "$INPUT_DIR"/*; do
  echo "input file $input_file"

  # 1) Create a job
  # A file in tpr format is required (in this example the file is named run.tpr)
  filename=$(basename "$input_file")
  filename="${filename%.*}"
  echo "filename $filename"
  resp=$(curl -s -X POST https://gmd.ceitec.cz/api/annotate -F "tpr=@$input_file" -F "keep=true")
  uuid=$(echo "$resp" | jq -r .uuid)

  # 2) Wait until the job is completed
  while true; do
    status=$(curl -s https://gmd.ceitec.cz/api/annotate/"$uuid" | jq -r .status)
    echo "Status: $status"
    [[ $status == "completed" ]] && break
    sleep 5
  done

  # 3) Download the results
  curl -s https://gmd.ceitec.cz/api/annotate/"$uuid"/results -o "$OUTPUT_DIR/$filename-simulation-annotation.json"
  echo "output saved in: $OUTPUT_DIR/$filename-simulation-annotation.json"

done
