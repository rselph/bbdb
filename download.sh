#!/bin/bash

files=$(curl https://www.backblaze.com/cloud-storage/resources/hard-drive-test-data 2>/dev/null | egrep --only-matching 'https://[^\"]+data_[[:alnum:]_]+\.zip')
for file in ${files} ; do
    echo $(basename ${file})
    [[ -f $(basename ${file}) ]] || curl -O ${file}
done
