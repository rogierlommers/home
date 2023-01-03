FROM ubuntu
LABEL description="Quick-notes from Rogier Lommers"
LABEL maintainer="Rogier Lommers <rogier@lommers.org>"

# add binary and assets
COPY --chown=1000:1000 ./bin/quick-note/quick-note /app/quick-note/quick-note
COPY --chown=1000:1000 ./bin/dist /app/dist

# binary will serve on 8080
EXPOSE 8080

# make binary executable
RUN chmod +x /bin/quick-note

# set default dist directory
ENV DIST_DIRECTORY "/bin/dist"

# run binary
CMD ["/bin/quick-note"]
