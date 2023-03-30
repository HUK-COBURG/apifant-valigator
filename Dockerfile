FROM golang:1.19-alpine AS golang

WORKDIR /tmp/validator
ADD cmd ./cmd
ADD go.mod .
ADD go.sum .
WORKDIR /tmp/validator/cmd
RUN GOARCH=arm64 go build -v -ldflags '-s -w' -o "out/validator" .

FROM node:16-alpine

WORKDIR /usr/src/spectral

RUN apk add bash jq curl unzip tree

RUN npm install -g @stoplight/spectral@5.9.2

ENV NODE_ENV production

ADD docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
ENTRYPOINT [ "/docker-entrypoint.sh" ]

COPY --from=golang /tmp/validator/cmd/out/validator /usr/local/bin/validator
RUN chmod +x /usr/local/bin/validator

CMD ["validator"]