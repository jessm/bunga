# Build react app
FROM node:lts-alpine3.15 as nodebuild
WORKDIR /app
ENV PATH /app/node_modules/.bin:$PATH
COPY package.json ./
COPY package-lock.json ./
RUN npm install
COPY javascript ./javascript
COPY styles ./styles
RUN npm run build:prod

# Build golang binary
FROM golang:1.18-buster AS gobuild
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY pkg/*.go ./
RUN go build -o /bungaServer

# Deploy
FROM gcr.io/distroless/base-debian10
WORKDIR /app
COPY --from=nodebuild /app/assets ./assets
COPY assets/cards ./assets/cards
COPY templates ./templates
COPY --from=gobuild /bungaServer ./bungaServer

ENTRYPOINT ["/app/bungaServer"]
