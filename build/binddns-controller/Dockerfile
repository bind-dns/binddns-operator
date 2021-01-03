FROM alpine:3.12

ARG ARCH=amd64

COPY bin/${ARCH}/binddns-operator /app/

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf

WORKDIR /app

CMD ["/app/binddns-operator"]
