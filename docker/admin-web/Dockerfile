FROM node:24-alpine

WORKDIR /app

COPY docker/admin-web/server.mjs ./server.mjs
COPY docker/admin-web/index.html ./index.html

EXPOSE 5173

CMD ["node", "./server.mjs"]
