export ENVIRONMENT=$(echo ${ENV_VARS} | jq -r '.ENVIRONMENT')
export DISCORD_TOKEN=$(echo ${ENV_VARS} | jq -r '.DISCORD_TOKEN')
export MIGRATE=$(echo ${ENV_VARS} | jq -r '.MIGRATE')

/main