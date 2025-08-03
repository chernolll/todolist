# 第一阶段：构建Golang应用
FROM golang:1.24.5-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY *.go ./

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# 第二阶段：创建最终镜像
FROM alpine:latest

# 安装ca证书和SQLite运行时
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# 从构建阶段复制应用
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]