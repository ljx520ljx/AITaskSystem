# 🤖 AI TaskPilot - 智能任务决策引擎

欢迎使用 **AI TaskPilot**！这是一个基于 **DeepSeek R1** 大语言模型构建的智能任务决策系统，而非普通的待办事项列表。

它能听懂你的自然语言指令，自动分析任务的优先级、截止时间、工时预测。

---

## ✨ 核心亮点

### 🧠 全流程 AI 托管
* **自然语言交互**：只需输入 "下周五前必须上线新官网，需要先购买服务器"，AI 会自动拆解。
* **依赖反向推导**：当创建高优任务时，AI 会自动扫描旧任务，发现前置依赖并自动提升其优先级（*注：AI 判断依赖关系仅供参考，支持人工修正*）。
* **智能工时预测**：根据任务复杂度自动估算所需时间。

### 📊 智能周报生成
* **一键生成**：自动生成 Markdown 格式的周报，智能总结工作亮点与产出。

---

## 🛠️ 环境准备

在开始之前，请确保你的环境已安装：

* **Go (Golang)**: 版本 `1.21` 或更高
* **Git**: 用于代码版本控制

---

## 🚀 快速启动指南

### 第一步：克隆项目

```bash
git clone https://github.com/ljx520ljx/AITaskSystem
cd AITaskSystem
```

### 第二步：配置 API Key 🔑

本项目依赖 **DeepSeek** (或兼容 OpenAI 格式的模型) 提供智能服务。
请在启动前，根据你的操作系统设置环境变量：

**🐧 Linux / 🍎 macOS (终端):**
```bash
export DEEPSEEK_API_KEY="你的真实DeepSeek密钥"
```

**🪟 Windows (PowerShell):**
```powershell
$env:DEEPSEEK_API_KEY="你的真实DeepSeek密钥"
```

**🪟 Windows (CMD):**
```cmd
set DEEPSEEK_API_KEY=你的真实DeepSeek密钥
```

> **提示**：如果你还没有 Key，可以去 DeepSeek 开放平台申请。如果没有设置此环境变量，AI 相关功能将无法正常工作。

### 第三步：下载依赖并运行

在项目根目录下执行：

```bash
# 1. 整理并下载 Go 依赖包
go mod tidy

# 2. 启动后端服务
go run main.go
```

当看到终端输出以下信息时，代表启动成功：

```text
[GIN-debug] Listening and serving HTTP on :8080
2025/xx/xx xx:xx:xx Server starting on http://localhost:8080
```

---

## 🖥️ 如何使用

### 1. 进入前端界面
后端启动后，打开浏览器（推荐 Chrome 或 Edge），访问：

👉 **[http://localhost:8080](http://localhost:8080)**

你将看到 AI TaskPilot 界面。

### 2. 功能操作指南

#### 📝 创建任务
在左侧深色区域的输入框中，输入任何指令。

> **示例 1**： "紧急修复支付接口 Bug，预计耗时 2 小时"
>
> **示例 2**： "下周三前完成数据库迁移，这依赖于购买新服务器"

按 `Ctrl + Enter` 或点击“**执行指令**”，AI 会自动分析并创建任务。

#### 👁️ 查看与管理
* 右侧看板会根据 **依赖关系 > 截止时间 > 优先级** 自动排序。
* 点击任务左侧的方框 ✅ 可以完成/取消完成。
* 点击任务标题 ✏️ 可以手动修正 AI 的判断。

#### 📑 生成周报
* 点击左下角的 "**生成周报总结**" 按钮。
* 稍等片刻，DeepSeek 会为你撰写一份排版精美的 Markdown 工作总结。

---

## 🏗️ 技术栈

* **Backend**: Go (Golang), Gin Web Framework, GORM (SQLite)
* **Frontend**: Vue.js 3, Tailwind CSS
* **AI Engine**: DeepSeek R1