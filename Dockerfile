FROM golang:1.19-alpine AS golang

WORKDIR /tmp/valigator
ADD cmd ./cmd
ADD go.mod .
ADD go.sum .
WORKDIR /tmp/valigator/cmd
RUN GOARCH=arm64 go build -v -ldflags '-s -w' -o "out/valigator" .

FROM node:16-alpine

WORKDIR /usr/src/valigator

RUN apk add bash jq curl unzip tree

RUN npm install -g @stoplight/spectral@5.9.2

ENV NODE_ENV production

ADD docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
ENTRYPOINT [ "/docker-entrypoint.sh" ]

COPY --from=golang /tmp/valigator/cmd/out/valigator /usr/local/bin/valigator
RUN chmod +x /usr/local/bin/valigator

COPY valigator.json /usr/src/valigator/valigator.json

CMD ["valigator"]