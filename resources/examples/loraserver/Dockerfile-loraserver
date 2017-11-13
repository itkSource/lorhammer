FROM debian:9.2

# get version list : https://docs.loraserver.io/loraserver/overview/downloads/
ARG VERSION=0.22.0

RUN echo 'Acquire::http::Pipeline-Depth "0";' > /etc/apt/apt.conf.d/http-pipeline && \
    apt update && \
    apt install -y procps wget

RUN wget https://dl.loraserver.io/tar/loraserver_${VERSION}_linux_amd64.tar.gz -O /tmp/loraserver.tgz && \
    tar xvf /tmp/loraserver.tgz -C /usr/local/bin && \
    chmod 755 /usr/local/bin/loraserver && \
    rm /tmp/loraserver.tgz

ADD ./rsa/issued/ /etc/ssl/certs/
ADD ./rsa/private/ /etc/ssl/private/

EXPOSE 8000

HEALTHCHECK CMD bash -c '[[ $(ps faux | grep loraserver | grep -v grep | wc -l) > 0 ]]' && exit 0 || exit 1

CMD ["loraserver"]
