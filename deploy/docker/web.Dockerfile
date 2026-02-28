FROM node:22-alpine AS builder

WORKDIR /app

RUN corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --prod=false

COPY web/. .
COPY shared /shared
RUN pnpm build

FROM node:22-alpine AS runtime

WORKDIR /app

ENV NODE_ENV=production

RUN corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --prod

COPY --from=builder /app/build /app/build

EXPOSE 3000

CMD ["node", "build/index.js"]
