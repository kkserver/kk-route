FROM alpine:latest

COPY ./kk-route /bin/kk-route

ENV KK_NAME kk.

ENV KK_ARGS --remote

EXPOSE 87

CMD kk-go $KK_NAME --local 0.0.0.0:87 $KK_ARGS
