et -e

if [ "$1" = "tugbot" ]; then
  if [ -S /var/run/docker.sock ]; then
    chown -R tugbot:tugbot /var/run/docker.sock
  fi
  exec gosu tugbot:tugbot "$@"
fi

exec "$@"
