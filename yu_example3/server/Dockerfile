FROM debian:jessie

LABEL version=1.0.0
#ADD ./* /usr/local/bin/mpServer
#COPY mpServer /usr/local/bin/mpServer
WORKDIR /usr/local/bin/mpServer
COPY  .  /usr/local/bin/mpServer
ENTRYPOINT ["/usr/local/bin/mpServer/mpServer"]
