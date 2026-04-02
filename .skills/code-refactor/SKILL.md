---
name: code-refactor
description: "智能代码重构助手，用于分析和改进现有代码结构、性能和可维护性"
---

# 代码重构技能

## 技能描述

此技能帮助开发者安全地重构代码，提供最佳实践建议、自动化重构方案和代码质量改进建议。

## 使用场景

- 优化代码结构和可读性
- 消除重复代码（DRY原则）
- 改善代码性能
- 提升代码可维护性
- 遵循设计模式和最佳实践

## 重构流程

### 1. 分析阶段

```python
def analyze_code(code: str, language: str) -> dict:
    """
    分析代码质量问题
    """
    issues = []
    suggestions = []

    # 检查重复代码
    # 检查过长函数/方法
    # 检查过深嵌套
    # 检查命名规范
    # 检查代码复杂度

    return {
        "issues": issues,
        "suggestions": suggestions,
        "complexity_score": 0
    }
```

### 2. 重构策略

#### 常见重构模式

**提取方法（Extract Method）**

```python
# 重构前
def process_order(order):
    # 验证订单
    if not order.get('items'):
        raise ValueError("订单为空")
    if order.get('total') <= 0:
        raise ValueError("金额无效")

    # 计算税费
    tax = order['total'] * 0.1
    # 计算折扣
    discount = order['total'] * 0.05 if order['total'] > 100 else 0

    # 更新库存
    for item in order['items']:
        update_inventory(item['id'], item['quantity'])

    # 发送通知
    send_email(order['customer_email'], f"订单金额: {order['total'] + tax - discount}")

    return order['total'] + tax - discount

# 重构后
def process_order(order):
    _validate_order(order)
    tax = _calculate_tax(order['total'])
    discount = _calculate_discount(order['total'])
    _update_inventory(order['items'])
    _send_notification(order['customer_email'], order['total'], tax, discount)
    return order['total'] + tax - discount

def _validate_order(order):
    if not order.get('items'):
        raise ValueError("订单为空")
    if order.get('total') <= 0:
        raise ValueError("金额无效")

def _calculate_tax(amount):
    return amount * 0.1

def _calculate_discount(amount):
    return amount * 0.05 if amount > 100 else 0

def _update_inventory(items):
    for item in items:
        update_inventory(item['id'], item['quantity'])

def _send_notification(email, total, tax, discount):
    send_email(email, f"订单金额: {total + tax - discount}")
```

**合并重复代码**

```python
# 重构前
def process_payment_credit(card_number, amount):
    validate_card(card_number)
    process_transaction(card_number, amount)
    log_transaction("credit", card_number, amount)

def process_payment_debit(account_number, amount):
    validate_account(account_number)
    process_transaction(account_number, amount)
    log_transaction("debit", account_number, amount)

# 重构后
def process_payment(payment_type, identifier, amount):
    validators = {
        "credit": validate_card,
        "debit": validate_account
    }

    validator = validators.get(payment_type)
    if validator:
        validator(identifier)
        process_transaction(identifier, amount)
        log_transaction(payment_type, identifier, amount)
```

### 3. 安全重构检查清单

- [ ] 运行现有测试套件确保功能不变
- [ ] 小步提交，每次重构后立即测试
- [ ] 保持外部API不变（除非有计划）
- [ ] 更新相关文档和注释
- [ ] 检查性能影响
- [ ] 代码审查

## 重构决策树

```
是否有重复代码？
├─ 是 → 提取公共方法/类
└─ 否 → 检查方法长度
    ├─ > 50行 → 拆分为小方法
    └─ ≤ 50行 → 检查复杂度
        ├─ 圈复杂度 > 10 → 简化条件逻辑
        └─ 圈复杂度 ≤ 10 → 检查命名
            ├─ 命名不清晰 → 重命名
            └─ 命名清晰 → 检查耦合度
                ├─ 高耦合 → 应用依赖注入
                └─ 低耦合 → 完成
```

## 代码坏味道检测

### 必须重构的信号

1. **重复代码** - 相同代码出现多次
2. **过长函数** - 超过50行（或20行核心逻辑）
3. **过大类** - 超过500行
4. **过长参数列表** - 超过4个参数
5. **发散式变化** - 一个类因多个原因变化
6. **霰弹式修改** - 一个变化影响多个类
7. **依恋情结** - 方法过度使用其他类的数据
8. **数据泥团** - 经常一起出现的数据组

## 重构示例

### 示例1：简化条件表达式

```python
# 重构前
def get_discount(customer, order):
    if customer.is_premium:
        if order.amount > 1000:
            return 0.2
        elif order.amount > 500:
            return 0.15
        else:
            return 0.1
    else:
        if order.amount > 1000:
            return 0.1
        elif order.amount > 500:
            return 0.05
        else:
            return 0

# 重构后
def get_discount(customer, order):
    discount_rules = {
        (True, lambda a: a > 1000): 0.2,
        (True, lambda a: a > 500): 0.15,
        (True, lambda a: True): 0.1,
        (False, lambda a: a > 1000): 0.1,
        (False, lambda a: a > 500): 0.05,
        (False, lambda a: True): 0
    }

    for (is_premium, condition), discount in discount_rules.items():
        if customer.is_premium == is_premium and condition(order.amount):
            return discount
    return 0
```

### 示例2：引入参数对象

```python
# 重构前
def create_user(name, email, phone, address, city, zipcode, country):
    pass

def update_user(user_id, name, email, phone, address, city, zipcode, country):
    pass

# 重构后
class Address:
    def __init__(self, street, city, zipcode, country):
        self.street = street
        self.city = city
        self.zipcode = zipcode
        self.country = country

class UserInfo:
    def __init__(self, name, email, phone, address):
        self.name = name
        self.email = email
        self.phone = phone
        self.address = address

def create_user(user_info: UserInfo):
    pass

def update_user(user_id, user_info: UserInfo):
    pass
```

## 开始执行重构

### 根据分析，使用可用工具，执行重构

## 重构完成后生成,重构分析报告

### 发现的问题

1. [问题类型] 问题描述 (位置: 行号)
2. ...

### 建议的重构方案

#### 重构1: [重构名称]

- **原因**: 为什么要重构
- **步骤**: 具体操作步骤
- **风险**: 潜在风险及应对

### 重构前后对比

[展示代码对比]

### 测试建议

- [ ] 测试用例1
- [ ] 测试用例2

### 预期收益

- 可读性提升: [描述]
- 可维护性提升: [描述]
- 性能影响: [正面/负面/中性]

```

## 注意事项

1. **不要过度重构** - 保持简单，只重构真正需要改进的部分
2. **保持测试覆盖** - 重构前确保有足够的测试
3. **渐进式改进** - 小步快跑，持续集成
4. **团队协作** - 重构前与团队沟通，避免冲突
5. **文档同步** - 更新相关文档和注释
```
