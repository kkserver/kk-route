FROM alpine:latest

RUN echo "Asia/shanghai" >> /etc/timezone

COPY ./main /bin/kk-route

RUN chmod +x /bin/kk-route

RUN mkdir /lib

RUN mkdir /lib/lua

COPY ./config /config

COPY ./app.ini /app.ini

ENV KK_ENV_CONFIG /config/env.ini

VOLUME /config

VOLUME /lib/lua

CMD kk-route $KK_ENV_CONFIG

