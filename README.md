# QuotePro - 报价回复助手

> 把客户一句需求，自动变成：报价单 + 回复话术 + 参数确认清单 + 附件包。

面向外贸公司 / 定制工厂销售的 AI 报价工具。用户打开系统 → 粘贴客户需求 → 5 分钟内生成并发出报价。

---

## 项目结构

```
LamMap/
│
├── README.md                          # 项目说明文档
├── doc/
│   ├── 任务.md                         # 产品需求文档（PRD）
│   └── 开始任务.md                      # 开发任务清单（带勾选追踪）
│
├── frontend/                           # 前端 - Vue 3
│   ├── index.html                      # 入口 HTML
│   ├── package.json                    # 依赖管理
│   ├── vite.config.ts                  # Vite 构建配置（含 API 代理）
│   ├── tsconfig.json                   # TypeScript 配置
│   └── src/
│       ├── main.ts                     # 应用入口（挂载 Vue/Pinia/Router/ElementPlus）
│       ├── App.vue                     # 根组件
│       ├── style.css                   # 全局样式（Tailwind CSS）
│       ├── vite-env.d.ts               # 类型声明
│       │
│       ├── router/
│       │   └── index.ts                # 路由配置（含登录守卫）
│       │
│       ├── stores/                     # Pinia 状态管理
│       │   ├── user.ts                 # 用户状态（登录/注册/登出）
│       │   └── quote.ts               # 报价状态（解析/生成/保存）
│       │
│       ├── utils/
│       │   └── request.ts              # Axios 封装（拦截器/Token/错误处理）
│       │
│       ├── layouts/
│       │   └── MainLayout.vue          # 主布局（侧边栏 + 顶栏 + 内容区）
│       │
│       ├── views/                      # 页面组件
│       │   ├── Login.vue               # 登录/注册页
│       │   ├── Dashboard.vue           # 工作台首页
│       │   ├── Settings.vue            # 系统设置
│       │   ├── quotes/
│       │   │   ├── NewQuote.vue        # ⭐ 核心：新建报价（三栏布局）
│       │   │   ├── QuoteDetail.vue     # 报价详情/编辑
│       │   │   └── QuoteHistory.vue    # 历史报价列表
│       │   ├── products/
│       │   │   ├── ProductList.vue     # 产品资料库列表
│       │   │   └── ProductDetail.vue   # 产品详情/编辑
│       │   └── templates/
│       │       └── TemplateList.vue    # 模板管理
│       │
│       └── components/quote/           # 报价相关子组件
│           ├── ParamField.vue          # 可编辑参数字段
│           ├── QuotationPreview.vue    # 报价单预览（行内编辑）
│           ├── ReplyVersions.vue       # 回复话术（3版本+语气切换）
│           ├── ConfirmationChecklist.vue # 参数确认清单
│           └── AttachmentPack.vue      # 附件包管理
│
└── backend/                            # 后端 - Go
    ├── main.go                         # 服务入口（路由注册）
    ├── go.mod / go.sum                 # Go 模块依赖
    │
    ├── config/
    │   └── config.go                   # 配置管理（环境变量）
    │
    ├── models/                         # 数据库模型（GORM）
    │   ├── db.go                       # 数据库连接 + 自动迁移
    │   ├── user.go                     # 用户表（含密码加密）
    │   ├── product.go                  # 产品表
    │   ├── quote.go                    # 报价表 + 报价明细表
    │   └── template.go                 # 模板表
    │
    ├── middleware/
    │   ├── cors.go                     # CORS 跨域中间件
    │   └── auth.go                     # JWT 认证中间件
    │
    ├── handlers/                       # API 处理器
    │   ├── response.go                 # 统一响应格式
    │   ├── auth.go                     # 注册/登录/获取用户信息
    │   ├── product.go                  # 产品 CRUD + 导入
    │   ├── quote.go                    # 报价 CRUD + AI 解析/生成 + 复制
    │   ├── template.go                 # 模板 CRUD
    │   ├── dashboard.go                # 工作台统计 + 最近报价
    │   └── upload.go                   # 文件上传
    │
    ├── services/
    │   └── ai.go                       # OpenAI API 集成（需求解析 + 报价生成）
    │
    └── uploads/                        # 文件上传存储目录
```

---

## 技术栈


| 层级          | 技术                 | 说明                        |
| ----------- | ------------------ | ------------------------- |
| **前端框架**    | Vue 3 + TypeScript | 组合式 API (Composition API) |
| **构建工具**    | Vite 8             | 极速开发服务器 + 构建              |
| **UI 组件库**  | Element Plus       | 企业级 Vue 3 组件库             |
| **样式**      | Tailwind CSS 4     | 原子化 CSS                   |
| **状态管理**    | Pinia 3            | Vue 官方状态管理                |
| **路由**      | Vue Router 4       | SPA 路由管理                  |
| **HTTP 请求** | Axios              | 封装拦截器、Token、错误处理          |
| **后端框架**    | Gin                | 高性能 Go Web 框架             |
| **ORM**     | GORM               | Go 语言 ORM                 |
| **数据库**     | SQLite             | 轻量级本地数据库（MVP）             |
| **认证**      | JWT + bcrypt       | Token 认证 + 密码加密           |
| **AI**      | OpenAI API (GPT)   | 需求解析 + 报价生成               |


