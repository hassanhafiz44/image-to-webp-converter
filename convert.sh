#!/bin/bash
docker compose run --rm php php src/convert.php "$@"
