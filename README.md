嗨！这是 AITaskSystem (你的 AI 任务管家)

欢迎！这是一个挺酷的任务管理后端项目，用的是 Go + Gin + GORM + SQLite 这一套“黄金搭档”。
最大的亮点是啥？它支持用大白话加任务（我们内置了一个模拟的 AI 引擎），你说人话，它就能听懂！

🛠 用到的家伙什

语言: Go 1.21+ (得有这个！)

框架: Gin Web Framework (跑得飞快)

数据库: SQLite (本地存数据，省心又省事)

架构: 分层架构 (Model -> Repo -> Service -> Handler，条理清晰！)

🚀 跑起来！

1. 启动服务

先把 Go 环境装好哈，然后只需三步：

# 1. 进到项目目录里去 (一定要在有 go.mod 的那一层哦)
# 2. 让 Go 整理一下依赖
go mod tidy

# 3. 启动！
go run main.go


当你看到屏幕上蹦出 Server starting on http://localhost:8080，恭喜，服务已经跑起来啦！

2. 来玩玩看 (接口测试)

别关刚才那个窗口，新开一个终端窗口来发指令试试。

A. 试试 AI 加任务 (核心玩法)

直接把这句话扔给它，看它怎么解析：

curl -X POST http://localhost:8080/api/v1/tasks/ai \
-H "Content-Type: application/json" \
-d '{"prompt": "明天需要紧急开发支付接口代码"}'


你会看到的惊喜 (终端返回的数据):

{
"message": "AI Task Created",
"task": {
"ID": 1,
"title": "明天需要紧急开发支付接口代码",
"status": "Pending",
"priority": "High",          // 看到没？捕捉到 "紧急" 啦！
"estimated_hours": 4,        // 提到 "开发"，自动预估 4 小时
"due_date": "202x-xx-xx..."  // "明天" 这里的日期自动算好了
}
}


B. 看看任务列表

查查你都记了啥：

curl http://localhost:8080/api/v1/tasks


C. 生成自动周报

干完活了？先要把任务状态改成 Completed (已完成)，然后来看看总结报告：

curl http://localhost:8080/api/v1/report


✅ 测测它的脑子 (自动化测试)

跑个单元测试，看看我们的 AI 逻辑是不是一直在线：

go test ./service/... -v


📂 目录大概是这样

model/: 长啥样的数据都在这

repository/: 专门负责跟数据库打交道

service/: 这里是“大脑”，业务逻辑和 AI 解析都在这

handler/: 接待员，负责收发 HTTP 请求