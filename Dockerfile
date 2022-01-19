FROM golang:1.17-alpine3.15
WORKDIR /fileBrowser
EXPOSE 10086
RUN echo "A file browser." \
    && apk add --no-cache git \
    && git clone https://github.com/Tan90degrees/fileBrowser.git . \
    && go build -ldflags "-w -s" -o fileBrowser.out
CMD [ "./fileBrowser.out" ]