FROM ubuntu
LABEL description="Quick-notes from Rogier Lommers"
LABEL maintainer="Rogier Lommers <rogier@lommers.org>"

# add binary and assets
COPY --chown=1000:1000 ./bin/quick-note/quick-note /app/quick-note
COPY --chown=1000:1000 ./bin/dist /app/dist

# binary will serve on 3000
EXPOSE 3000

# make binary executable
RUN chmod +x /app/quick-note

# set default dist directory
ENV DIST_DIRECTORY "/app/dist"

# run binary
CMD ["/app/quick-note"]
