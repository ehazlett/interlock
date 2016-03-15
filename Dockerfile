FROM alpine:latest
RUN apk add -U iptables
COPY cmd/interlock/interlock /bin/interlock
WORKDIR /bin
ENTRYPOINT ["/bin/interlock"]
EXPOSE 8080
LABEL interlock.ext
CMD ["-h"]
