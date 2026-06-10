# Bakes the freshly generated self-signed cert into the Gitea image. A bind
# mount of ./certs works locally but NOT under docker-in-docker (CI), where the
# mount would reference the daemon's filesystem, not the job's. The build
# context (this dir) is shipped to the daemon, so COPY sees the certs either way.
# Throwaway, localhost/CI only - never reuse this image or cert.
FROM gitea/gitea:1.22
COPY certs/ /certs/
# /certs must be traversable (0755) and the files readable by the unprivileged
# gitea user; COPY alone can leave the dir without its execute bit.
RUN chmod 0755 /certs && chmod 0644 /certs/server.crt /certs/server.key
