# 构建应用
FROM golang:1.25.0-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 配置Go模块代理为国内源
ENV GOPROXY=https://goproxy.cn,direct
ENV GONOPROXY=none
ENV GOSUMDB=sum.golang.org

# 下载依赖
RUN go mod download

# 复制所有代码
COPY . .

# 构建应用
RUN go build -o sitchi ./cmd/manage/

# 运行应用
FROM alpine:latest

# 安装 ca-certificates 以支持 HTTPS
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /app

# 复制构建好的可执行文件
COPY --from=builder /app/sitchi .
# 复制配置文件目录
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 5050

# 设置启动命令
CMD ["./sitchi", "runServer"]
