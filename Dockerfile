# Build Filebrowser
FROM node:16.20 AS filebrowser
WORKDIR /filebrowser/
COPY filebrowser-ui/package.json ./
RUN npm install
COPY filebrowser-ui/ ./
RUN npm run build

# Build Angular UI
FROM node:latest AS ui
WORKDIR /ui/
COPY ui/package.json ./
RUN npm install
COPY ui/ ./
COPY --from=filebrowser /filebrowser/dist/ ./src/assets/filebrowser_client/
RUN npm run build

# Build Backend
FROM golang:1.20 AS backend
WORKDIR /backend
COPY go.mod go.sum ./
RUN go mod download
COPY *.go config.yaml ./
COPY other/systemd/sfui.service ./other/systemd/sfui.service
COPY --from=ui /ui/dist/sf-ui/  ./ui/dist/sf-ui/
COPY other/docker/build_helper.go ./
COPY .git/refs/heads/main build_hash
RUN go run build_helper.go
RUN rm build_helper.go
RUN sh build.sh
RUN strip sfui


# Run SfUI
FROM alpine:latest
RUN apk add bash
WORKDIR /app/
COPY --from=backend /backend/sfui ./
COPY config.yaml ./
COPY other/ssh/known_hosts /root/.ssh/
EXPOSE 7171
ENTRYPOINT ["/app/sfui"]
