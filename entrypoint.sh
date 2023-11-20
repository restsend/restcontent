#!/bin/sh
ADDR=0.0.0.0:${PORT:=8080}
# output with logfile 
if [ -n "$LOGFILE" ]; then
    LOG="-log $LOGFILE"
fi

set -x 
./restcontent -addr $ADDR $LOG $@