#!/bin/bash
export TZ=$(readlink /etc/localtime 2>/dev/null | sed 's|.*/zoneinfo/||')
docker compose up
