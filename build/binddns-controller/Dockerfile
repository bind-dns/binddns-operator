FROM centos:7

ARG ARCH=amd64

COPY bin/${ARCH}/binddns-controller /app/
COPY webapp/ /app/webapp/
COPY rndc /etc/named/

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf

WORKDIR /app

CMD ["/app/binddns-controller"]
