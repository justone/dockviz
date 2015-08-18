FROM scratch
COPY dockviz /
ENV IN_DOCKER true
ENTRYPOINT ["/dockviz"]
