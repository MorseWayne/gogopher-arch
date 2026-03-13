# 第一阶段：编译前端
FROM node:20-alpine as builder

WORKDIR /app
COPY package*.json ./
RUN npm install --silent
COPY . .
RUN npm run build

# 第二阶段：运行 Nginx 托管
FROM nginx:stable-alpine
COPY --from=builder /app/dist /usr/share/nginx/html

# 这里我们可以配置 Nginx 反向代理，但为了 MVP 简单，我们直接在 Docker Compose 映射端口
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
