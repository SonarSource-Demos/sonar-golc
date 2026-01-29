#!/bin/sh
set -e

# Copy static assets into data dir so ResultsAll can serve them (ResultsAll expects dist/ and imgs/ under CWD)
if [ -d /app/dist ] && [ ! -d /data/dist ]; then
	cp -r /app/dist /data/
fi
if [ -d /app/imgs ] && [ ! -d /data/imgs ]; then
	cp -r /app/imgs /data/
fi

# Run analysis (config from /config via GOLC_CONFIG_FILE)
/app/golc -devops "${GOLC_DEVOPS:-Github}"

# Serve results on PORT (default 8092)
exec /app/ResultsAll
