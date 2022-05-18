#!/bin/bash
# Recompiles index.ts whenever it changes.
while [ 1 ]; do 
    if [ index.ts -nt .touch ]; then 
        echo "New change"
        touch .touch
        tsc index.ts --outFile index.js --target es2017
    fi
    sleep 1
done