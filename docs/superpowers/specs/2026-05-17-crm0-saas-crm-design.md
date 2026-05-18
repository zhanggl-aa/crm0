# CRM0 - SaaS 订阅业务 CRM 系统设计

## 概述

面向 SaaS/订阅业务的 CRM 系统，核心卖点是四大智能算法：客户流失预测、智能客户分层、LTV 预测与获客优化、下一步最佳行动推荐。通过算法驱动而非单纯数据管理来吸引客户购买。

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue3 + TypeScript + Vite + Element Plus + ECharts + Pinia |
| API 服务 | Go + Fiber |
| 算法服务 | Python + FastAPI + scikit-learn + XGBoost + lifelines |
| 数据库 | PostgreSQL (crm0_db, user: postgres, password: 123456) |
| 缓存 | Redis (模型缓存、计算结果) |
| 部署 | 本地开发优先 |

## 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                    Vue3 + TypeScript                      │
│                  (Vite · Pinia · Element Plus)            │
└──────────────────────────┬──────────────────────────────┘
                           │ HTTP/REST
┌──────────────────────────▼──────────────────────────────┐
│                  Go API Gateway (Fiber)                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────┐ │
│  │ 认证/权限  │ │ 客户管理  │ │ 订阅管理  │ │ 数据分析API │ │
│  └──────────┘ └──────────┘ └──────────┘ └─────┬──────┘ │
│                              JWT · RBAC · 限流           │
└──────────────────────────┬──────────────────────────────┘
               │            │
               │ HTTP REST   │ 内部HTTP (算法调用)
               ▼            ▼
┌──────────────────┐  ┌────────────────────────────────────┐
│   PostgreSQL      │  │     Python Algorithm Service        │
│   (crm0_db)       │  │          (FastAPI)                  │
│ ┌──────────────┐ │  │  ┌──────────────────────────────┐  │
│ │ 客户/订阅/    │ │  │  │ 流失预测 (XGBoost)            │  │
│ │ 事件/行为     │ │  │  │ 智能分层 (K-Means/RFM)       │  │
│ │ 算法结果      │ │  │  │ LTV预测 (生存分析)            │  │
│ └──────────────┘ │  │  │ NBA推荐 (协同过滤)            │  │
└──────────────────┘  │  └──────────────────────────────┘  │
                       │  Redis (模型缓存 · 计算结果)       │
                       └────────────────────────────────────┘
```

**核心设计原则：**
- Go 是唯一对外暴露的入口，Python 服务不直接对外，只接受 Go 的内部调用
- 算法结果写回 PostgreSQL，前端通过 Go API 查询
- 算法服务支持异步任务：Go 发起计算请求 → Python 返回 task_id → Go 轮询/WebSocket 推送结果
- Redis 缓存算法计算结果和模型预测，避免重复计算

## 数据模型

### 核心实体关系

```
┌─────────┐     1:N     ┌──────────────┐     1:N     ┌─────────────┐
│  Tenant  │────────────│   Customer    │────────────│  Subscription│
│ (租户)   │            │  (客户)       │            │  (订阅)      │
└─────────┘            └──────┬───────┘            └─────────────┘
                              │ 1:N
                       ┌──────▼───────┐
                       │  UserEvent   │
                       │ (行为事件)    │
                       └──────────────┘

-- 算法结果独立存储
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ChurnPrediction│ │CustSegment   │ │LTVPrediction │ │NBARecommend  │
│ 流失预测结果   │ │ 客户分层结果  │ │ LTV预测结果   │ │ 行动推荐结果  │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

### 核心表

| 表 | 关键字段 | 说明 |
|---|---|---|
| `tenants` | id, name, plan, settings(JSONB) | 多租户隔离，settings 存租户级配置 |
| `users` | id, tenant_id, email, role, password_hash | RBAC: admin/manager/member |
| `customers` | id, tenant_id, name, email, company, tags(JSONB), custom_fields(JSONB) | custom_fields 支持租户自定义字段 |
| `subscriptions` | id, customer_id, plan_id, status, started_at, canceled_at, mrr | status: active/past_due/canceled/trial |
| `plans` | id, tenant_id, name, price, billing_cycle | 订阅计划定义 |
| `user_events` | id, customer_id, event_type, properties(JSONB), occurred_at | 行为事件流：login/feature_use/support_ticket/payment_failed 等 |
| `churn_predictions` | id, customer_id, risk_score, risk_level, factors(JSONB), predicted_at | risk_level: high/medium/low |
| `customer_segments` | id, customer_id, segment_type, segment_name, score, updated_at | segment_type: rfm/behavioral/value |
| `ltv_predictions` | id, customer_id, predicted_ltv, confidence, model_version, predicted_at | 含置信区间 |
| `nba_recommendations` | id, customer_id, action_type, action_detail(JSONB), expected_impact, priority | action_type: email/call/discount/feature_guide/onboarding_check |

