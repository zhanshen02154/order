FROM alpine:3.23.2

ARG CONSUL_HOST=127.0.0.1
ARG CONSUL_PORT=8500
ARG CONSUL_PREFIX=/micro/

ENV CONSUL_HOST=$CONSUL_HOST \
	CONSUL_PORT=$CONSUL_PORT \
	CONSUL_PREFIX=$CONSUL_PREFIX

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
	&& apk add --no-cache tzdata \
	&& ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN addgroup -g 1002 app \
	&& adduser -S -D -u 1002 -G app app

COPY --chown=app:app --chmod=500 order /home/app/order

WORKDIR /home/app
USER app

CMD [ "./order" ]