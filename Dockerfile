# ==================== 后端构建阶段 ====================
FROM golang:1.25.0-alpine AS backend-builder

WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 配置Go模块代理为国内源
ENV GOPROXY=https://goproxy.cn,direct
ENV GONOPROXY=none
ENV GOSUMDB=sum.golang.org

# 下载依赖
RUN go mod download

# 复制所有代码（包含后端、前端、配置）
COPY . .

# 构建后端应用
RUN go build -o sitchi ./cmd/manage/

# ==================== 前端构建阶段 ====================
FROM node:22-alpine AS frontend-builder

WORKDIR /app

# 先复制 package.json 和 package-lock.json
COPY --from=backend-builder /app/webui/package.json /app/webui/package-lock.json ./webui/

WORKDIR /app/webui

# 安装依赖（只要 package.json 不变，就会使用缓存）
RUN npm install

# 复制前端源代码和配置（这些文件经常变动）
COPY --from=backend-builder /app/webui/ ./
COPY --from=backend-builder /app/configs ../configs/

# 构建前端（生产模式）
RUN npm run build

# ==================== 运行阶段 ====================
FROM alpine:latest

# 安装必要工具
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 复制后端可执行文件
COPY --from=backend-builder /app/sitchi .

# 复制配置文件目录
COPY --from=backend-builder /app/configs ./configs

# 复制前端构建产物
COPY --from=frontend-builder /app/webui/dist ./webui/dist

# 暴露端口
EXPOSE 5050

# 设置启动命令
CMD ["./sitchi", "runServer"]
