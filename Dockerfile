FROM node:20-alpine AS frontend-build

ARG NPM_CONFIG_REGISTRY=https://registry.npmmirror.com
ARG FRONTEND_SITE_URL=http://127.0.0.1:8090
ARG FRONTEND_API_BASE=

WORKDIR /app/web/frontend-app
COPY web/frontend-app/package*.json ./
RUN npm ci --registry="${NPM_CONFIG_REGISTRY}" --prefer-offline --no-audit --progress=false
COPY web/frontend-app ./
ENV FRONTEND_SITE_URL="${FRONTEND_SITE_URL}"
ENV FRONTEND_API_BASE="${FRONTEND_API_BASE}"
RUN npm run build

FROM node:20-alpine AS admin-build

ARG NPM_CONFIG_REGISTRY=https://registry.npmmirror.com

WORKDIR /app/web/admin-app
COPY web/admin-app/package*.json ./
RUN npm ci --registry="${NPM_CONFIG_REGISTRY}" --prefer-offline --no-audit --progress=false
COPY web/admin-app ./
RUN npm run build

FROM golang:1.22-alpine AS go-build

ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN GOPROXY="${GOPROXY}" GOSUMDB="${GOSUMDB}" go mod download
COPY . .
COPY --from=frontend-build /app/web/frontend /app/web/frontend
COPY --from=admin-build /app/web/admin-vue /app/web/admin-vue
RUN GOCACHE=/tmp/go-build go build -buildvcs=false -o /out/devhub .

FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -H -u 10001 devhub

WORKDIR /app
COPY --from=go-build /out/devhub /app/devhub
COPY --from=go-build /app/web/frontend /app/web/frontend
COPY --from=go-build /app/web/admin-vue /app/web/admin-vue
COPY db /app/db
RUN mkdir -p /app/storage /app/plugins-local \
  && chown -R devhub:devhub /app/storage /app/plugins-local

ENV PORT=8090
ENV CMS_STORE=mysql
ENV DB_HOST=mysql
ENV DB_PORT=3306
ENV DB_USER=devhub
ENV DB_NAME=devhub

EXPOSE 8090
USER devhub
CMD ["/app/devhub"]
