# VPSPing - VPS 延迟监控工具

一个轻量级的命令行工具，用于周期性测试多个 VPS 服务器的网络延迟，帮助用户监控 VPS 的网络质量。

## 功能特性

- **ICMP Ping 测试**：使用 ICMP 协议进行延迟测试
- **并发测试**：支持同时测试多个 VPS
- **数据持久化**：使用 SQLite 数据库存储历史数据
- **统计分析**：计算平均、最高、最低延迟和丢包率
- **可视化展示**：命令行 ASCII 折线图显示延迟趋势
- **多种输出**：支持日志文件和 JSON 格式输出
- **灵活配置**：YAML 配置文件管理 VPS 列表

## 安装

### 从源码编译

```bash
git clone <repository-url>
cd VPSPing
go build -o vpsping ./cmd/vpsping
```

### 使用 Docker

```bash
# 构建镜像（使用默认代理）
docker build -t vpsping:latest .

# 构建镜像（自定义代理）
docker build \
  --build-arg GOPROXY=https://goproxy.io,direct \
  --build-arg ALPINE_MIRROR=mirrors.tuna.tsinghua.edu.cn \
  -t vpsping:latest .

# 运行容器
docker run -d \
  --name vpsping \
  --restart unless-stopped \
  -v $(pwd)/config:/app/config:ro \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/output:/app/output \
  vpsping:latest
```

#### 代理配置说明

Dockerfile 默认使用以下代理：
- **Go 模块代理**: `https://goproxy.cn,direct`
- **Alpine 镜像源**: `mirrors.aliyun.com`（仅域名）

可用的代理选项：
- **Go 模块代理**:
  - `https://goproxy.cn,direct` (七牛云)
  - `https://goproxy.io,direct` (全球加速)
  - `https://proxy.golang.org,direct` (官方)
  
- **Alpine 镜像源**（仅域名，不包含协议和路径）:
  - `mirrors.aliyun.com` (阿里云)
  - `mirrors.tuna.tsinghua.edu.cn` (清华)
  - `mirrors.ustc.edu.cn` (中科大)

### 使用 Docker Compose

```bash
# 复制环境配置文件
cp .env.example .env

# 根据需要修改配置
vim .env

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重新构建
docker-compose build --no-cache
```

### 系统要求

- Go 1.21 或更高版本
- macOS / Linux / Windows
- ICMP Ping 可能需要 root/管理员权限

## 快速开始

### 1. 初始化配置文件

```bash
./vpsping init
```

这将创建一个示例配置文件 `config.yaml`。

### 2. 编辑配置文件

编辑 `config.yaml`，添加你的 VPS 信息：

```yaml
vps:
  - name: "Tokyo-1"
    host: "tokyo1.example.com"
    enabled: true
  - name: "Singapore-1"
    host: "sg1.example.com"
    enabled: true

ping:
  interval: "15m"      # 测试间隔
  count: 4             # 每次测试的 Ping 次数
  timeout: "5s"        # 超时时间
  privileged: false    # 是否使用特权模式

storage:
  database: "./data/vpsping.db"
  log_file: "./logs/vpsping.log"
  json_output: "./output/results.json"

display:
  chart_width: 80
  chart_height: 20
  time_range: "24h"
```

### 3. 执行测试

```bash
# 执行一次测试
./vpsping test

# 启动持续监控
./vpsping run
```

## 命令说明

### vpsping run

启动持续监控，按照配置的间隔时间定期测试所有 VPS 的延迟。

```bash
./vpsping run [-c config.yaml]
```

### vpsping test

对所有启用的 VPS 执行一次延迟测试。

```bash
./vpsping test [-c config.yaml]
```

### vpsping stats

显示指定 VPS 或所有 VPS 的统计信息。

```bash
# 显示所有 VPS 的统计信息
./vpsping stats

# 显示指定 VPS 的统计信息
./vpsping stats Tokyo-1
```

### vpsping chart

显示延迟趋势图。

```bash
# 显示所有 VPS 的趋势图
./vpsping chart

# 显示指定 VPS 的趋势图
./vpsping chart Tokyo-1
```

### vpsping list

列出配置中的所有 VPS 及其状态。

```bash
./vpsping list
```

### vpsping add

添加一个新的 VPS 到监控列表。

```bash
./vpsping add <name> <host>
```

示例：
```bash
./vpsping add Tokyo-2 tokyo2.example.com
```

### vpsping remove

从监控列表中删除指定的 VPS。

**注意**：删除 VPS 时会同时删除该 VPS 的所有历史测试数据和统计信息。

```bash
./vpsping remove <name>
```

示例：
```bash
./vpsping remove Tokyo-2
```

## 持续运行

`vpsping run` 命令会持续运行并定期测试 VPS 延迟。要让它在后台持续运行，有多种方式：

### 快速启动（使用 nohup）

```bash
nohup ./vpsping run > vpsping.out 2>&1 &

# 查看输出
tail -f vpsping.out

# 停止
ps aux | grep vpsping
kill <PID>
```

### 生产环境部署

详细的部署方案请参考 [持续运行指南](docs/RUNNING.md)，包括：

- **systemd**（Linux 推荐）
- **launchd**（macOS 推荐）
- **Docker**（容器化部署）
- **screen/tmux**（开发环境）
- **Supervisor**（Python 环境）

### vpsping init

创建一个示例配置文件。

