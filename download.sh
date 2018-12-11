#!/bin/bash

files=$(curl https://www.backblaze.com/b2/hard-drive-test-data.html 2>/dev/null | egrep --only-matching https://[^\"]+data_[[:alnum:]_]+\\.zip)
for file in ${files} ; do
    echo $(basename ${file})
    [[ -f $(basename ${file}) ]] || wget ${file}
done
