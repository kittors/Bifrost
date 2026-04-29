FROM node:24.11.1-alpine AS builder

WORKDIR /workspace

RUN corepack enable && corepack prepare pnpm@10.33.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY scripts/package-manager ./scripts/package-manager
COPY apps/admin ./apps/admin
COPY packages ./packages

RUN pnpm install --frozen-lockfile
RUN pnpm --filter @bifrost/admin build

FROM nginx:1.29-alpine

COPY docker/admin-web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /workspace/apps/admin/dist /usr/share/nginx/html

EXPOSE 5173
