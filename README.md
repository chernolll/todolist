# TodoList 应用

这是一个基于 **Golang** 🦦 和 **SQLite** 🗄️ 的简单待办事项管理应用，前端使用静态 HTML 文件 🌐，后端使用 Gin 框架 🍸 提供 RESTful API，并通过 Docker 🐳 和 Nginx 🚦 部署。

## ✨ 功能

1. **任务管理** 📝：
    - ➕ 添加任务
    - ✅ 切换任务完成状态
    - 🗑️ 删除任务
    - 🚮 清除已完成任务
2. **历史记录** 📜：
    - 👀 查看任务的操作历史（添加、完成、删除等）
3. **前后端分离** 🔗：
    - 🌍 前端通过 Nginx 提供静态文件服务
    - 🔌 后端提供 RESTful API
4. **数据持久化** 💾：
    - 使用 SQLite 存储任务和历史记录

---

## 🗂️ 项目结构

```
.
├── main.go             # 后端 Golang 源码
├── static/             # 前端静态 HTML、CSS、JS 文件
│   ├── index.html
│   └── ...
├── nginx.conf          # Nginx 配置文件
├── Dockerfile          # 后端 Docker 构建文件
├── docker-compose.yml  # 一键部署配置
├── README.md
└── ...
```

## 🚀 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/chernolll/todolist.git
cd todolist
```

### 2. 构建并启动服务

确保已安装 [Docker](https://www.docker.com/) 🐳 和 [Docker Compose](https://docs.docker.com/compose/) 🛠️。

```bash
docker-compose up --build
```

- 后端服务监听 `http://localhost:5175` ⚙️
- 前端页面通过 Nginx 访问 `http://localhost:5176` 🌐

## 🛠️ 常用命令

- 停止服务：`docker-compose down` ⏹️
- 查看日志：`docker-compose logs -f` 📋
- 重新构建：`docker-compose up --build` 🔄

## 📝 License

MIT License