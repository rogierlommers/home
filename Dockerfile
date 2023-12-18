FROM ubuntu
LABEL description="Home server from Rogier Lommers"
LABEL maintainer="Rogier Lommers <rogier@lommers.org>"

# add binary and assets
COPY --chown=1000:1000 ./bin/home /app/home

# isntall CA certificates (needed for smtp sending)
RUN apt-get update && apt-get install ca-certificates -y && update-ca-certificates

# binary will serve on 3000
EXPOSE 3000

# make binary executable
RUN chmod +x /app/home

# run binary
CMD ["/app/home"]
