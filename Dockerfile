FROM golang:1.16 as builder

WORKDIR /src

COPY . /src

RUN go build -o /bin/listmailer -v .

FROM golang:1.16 as app

WORKDIR /src

COPY --from=builder /bin/listmailer /bin/listmailer

ENTRYPOINT [ "/bin/listmailer" ]