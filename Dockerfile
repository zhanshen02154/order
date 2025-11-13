FROM alpine:3.22.2

ARG CONSUL_HOST=127.0.0.1
ARG CONSUL_PORT=8500
ARG CONSUL_PREFIX=/micro/

ENV CONSUL_HOST=$CONSUL_HOST \
	CONSUL_PORT=$CONSUL_PORT \
	CONSUL_PREFIX=$CONSUL_PREFIX

COPY order /var/order

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && chmod +x /var/order

WORKDIR /var

CMD [ "./order" ]