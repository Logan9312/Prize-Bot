docker build -t discord-auction-bot .
docker run -e "ENV_VARS=$(<./scripts/env_vars.json)" -p 8080:8080 discord-auction-bot