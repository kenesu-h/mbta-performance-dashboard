FROM node:16.20.1

WORKDIR /app
COPY . .

EXPOSE 80

RUN npm ci
RUN npm run build

CMD ["npx", "serve", "-s", "dist", "-l", "80"]
