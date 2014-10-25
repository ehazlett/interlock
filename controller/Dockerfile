FROM debian:jessie
RUN apt-get update
RUN apt-get install -y haproxy ca-certificates --no-install-recommends
ADD controller /usr/local/bin/controller
EXPOSE 8080 8443
ENTRYPOINT ["/usr/local/bin/controller"]
CMD []
