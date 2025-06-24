FROM golang:1.21-alpine AS builder

WORKDIR /redix/

RUN apk update && apk add git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /usr/bin/redix

FROM scratch

WORKDIR /redix/

COPY --from=builder /usr/bin/redix /usr/bin/redix

CMD ["/usr/bin/redix", "/etc/redix/redix.hcl"]