FROM alpine:3.2
ADD html /html
ADD message-web /message-web
WORKDIR /
ENTRYPOINT [ "/message-web" ]