```bash
./vpsping init
```

## 配置文件搜索路径

当不指定 `-c` 参数时，程序会按以下优先级顺序搜索配置文件：

1. **用户主目录**（最高优先级）
   - 路径：`~/.vpsping/config.yaml`
   - 适用：全局配置，多项目共享

2. **当前工作目录**
   - 路径：`./config.yaml`
   - 适用：项目特定配置

3. **config 子目录**（最低优先级）
   - 路径：`./config/config.yaml`
   - 适用：项目推荐的标准位置

也可以使用 `-c` 参数明确指定配置文件路径：

```bash
./vpsping test -c /path/to/config.yaml
```

## 输出示例

### 测试结果表格

```
[2024-01-15 10:30:00] Testing VPS servers...

┌─────────────┬──────────┬──────────┬──────────┬─────┬──────────┬────────┐
│  VPS NAME   │ AVG (MS) │ MIN (MS) │ MAX (MS) │ TTL │ LOSS (%) │ STATUS │
├─────────────┼──────────┼──────────┼──────────┼─────┼──────────┼────────┤
│ Tokyo-1     │ 45.2     │ 42.1     │ 48.9     │ 54  │ 0.0      │ OK     │
│ Singapore-1 │ 78.5     │ 75.2     │ 82.3     │ 48  │ 0.0      │ OK     │
│ US-West-1   │ 156.3    │ 148.7    │ 162.1    │ 42  │ 2.5      │ LOSS   │
└─────────────┴──────────┴──────────┴──────────┴─────┴──────────┴────────┘
```

### 统计信息

```
Statistics for Tokyo-1 (Last 24 hours):
  Average Latency: 46.8 ms
  Minimum Latency: 38.2 ms
  Maximum Latency: 89.5 ms
  Average TTL: 54
  Packet Loss: 0.3%
  Total Tests: 96
```

### 延迟趋势图

```
Latency Trend for Tokyo-1 (Last 24 hours)
  100ms ┤
   90ms ┤                    ╭─╮
   80ms ┤              ╭────╯ ╰╮
   70ms ┤         ╭───╯       ╰──╮
   60ms ┤    ╭───╯               ╰──╮
   50ms ┤╭──╯                       ╰─╮
   40ms ┤╯                            ╰──
   30ms ┤
        └─────────────────────────────────
        00:00  04:00  08:00  12:00  16:00  20:00
```

## 配置说明

### VPS 配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | VPS 名称，用于标识 |
| host | string | 是 | VPS 主机地址（IP 或域名） |
| enabled | bool | 否 | 是否启用，默认为 true |

### Ping 配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| interval | string | 15m | 测试间隔时间 |
| count | int | 4 | 每次测试的 Ping 次数 |
| timeout | string | 5s | 超时时间 |
| privileged | bool | false | 是否使用特权模式（需要 root 权限） |

### 存储配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| database | string | ./data/vpsping.db | SQLite 数据库文件路径 |
| log_file | string | ./logs/vpsping.log | 日志文件路径 |
| json_output | string | ./output/results.json | JSON 输出文件路径 |

### 显示配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| chart_width | int | 80 | 图表宽度（字符数） |
| chart_height | int | 20 | 图表高度（行数） |
| time_range | string | 24h | 统计和图表的时间范围 |

## 注意事项

### ICMP 权限

- **非特权模式**（`privileged: false`）：使用 UDP 进行 Ping 测试，不需要 root 权限，但某些网络环境可能不支持
- **特权模式**（`privileged: true`）：使用原始 ICMP 套接字，需要 root/管理员权限

### macOS 用户

在 macOS 上，非特权模式可能无法正常工作。建议使用特权模式：

```yaml
ping:
  privileged: true
```

然后使用 sudo 运行：

```bash
sudo ./vpsping test
```

### 数据存储

- SQLite 数据库会随着时间增长，建议定期清理历史数据
- 日志文件会持续追加，建议配置日志轮转

## 项目结构

```
VPSPing/
├── cmd/
│   └── vpsping/
│       └── main.go          # 主程序入口
├── internal/
│   ├── config/              # 配置管理
│   ├── models/              # 数据模型
│   ├── pinger/              # Ping 测试功能
│   ├── storage/             # 数据存储
│   ├── output/              # 输出处理
│   ├── stats/               # 统计分析
│   └── scheduler/           # 调度器
├── config/
│   └── config.yaml          # 示例配置文件
├── docs/
│   ├── PRD.md               # 产品需求文档
│   └── TODO.md              # 任务拆解文档
├── data/                    # 数据库文件目录
├── logs/                    # 日志文件目录
├── output/                  # JSON 输出目录
├── go.mod
├── go.sum
└── README.md
```

## 技术栈

- **语言**：Go 1.21+
- **Ping 库**：[github.com/prometheus-community/pro-bing](https://github.com/prometheus-community/pro-bing)
- **命令行框架**：[github.com/spf13/cobra](https://github.com/spf13/cobra)
- **配置管理**：[github.com/spf13/viper](https://github.com/spf13/viper)
- **ORM**：[gorm.io/gorm](https://gorm.io/gorm)
- **数据库**：SQLite 3
- **表格输出**：[github.com/olekukonko/tablewriter](https://github.com/olekukonko/tablewriter)
- **终端颜色**：[github.com/fatih/color](https://github.com/fatih/color)

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
