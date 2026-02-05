@echo off
docker compose run --rm php php src/convert.php %*
