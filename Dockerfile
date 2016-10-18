FROM alpine:latest

RUN echo "Asia/shanghai" >> /etc/timezone

COPY ./main /bin/kk-route

RUN chmod +x /bin/kk-route

ENV KK_NAME kk.

ENV KK_ARGS --remote

EXPOSE 87

CMD kk-route $KK_NAME --local 0.0.0.0:87 $KK_ARGS
