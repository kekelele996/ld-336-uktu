FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json .npmrc ./
RUN npm config set registry https://registry.npmmirror.com && npm install --no-audit --no-fund
COPY . .
RUN npm run build -- --configuration production

FROM nginx:1.25-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist/medasset-frontend/browser /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
