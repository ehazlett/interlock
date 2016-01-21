FROM scratch
COPY cmd/interlock/interlock /bin/interlock
WORKDIR /bin
ENTRYPOINT ["/bin/interlock"]
EXPOSE 8080
CMD ["-h"]
