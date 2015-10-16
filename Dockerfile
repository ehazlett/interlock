FROM debian:jessie
RUN echo deb http://httpredir.debian.org/debian jessie-backports main | \
      tee /etc/apt/sources.list.d/backports.list
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get install -y debian-keyring
RUN gpg --keyring /usr/share/keyrings/debian-keyring.gpg \
      --export bernat@debian.org | \
         apt-key add -
RUN echo deb http://haproxy.debian.net jessie-backports-1.6 main | \
      tee /etc/apt/sources.list.d/haproxy.list
RUN apt-get update
RUN apt-get install -y --no-install-recommends haproxy -t jessie-backports-1.6
RUN apt-get install -y --no-install-recommends nginx-full ca-certificates
COPY interlock/interlock /usr/local/bin/interlock
EXPOSE 80 443
ENTRYPOINT ["/usr/local/bin/interlock"]
