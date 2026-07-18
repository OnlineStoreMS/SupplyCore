# SupplyCore — 供应商与采购管理（VMS / PMS）

独立应用，与 [UserCore](../UserCore)（IAM）、[ProductCore](../ProductCore)（PIM）并列部署。

| 组件 | 端口 | 说明 |
|------|------|------|
| API | 8092 | Go + Gin + GORM |
| Web | 5175 | Vue 3 + Element Plus |

## 当前能力

### M1 · VMS
- 供应商 CRUD、发货地址、SKU 供货报价
- `GET /api/v1/admin/skus/{skuId}/supply-options`

### M2 · PMS（采购单核心）
- 采购单 `purchase_orders` / 明细 `purchase_order_items`
- 状态：草稿 → 已下单 → 已付款 → 已完成 / 取消
- 从供货报价带入单价与货号
- 前端：采购单列表、新建/编辑、详情与状态操作

### M3 · 物流 / 付款 / 附件
- `purchase_shipments`、`purchase_payments`、`purchase_attachments`
- 物流单号、预计到货、运输中状态自动同步 PO
- 多次付款记录，自动汇总 PO 付款状态（未付 / 部分 / 已付）
- MinIO 附件上传（供货商销售单、付款截图等，`storage.driver: minio`）
- 采购单详情 Tab：物流 / 付款 / 附件

### M4 · 采购闭环（对齐普源云 ERP 采购）
- **采购建议**：缺货 / 预警 / 无库存（对接 WarehouseCore 后生效）
- **采购账号**：1688 / 淘供销 / 其他渠道账号管理
- **收包入库**：收货记录、包裹扫描入库（可生成入库单）
- **采购入库**：入库单、入库审核、财务审核；回写采购单 `receivedQty`
- **采购入库分拣**：SKU 扫描作业台（骨架）
- **采购退回**：退回单、退回审核、财务审核
- 采购单可选写入外部订单引用 `ref_so_id` / `ref_trace_id`（供 OMS 对接）

## 快速开始

```bash
# 1. 数据库（PostgreSQL）
make init-db APP_PASSWORD=你的密码
# 或手动创建 dbname=supplycore

# 2. 配置
cp configs/config.example.yaml configs/config.yaml
# jwt_secret 必须与 UserCore 一致

# 3. 后端
make tidy
make run

# 4. 前端
cd web && npm install && npm run dev
```

浏览器访问 http://localhost:5175 ，从 UserCore 应用中心进入（需配置 SupplyCore 应用）。

## 数据库权限问题

若启动报错 `relation "suppliers" already exists` 或 `must be owner of table suppliers`，说明表由 `postgres` 用户创建，而应用使用 `supplycore` 连接。GORM 在 `information_schema` 中看不到已有表，会误判为需要建表。

以超级用户执行：

```bash
chmod +x deploy/fix_db_permissions.sh
./deploy/fix_db_permissions.sh
```

## UserCore 注册

在 UserCore 应用中心应出现 **供应链中心**。新环境 seed 会自动写入；已有环境启动 UserCore 时会 `EnsureApps` upsert。

权限码：

- `supply:read` — 查看供应商与报价
- `supply:write` — 编辑供应商、地址、报价

## 环境变量（前端）

| 变量 | 默认 |
|------|------|
| `VITE_PORTAL_URL` | `http://localhost:5174` |
| `VITE_API_GATEWAY` | 未设置则直连 8092 |