---

## 数据模型

### User（用户）


| 字段        | 类型     | 说明                    |
| --------- | ------ | --------------------- |
| id        | uint   | 主键                    |
| email     | string | 邮箱（唯一索引）              |
| password  | string | 密码（bcrypt 加密，API 不返回） |
| name      | string | 姓名                    |
| company   | string | 公司名称                  |
| createdAt | time   | 创建时间                  |
| updatedAt | time   | 更新时间                  |


### Product（产品）


| 字段           | 类型      | 说明       |
| ------------ | ------- | -------- |
| id           | uint    | 主键       |
| userId       | uint    | 所属用户     |
| name         | string  | 产品名称     |
| sku          | string  | SKU / 型号 |
| category     | string  | 分类       |
| description  | text    | 简介       |
| material     | string  | 材质       |
| size         | string  | 尺寸       |
| color        | string  | 颜色       |
| process      | string  | 工艺       |
| packaging    | string  | 包装方式     |
| price        | float64 | 参考价格     |
| moq          | int     | 最小起订量    |
| leadTime     | string  | 默认交期     |
| paymentTerms | string  | 付款条款     |
| attachments  | int     | 附件数量     |


### Quote（报价）


| 字段               | 类型          | 说明                      |
| ---------------- | ----------- | ----------------------- |
| id               | uint        | 主键                      |
| userId           | uint        | 所属用户                    |
| quoteNumber      | string      | 报价编号（唯一，如 QT-2026-001）  |
| customerName     | string      | 客户名称                    |
| country          | string      | 国家/地区                   |
| currency         | string      | 币种（USD/EUR/CNY 等）       |
| deliveryAddress  | string      | 交付地址                    |
| status           | string      | 状态：草稿 / 已发送 / 待确认 / 已成交 |
| totalAmount      | float64     | 总金额                     |
| rawRequirement   | text        | 客户原始需求文本                |
| parsedParams     | text (JSON) | AI 解析后的结构化参数            |
| replyVersions    | text (JSON) | 生成的回复话术（3个版本）           |
| confirmationList | text (JSON) | 参数确认清单                  |
| attachmentList   | text (JSON) | 附件包列表                   |
| items            | []QuoteItem | 报价明细（一对多关联）             |


### QuoteItem（报价明细）


| 字段          | 类型      | 说明                       |
| ----------- | ------- | ------------------------ |
| id          | uint    | 主键                       |
| quoteId     | uint    | 所属报价                     |
| productName | string  | 产品名称                     |
| model       | string  | 型号                       |
| specs       | string  | 规格描述                     |
| quantity    | int     | 数量                       |
| unitPrice   | float64 | 单价                       |
| totalPrice  | float64 | 小计（quantity × unitPrice） |


### Template（模板）


| 字段       | 类型     | 说明                                         |
| -------- | ------ | ------------------------------------------ |
| id       | uint   | 主键                                         |
| userId   | uint   | 所属用户                                       |
| name     | string | 模板名称                                       |
| category | string | 分类：quotation / email / chat / confirmation |
| content  | text   | 模板内容（支持 `{{变量}}` 插值）                       |


---

## API 接口

### 认证 `/api/auth`


| 方法   | 路径                   | 说明         | 鉴权  |
| ---- | -------------------- | ---------- | --- |
| POST | `/api/auth/register` | 注册         | 否   |
| POST | `/api/auth/login`    | 登录（返回 JWT） | 否   |
| GET  | `/api/auth/profile`  | 获取当前用户信息   | 是   |


### 产品 `/api/products`


| 方法     | 路径                     | 说明          |
| ------ | ---------------------- | ----------- |
| GET    | `/api/products`        | 产品列表（分页/搜索） |
| GET    | `/api/products/:id`    | 产品详情        |
| POST   | `/api/products`        | 新建产品        |
| PUT    | `/api/products/:id`    | 更新产品        |
| DELETE | `/api/products/:id`    | 删除产品        |
| POST   | `/api/products/import` | Excel 批量导入  |


### 报价 `/api/quotes`


| 方法     | 路径                          | 说明                |
| ------ | --------------------------- | ----------------- |
| POST   | `/api/quotes/parse`         | AI 解析客户需求 → 结构化参数 |
| POST   | `/api/quotes/generate`      | 生成报价单 + 话术 + 确认清单 |
| POST   | `/api/quotes`               | 保存报价              |
| GET    | `/api/quotes`               | 报价列表（分页/筛选）       |
| GET    | `/api/quotes/:id`           | 报价详情              |
| PUT    | `/api/quotes/:id`           | 更新报价              |
| DELETE | `/api/quotes/:id`           | 删除报价              |
| POST   | `/api/quotes/:id/duplicate` | 基于旧报价新建           |


