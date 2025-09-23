# WSL 系统调试指南

## 问题描述
您遇到的 `Could not execute systemctl: at /usr/bin/deb-systemd-invoke line 148` 错误是 WSL 环境中常见的 systemd 兼容性问题。

## 快速解决方案

### 步骤 1: 进入 WSL 环境
首先确保您在 WSL Ubuntu 环境中，而不是 Windows Git Bash：

```bash
# 在 Windows PowerShell 或 CMD 中运行
wsl

# 或者指定发行版
wsl -d Ubuntu
```

### 步骤 2: 运行诊断脚本
```bash
# 赋予执行权限
chmod +x wsl_debug.sh

# 运行诊断
./wsl_debug.sh
```

### 步骤 3: 运行修复脚本
```bash
# 赋予执行权限
chmod +x wsl_fix.sh

# 以 root 权限运行修复
sudo ./wsl_fix.sh
```

### 步骤 4: 重启 WSL
在 Windows PowerShell 中运行：
```powershell
wsl --shutdown
```

然后重新进入 WSL：
```powershell
wsl
```

## 详细修复说明

### 1. systemd 支持配置
修复脚本会创建 `/etc/wsl.conf` 文件启用 systemd：
```ini
[boot]
systemd=true

[user]
default=root

[network]
generateHosts=true
generateResolvConf=true
```

### 2. openssh-server 修复
- 清理损坏的包状态
- 临时替换 systemctl 以跳过服务启动
- 重新配置 openssh-server
- 恢复原始 systemctl

### 3. 服务管理
创建 `/usr/local/bin/wsl-service` 脚本用于手动管理服务：
```bash
# 启动 SSH 服务
wsl-service ssh start

# 停止 SSH 服务
wsl-service ssh stop

# 查看 SSH 服务状态
wsl-service ssh status
```

## 手动修复方法（如果脚本失败）

### 方法 1: 跳过 systemctl 错误
```bash
# 强制完成包配置
sudo dpkg --configure -a --force-depends

# 或者重新配置 openssh-server
sudo dpkg-reconfigure openssh-server
```

### 方法 2: 手动启动 SSH
```bash
# 直接启动 SSH daemon
sudo /usr/sbin/sshd -D &

# 或使用传统 service 命令
sudo service ssh start
```

### 方法 3: 启用 WSL2 systemd（推荐）
1. 确保使用 WSL2：
   ```powershell
   wsl --set-version Ubuntu 2
   ```

2. 配置 systemd：
   ```bash
   echo -e "[boot]\nsystemd=true" | sudo tee /etc/wsl.conf
   ```

3. 重启 WSL：
   ```powershell
   wsl --shutdown
   ```

## 验证修复结果

### 检查 systemd 状态
```bash
# 检查 systemd 是否运行
systemctl --version
systemctl is-system-running

# 检查服务状态
systemctl status ssh
```

### 检查 SSH 服务
```bash
# 检查 SSH 是否运行
sudo systemctl status ssh

# 或使用传统方法
sudo service ssh status

# 测试 SSH 连接
ssh localhost
```

## 常见问题解答

### Q: 为什么会出现 systemctl 错误？
A: WSL1 不支持 systemd，WSL2 需要显式启用。这个错误通常出现在包安装时尝试启动服务。

### Q: 修复后仍然有问题怎么办？
A: 
1. 确认您使用的是 WSL2
2. 检查 Windows 版本是否支持 WSL2 systemd
3. 尝试重新安装 WSL 发行版

### Q: 如何检查 WSL 版本？
A: 在 Windows PowerShell 中运行：
```powershell
wsl --list --verbose
```

### Q: 可以不使用 systemd 吗？
A: 可以，使用传统的 `service` 命令或直接运行服务程序。

## 针对您的项目的特殊配置

由于您的项目是 Edge-LLM-Infra，建议：

1. **使用 Docker**（推荐）：
   ```bash
   cd docker/scripts
   ./llm_docker_run.sh
   ./llm_docker_into.sh
   ```

2. **或者在 WSL 中构建**：
   ```bash
   # 确保依赖已安装
   sudo apt update
   sudo apt install -y libzmq3-dev libgoogle-glog-dev libboost-all-dev

   # 构建项目
   mkdir build && cd build
   cmake ..
   make -j$(nproc)
   ```

## 联系支持
如果问题仍然存在，请提供：
1. `wsl_debug.sh` 的完整输出
2. WSL 版本信息 (`wsl --version`)
3. 具体的错误消息
