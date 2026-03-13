# 使用 golang 镜像作为运行环境，因为沙盒引擎需要调用 'go run' 命令
FROM golang:1.22-alpine

# 安装基础构建工具（如需要编译 C 插件等）
RUN apk add --no-cache build-base

WORKDIR /app

# 拷贝整个 monorepo 代码以处理本地包依赖
COPY . .

# 编译沙盒引擎二进制文件
RUN go build -o sandbox-main ./src/services/sandbox-engine/main.go

# 暴露端口
EXPOSE 8081

# 启动引擎
CMD ["./sandbox-main"]
