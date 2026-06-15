#!/bin/sh
if [ "$(cat /proc/sys/crypto/fips_enabled 2>/dev/null)" = "1" ]; then
  export GODEBUG="${GODEBUG:+${GODEBUG},}fips140=on"
else
  export GODEBUG="${GODEBUG:+${GODEBUG},}fips140=off"
fi
exec "$@"
