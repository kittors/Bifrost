FROM node:24-alpine

WORKDIR /app

COPY docker/mock-service/server.mjs ./server.mjs

EXPOSE 8080

CMD ["node", "./server.mjs"]
