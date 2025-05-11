FROM debian:stable-slim
COPY server /bin/server
CMD ["/bin/server"]