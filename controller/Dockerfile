FROM debian:jessie
RUN apt-get update
RUN apt-get install -y haproxy
ADD controller /usr/local/bin/controller
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/controller"]
CMD []