### 多租户策略

共享数据库 + tenant_id 隔离（所有查询带 tenant_id 过滤），适合中小企业规模，成本低。

## 四大核心算法

### 1. 客户流失预测 (Churn Prediction)

- **输入**：客户近90天的行为事件（登录频率、功能使用、支持工单、支付失败次数）+ 订阅状态 + 订阅时长
- **模型**：XGBoost 二分类（流失/留存）
- **特征工程**：登录频率趋势（7/14/30天滑动窗口）、功能使用广度、支付失败率、支持工单情感分、订阅天数、MRR变化率
- **输出**：risk_score (0-1)、risk_level (high/medium/low)、top 3 流失因素
- **触发**：每日定时批量计算 + 客户关键事件触发增量更新

### 2. 智能客户分层 (Customer Segmentation)

- **RFM分层**：基于 Recency（最近活跃）、Frequency（交互频率）、Monetary（MRR贡献）三个维度，K-Means 聚类为 4-6 个群体
- **行为分层**：基于 feature_usage 向量，DBSCAN 密度聚类（自动发现群体数量）
- **价值分层**：基于 LTV 预测值 + 当前 MRR，分位数分层
- **输出**：每客户一个主分层 + 多个子维度标签，如 "高价值·近期活跃·重度使用者"

### 3. LTV预测与获客优化 (LTV Prediction)

- **模型**：生存分析（Kaplan-Meier + Cox 比例风险模型），预测客户在不同时间点的留存概率，结合 MRR 计算期望 LTV
- **输入**：订阅历史、行为特征、客户属性
- **输出**：predicted_ltv、confidence_interval、expected_lifetime_months
- **获客优化**：按获客渠道计算 CAC vs 预测 LTV，输出各渠道的 LTV/CAC 比值，推荐预算分配

### 4. 下一步最佳行动 (Next Best Action)

- **方法**：基于客户当前状态（分层+流失风险+生命周期阶段），用关联规则挖掘 + 协同过滤推荐最佳行动
- **行动空间**：email（关怀邮件）、call（人工回访）、discount（优惠续费）、feature_guide（功能引导）、onboarding_check（入驻检查）
- **输出**：top 3 推荐行动，含 expected_impact（预期提升留存率的百分点）和 priority

### 算法服务通用设计

- 模型训练：离线批量（每日凌晨），数据从 PostgreSQL 读取
- 模型推理：在线推理（Go 调用 → Python 返回结果），缓存到 Redis
- 模型版本管理：每次训练保存模型文件 + 版本号，支持回滚
- 冷启动：新租户前30天用规则引擎替代ML，积累足够数据后自动切换

## Go API 服务设计

### 目录结构

```
crm0/
├── backend/                    # Go 后端
│   ├── cmd/
│   │   └── server/main.go     # 入口
│   ├── internal/
│   │   ├── config/            # 配置加载
│   │   ├── middleware/         # JWT认证·RBAC·限流·租户隔离
│   │   ├── handler/           # Fiber handlers (controller层)
│   │   ├── service/           # 业务逻辑层
│   │   ├── repository/        # 数据访问层
│   │   ├── model/             # 数据模型 (entity + DTO)
│   │   └── algorithm/         # Python服务客户端
│   ├── migrations/            # SQL迁移文件
│   ├── go.mod
│   └── go.sum
├── algorithm/                  # Python 算法服务
│   ├── app/
│   │   ├── main.py            # FastAPI 入口
│   │   ├── routers/           # 各算法路由
│   │   ├── services/          # 算法实现
│   │   ├── models/            # Pydantic模型
│   │   └── ml/                # ML模型训练/推理/持久化
│   ├── requirements.txt
│   └── Dockerfile
├── frontend/                   # Vue3 前端
│   ├── src/
│   │   ├── views/             # 页面
│   │   ├── components/        # 组件
│   │   ├── stores/            # Pinia stores
│   │   ├── api/               # API调用层
│   │   ├── router/            # 路由
│   │   └── utils/             # 工具函数
│   ├── package.json
│   └── vite.config.ts
└── docs/
```

