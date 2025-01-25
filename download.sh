#!/bin/bash

files=$(curl https://www.backblaze.com/cloud-storage/resources/hard-drive-test-data 2>/dev/null | \
    grep -E --only-matching 'https://[^\"]+data_[[:alnum:]_]+\.zip')
for file in ${files} ; do
    basename "${file}"
    [[ -f $(basename "${file}") ]] || curl -O "${file}"
done
