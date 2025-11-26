# File Hub - 个人文件同步服务

[![构建状态](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/cgang/file-hub)
[![许可证](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)

> ⚠️ **开发中** - 此项目正在积极开发中，尚未准备好用于生产环境。

一个支持 WebDAV 的个人文件备份和同步服务，使用 PostgreSQL 存储元数据，采用高效的二进制差异同步。

[English](README.md) | 中文

## 📌 项目目标
- 跨设备实时文件同步
- 文件管理的网页界面
- 跨平台兼容性（Windows、macOS、Linux、Android、iOS）
- 简单的家庭/个人 NAS 系统部署
- 客户端代码维护在单独的仓库中

## 🔍 核心功能
### 📁 存储架构
- 使用 PostgreSQL 元数据的原生文件系统存储
- 简单的目录结构配置
- 基本配额管理

### 🔄 同步机制
- 用于带宽优化的二进制差异算法
- 冲突解决策略
- WebDAV 协议实现

### 🔐 安全性
- 数据传输和静态数据的端到端加密
- 数据库存储的认证凭证

### ⚡ 性能
- 差分编码传输
- 频繁文件的缓存
- 并行同步操作

## 🛠️ 配置
服务默认在 `config/config.yaml` 查找配置文件。您可以通过设置 `CONFIG_PATH` 环境变量来覆盖此设置，该变量遵循 PATH 约定（用冒号分隔的目录）。服务将按顺序在每个目录中搜索 `config.yaml` 直到找到为止。如果找不到配置文件，将使用默认值：
- 存储目录：`root`
- 端口：`8080`

示例配置在 `example/config.yaml` 中提供：
```yaml
storage:
  root_dir: "root"  # 相对于 example 目录
webdav:
  port: "8080"
```

要自定义服务，请设置包含 config.yaml 的 CONFIG_PATH 环境变量：
```bash
CONFIG_PATH=/path/to/config/directory ./bin/file-hub
```

或者使用多个目录（第一个匹配项获胜）：
```bash
CONFIG_PATH=/first/config/dir:/second/config/dir ./bin/file-hub
```

或者复制示例配置并根据需要修改：
```bash
cp example/config.yaml config/config.yaml
# 然后编辑 config/config.yaml 以满足您的需求
```

## 🚀 快速开始
```bash
# 克隆仓库
git clone https://github.com/cgang/file-hub.git
cd file-hub

# 设置数据库
createdb filehub
psql -d filehub -f scripts/database_schema.sql

# 使用示例配置构建并运行集成服务（存储在 example/root 中）
make run
```

要自定义服务，请设置 CONFIG_PATH 环境变量：
```bash
CONFIG_PATH=/path/to/config/directory ./bin/file-hub
```

## 🌐 集成网页界面
该项目具有一个使用 Svelte 构建的现代网页界面，直接嵌入在二进制文件中。单个二进制文件同时提供 WebDAV API 和网页界面：

- **网页界面**：可通过 `http://localhost:8080` 访问（或您配置的主机/端口）
- **WebDAV API**：可在 `http://localhost:8080/dav` 使用

开发用法：
```bash
# 安装网页界面依赖
make web-install

# 在开发模式下运行网页界面（与后端分离）
make web-dev

# 将前端资源构建到二进制文件中
make build
```

## 💡 贡献
有兴趣贡献吗？请参阅我们的 [CONTRIBUTING.md](docs/CONTRIBUTING.md) 指南。

## 🤖 AI 助手
在开发过程中，此项目得到了 Qwen AI 助手的帮助。AI 协助了代码生成、文档编写、重构和错误修复等工作。