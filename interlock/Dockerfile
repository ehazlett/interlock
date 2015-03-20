FROM debian:jessie
RUN apt-get update
RUN apt-get install -y haproxy ca-certificates --no-install-recommends
ADD interlock /usr/local/bin/interlock
EXPOSE 8080 8443
ENTRYPOINT ["/usr/local/bin/interlock"]
