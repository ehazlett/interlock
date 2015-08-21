FROM debian:jessie
RUN apt-get update
RUN apt-get install -y haproxy nginx-full ca-certificates --no-install-recommends
ADD interlock /usr/local/bin/interlock
EXPOSE 80 443
ENTRYPOINT ["/usr/local/bin/interlock"]
