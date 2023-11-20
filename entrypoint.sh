#!/bin/sh
ADDR=0.0.0.0:${PORT:=8080}

if [ -z "$DSN" ]; then
    DSN="file:data/restcontent.db"
fi

# output with logfile 
if [ -n "$LOGFILE" ]; then
    LOG="-log $LOGFILE"
fi

set -x 
SESSION_SECRET=$SESSION_SECRET ./restcontent -addr $ADDR -dsn $DSN $LOG $@