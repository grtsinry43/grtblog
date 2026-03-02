FROM node:22-alpine AS builder

WORKDIR /app

RUN corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --prod=false

COPY web/. .
COPY shared /shared

ARG APP_VERSION=dev
ARG BUILD_COMMIT=unknown
ENV APP_VERSION=${APP_VERSION} \
    BUILD_COMMIT=${BUILD_COMMIT}

RUN pnpm build

FROM node:22-alpine AS runtime

WORKDIR /app

ENV NODE_ENV=production

RUN corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --prod

COPY --from=builder /app/build /app/build
COPY deploy/docker/renderer-entrypoint.sh /usr/local/bin/renderer-entrypoint.sh
RUN chmod +x /usr/local/bin/renderer-entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/renderer-entrypoint.sh"]