### 模板 `/api/templates`


| 方法     | 路径                   | 说明   |
| ------ | -------------------- | ---- |
| GET    | `/api/templates`     | 模板列表 |
| GET    | `/api/templates/:id` | 模板详情 |
| POST   | `/api/templates`     | 新建模板 |
| PUT    | `/api/templates/:id` | 更新模板 |
| DELETE | `/api/templates/:id` | 删除模板 |


### 工作台 `/api/dashboard`


| 方法  | 路径                      | 说明                    |
| --- | ----------------------- | --------------------- |
| GET | `/api/dashboard/stats`  | 今日统计（新报价/待确认/待发送/已发送） |
| GET | `/api/dashboard/recent` | 最近 10 条报价             |


### 文件上传 `/api/upload`


| 方法   | 路径            | 说明                      |
| ---- | ------------- | ----------------------- |
| POST | `/api/upload` | 上传文件（图片/PDF/Excel/Word） |


> 除注册和登录外，所有接口需在 Header 中携带 `Authorization: Bearer <token>`。

---

## 前端路由


| 路径                | 页面    | 说明               |
| ----------------- | ----- | ---------------- |
| `/login`          | 登录/注册 | 无需登录             |
| `/dashboard`      | 工作台   | 统计 + 最近报价 + 快速入口 |
| `/quotes/new`     | 新建报价  | **核心页面**，三栏布局    |
| `/quotes/:id`     | 报价详情  | 查看/编辑已有报价        |
| `/quotes/history` | 历史报价  | 筛选/搜索/导出/复制新建    |
| `/products`       | 产品资料库 | 列表/搜索/导入         |
| `/products/:id`   | 产品详情  | 查看/编辑产品信息        |
| `/templates`      | 模板管理  | 分类/编辑/变量插入       |
| `/settings`       | 设置    | 个人信息 + AI 配置     |


---

## 环境要求

- **Node.js** >= 18
- **Go** >= 1.21
- **GCC**（SQLite 编译需要，Windows 可用 [tdm-gcc](https://jmeubank.github.io/tdm-gcc/)）

---

## 快速开始

### 1. 克隆项目

```bash
git clone <repo-url>
cd LamMap
```

### 2. 启动后端

```bash
cd backend

# （可选）配置环境变量
# Windows PowerShell:
$env:OPENAI_API_KEY="sk-your-api-key"
$env:OPENAI_MODEL="gpt-4o-mini"

# Linux/Mac:
# export OPENAI_API_KEY=sk-your-api-key

# 启动服务
go run main.go
```

后端默认运行在 `http://localhost:8080`，首次启动会自动创建 SQLite 数据库文件 `quotepro.db` 并完成表结构迁移。

### 3. 启动前端

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端默认运行在 `http://localhost:3000`，所有 `/api` 请求会自动代理到后端 8080 端口。

### 4. 访问系统

打开浏览器访问 `http://localhost:3000`，注册账号后即可使用。

---

## 环境变量


| 变量名              | 默认值                                        | 说明                     |
| ---------------- | ------------------------------------------ | ---------------------- |
| `PORT`           | `8080`                                     | 后端服务端口                 |
| `DB_PATH`        | `quotepro.db`                              | SQLite 数据库文件路径         |
| `JWT_SECRET`     | `quotepro-secret-key-change-in-production` | JWT 签名密钥（**生产环境必须修改**） |
| `OPENAI_API_KEY` | （空）                                        | OpenAI API Key         |
| `OPENAI_MODEL`   | `gpt-4o-mini`                              | 使用的 GPT 模型             |


---

## 构建部署

### 前端构建

```bash
cd frontend
npm run build
# 产出 dist/ 目录，可部署到 Nginx / CDN
```

### 后端构建

```bash
cd backend
go build -o quotepro-server .
./quotepro-server
```

---

## 核心使用流程

```
1. 注册/登录
       ↓
2. 点击「生成新报价」
       ↓
3. 粘贴客户需求（微信/邮件/WhatsApp 内容）或上传文件
       ↓
4. 点击「解析需求」→ AI 自动提取结构化参数
       ↓
5. 检查/修改识别结果（可编辑每个参数字段）
       ↓
6. 点击「开始生成」→ 生成 4 项内容：
   ├── 报价单（可行内编辑价格/数量）
   ├── 回复话术（3 版本 × 4 种语气）
   ├── 参数确认清单（中英双语，一键复制）
   └── 附件包（勾选推荐资料，打包下载）
       ↓
7. 导出 PDF / 复制消息 / 发送客户
```

---

## 开源许可

- **中文说明（默认）**：[许可证.md](许可证.md)（含 **AI 调用：每月超过 2000 次需事先授权** 的说明）
- **英文条文本**（Apache 2.0 全文 + 同上补充条款英文版）：[LICENSE](LICENSE)

---

## 开发进度

详见 `[doc/开始任务.md](doc/开始任务.md)`，其中已完成的任务用 `[x]` 标记，待完成的用 `[ ]` 标记。