### API 路由

```
/api/v1
├── /auth          POST login, POST register, POST refresh
├── /customers     CRUD + GET /:id/events + GET /:id/insights
├── /subscriptions CRUD + GET /metrics (MRR, churn rate等)
├── /plans         CRUD
├── /events        POST track, GET /timeline
├── /analytics
│   ├── /churn       GET predictions, POST trigger-prediction
│   ├── /segments    GET segments, POST trigger-segmentation
│   ├── /ltv         GET predictions, GET channel-roi
│   └── /nba         GET recommendations, POST trigger-nba
└── /dashboard     GET overview (聚合所有算法关键指标)
```

### 认证与多租户

- JWT 认证，token 中携带 tenant_id 和 role
- 中间件自动从 token 提取 tenant_id，所有查询自动注入租户过滤
- RBAC 三级：admin（全权限）、manager（管理+分析）、member（只读+客户操作）

### Go → Python 通信

- 同步调用：简单预测（单客户流失风险查询）直接 HTTP 调用，超时5s
- 异步任务：批量计算（全量分层、训练模型）→ Go 发起 POST → Python 返回 task_id → Go 轮询 GET /tasks/:id 或 WebSocket 推送

## 前端设计

### 页面结构

```
/login
/register
├── /dashboard              # 首页总览
├── /customers              # 客户列表（支持分层筛选）
│   ├── /:id                # 客户详情（含算法洞察卡片）
├── /subscriptions          # 订阅管理
├── /analytics
│   ├── /churn              # 流失预测看板（风险分布图、高危客户列表）
│   ├── /segments           # 客户分层看板（散点图、各群体画像）
│   ├── /ltv                # LTV看板（预测趋势、渠道ROI对比）
│   └── /nba                # 行动推荐（推荐列表、优先级排序）
└── /settings               # 租户设置、用户管理
```

### 技术选型

- **UI 框架**：Element Plus — 企业级组件库，表格/表单/图表开箱即用
- **图表**：ECharts — 流失风险分布、RFM散点图、LTV趋势、渠道ROI对比
- **状态管理**：Pinia
- **路由**：Vue Router 4
- **HTTP**：Axios + 拦截器（自动带JWT、租户头、错误处理）
- **布局**：左侧导航 + 顶栏 + 内容区，经典后台布局

### 关键交互

- **Dashboard**：顶部4个指标卡（总客户数、MRR、流失率、平均LTV），下方各算法的快捷入口和最新预警
- **客户详情页**：右侧展示算法洞察卡片（流失风险仪表盘、分层标签、LTV预测、推荐行动），可一键执行推荐
- **分析看板**：每个算法一个独立页面，含可视化图表 + 可操作的数据表格

## 错误处理

- **Go API**：统一错误响应格式 `{ code, message, details }`，业务错误用自定义 error type，算法服务不可达时返回降级结果（缓存数据或规则引擎兜底）
- **Python 算法**：数据不足时返回 `insufficient_data` 错误（需至少30天事件数据），模型加载失败时自动降级到上一版本
- **前端**：Axios 拦截器统一处理 401/403/500，算法计算中展示 loading 骨架屏

## 测试策略

- **Go**：单元测试（service 层 mock repository）+ API 集成测试（testcontainers 跑真实 PostgreSQL）
- **Python**：算法单元测试（mock 数据验证模型输入输出）+ 端到端测试（真实数据集验证预测准确率）
- **前端**：Vitest 单元测试 + Playwright E2E 关键路径

## 开发启动流程

1. `createdb -U postgres crm0_db` 建库
2. Go: `cd backend && go run cmd/server/main.go` (自动运行 migration)
3. Python: `cd algorithm && pip install -r requirements.txt && uvicorn app.main:app --port 8001`
4. 前端: `cd frontend && npm install && npm run dev`
