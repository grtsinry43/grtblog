FROM node:22-alpine AS builder

WORKDIR /app

RUN corepack enable

COPY admin/package.json admin/pnpm-lock.yaml admin/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --prod=false

COPY admin/. ./
COPY shared /shared

ENV VITE_APP_BASE=/admin/
RUN pnpm build

FROM nginx:1.27-alpine AS runtime

COPY --from=builder /app/dist /usr/share/nginx/html/admin

RUN cat > /etc/nginx/conf.d/default.conf <<'NGINX'
server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;

    location = / {
        return 302 /admin/;
    }

    location /admin/ {
        try_files $uri $uri/ /admin/index.html;
    }
}
NGINX

EXPOSE 80
