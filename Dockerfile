FROM scratch
COPY cmd/interlock/interlock /bin/interlock
WORKDIR /bin
ENTRYPOINT ["/bin/interlock"]
CMD ["-h"]
