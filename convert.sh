#!/bin/bash
TZ=$(readlink /etc/localtime 2>/dev/null | sed 's|.*/zoneinfo/||') docker compose run --rm php php src/convert.php "$@"
