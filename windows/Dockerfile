FROM golang:1.24.5-nanoserver AS build-backend
LABEL builder=true

ENV CGO_ENABLED=0
ENV GO111MODULE=on

COPY ./src /app
RUN cd /app && go get -d \
 && go build \
    -a \
    -ldflags "-s -w -extldflags \"-static\"" \
    -o /bin/main main.go \
 && go build \
    -a \
    -ldflags "-s -w -extldflags \"-static\"" \
    -o /bin/cat cat.go

FROM mcr.microsoft.com/windows/nanoserver:ltsc2022

ENV LANG=C.UTF-8
ENV SERVER_PORT=8080

COPY ./src/gattaca.txt /gattaca.txt

EXPOSE $SERVER_PORT
ENTRYPOINT [ "/main", "-listen-addr=0.0.0.0:${SERVER_PORT}" ]
CMD [ ]

COPY --from=build-backend /bin/main /main
COPY --from=build-backend /bin/cat /cat
