#!/bin/bash

# Linux后端开发环境设置脚本
# VideoCall System - Linux Backend Development Setup

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 打印函数
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_info() { echo -e "${CYAN}ℹ️  $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }
print_header() { echo -e "${PURPLE}🎯 $1${NC}"; }

# 检测Linux发行版
detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
        VERSION=$VERSION_ID
    else
        print_error "无法检测Linux发行版"
        exit 1
    fi
    
    print_info "检测到系统: $PRETTY_NAME"
}

# 更新包管理器
update_package_manager() {
    print_info "更新包管理器..."
    
    case $DISTRO in
        ubuntu|debian)
            sudo apt-get update
            sudo apt-get upgrade -y
            ;;
        centos|rhel|fedora)
            if command -v dnf &> /dev/null; then
                sudo dnf update -y
            else
                sudo yum update -y
            fi
            ;;
        arch|manjaro)
            sudo pacman -Syu --noconfirm
            ;;
        *)
            print_warning "未知的Linux发行版，请手动更新包管理器"
            ;;
    esac
    
    print_success "包管理器更新完成"
}

# 安装基础开发工具
install_basic_tools() {
    print_info "安装基础开发工具..."
    
    case $DISTRO in
        ubuntu|debian)
            sudo apt-get install -y \
                build-essential \
                cmake \
                ninja-build \
                git \
                curl \
                wget \
                unzip \
                pkg-config \
                ca-certificates \
                gnupg \
                lsb-release
            ;;
        centos|rhel|fedora)
            if command -v dnf &> /dev/null; then
                sudo dnf groupinstall -y "Development Tools"
                sudo dnf install -y cmake ninja-build git curl wget unzip pkgconfig
            else
                sudo yum groupinstall -y "Development Tools"
                sudo yum install -y cmake3 ninja-build git curl wget unzip pkgconfig
                # 创建cmake符号链接
                sudo ln -sf /usr/bin/cmake3 /usr/bin/cmake
            fi
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm \
                base-devel \
                cmake \
                ninja \
                git \
                curl \
                wget \
                unzip \
                pkgconf
            ;;
    esac
    
    print_success "基础开发工具安装完成"
}

# 安装Go语言环境
install_go() {
    print_info "安装Go语言环境..."
    
    # 检查是否已安装
    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go已安装，版本: $go_version"
        return 0
    fi
    
    # 下载并安装最新版Go
    local go_version="1.21.5"
    local go_archive="go${go_version}.linux-amd64.tar.gz"
    local go_url="https://golang.org/dl/${go_archive}"
    
    print_info "下载Go ${go_version}..."
    wget -O "/tmp/${go_archive}" "$go_url"
    
    # 删除旧版本并安装新版本
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "/tmp/${go_archive}"
    rm "/tmp/${go_archive}"
    
    # 设置环境变量
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc
    fi
    
    # 立即生效
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    export GOBIN=$GOPATH/bin
    
    # 创建Go工作目录
    mkdir -p $GOPATH/{bin,src,pkg}
    
    print_success "Go语言环境安装完成"
    go version
}

# 安装Node.js和npm
install_nodejs() {
    print_info "安装Node.js和npm..."
    
    # 使用NodeSource仓库安装最新LTS版本
    curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
    
    case $DISTRO in
        ubuntu|debian)
            sudo apt-get install -y nodejs
            ;;
        centos|rhel|fedora)
            if command -v dnf &> /dev/null; then
                sudo dnf install -y nodejs npm
            else
                sudo yum install -y nodejs npm
            fi
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm nodejs npm
            ;;
    esac
    
    # 安装常用全局包
    sudo npm install -g yarn pnpm pm2
    
    print_success "Node.js环境安装完成"
    node --version
    npm --version
}

# 安装Python开发环境
install_python() {
    print_info "安装Python开发环境..."
    
    case $DISTRO in
        ubuntu|debian)
            sudo apt-get install -y \
                python3 \
                python3-pip \
                python3-venv \
                python3-dev \
                python3-setuptools
            ;;
        centos|rhel|fedora)
            if command -v dnf &> /dev/null; then
                sudo dnf install -y python3 python3-pip python3-devel
            else
                sudo yum install -y python3 python3-pip python3-devel
            fi
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm python python-pip
            ;;
    esac
    
    # 升级pip
    python3 -m pip install --upgrade pip
    
    # 安装常用Python包
    python3 -m pip install --user \
        virtualenv \
        pipenv \
        poetry \
        black \
        flake8 \
        pytest
    
    print_success "Python开发环境安装完成"
    python3 --version
    pip3 --version
}

# 安装数据库
install_databases() {
    print_info "安装数据库..."
    
    case $DISTRO in
        ubuntu|debian)
            # PostgreSQL
            sudo apt-get install -y postgresql postgresql-contrib postgresql-client
            
            # Redis
            sudo apt-get install -y redis-server
            
            # MongoDB
            wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -
            echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu $(lsb_release -cs)/mongodb-org/6.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-6.0.list
            sudo apt-get update
            sudo apt-get install -y mongodb-org
            ;;
        centos|rhel|fedora)
            if command -v dnf &> /dev/null; then
                sudo dnf install -y postgresql postgresql-server postgresql-contrib redis mongodb-org
            else
                sudo yum install -y postgresql postgresql-server postgresql-contrib redis mongodb-org
            fi
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm postgresql redis mongodb
            ;;
    esac
    
    # 启动服务
    sudo systemctl enable postgresql redis-server mongod
    sudo systemctl start postgresql redis-server mongod
    
    print_success "数据库安装完成"
}

# 安装Docker
install_docker() {
    print_info "安装Docker..."
    
    # 卸载旧版本
    case $DISTRO in
        ubuntu|debian)
            sudo apt-get remove -y docker docker-engine docker.io containerd runc || true
            
            # 添加Docker官方GPG密钥
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
            
            # 添加Docker仓库
            echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
            
            sudo apt-get update
            sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
            ;;
        centos|rhel|fedora)
            sudo yum remove -y docker docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate docker-engine || true
            
            sudo yum install -y yum-utils
            sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm docker docker-compose
            ;;
    esac
    
    # 启动Docker服务
    sudo systemctl enable docker
    sudo systemctl start docker
    
    # 添加用户到docker组
    sudo usermod -aG docker $USER
    
    print_success "Docker安装完成"
    print_warning "请重新登录以使docker组权限生效"
}

# 安装开发工具
install_dev_tools() {
    print_info "安装开发工具..."
    
    case $DISTRO in
        ubuntu|debian)
            sudo apt-get install -y \
                vim \
                neovim \
                tmux \
                htop \
                tree \
                jq \
                httpie \
                net-tools \
                telnet \
                nc \
                lsof
            ;;
        centos|rhel|fedora)
            if command -v dnf &> /dev/null; then
                sudo dnf install -y vim neovim tmux htop tree jq httpie net-tools telnet nc lsof
            else
                sudo yum install -y vim neovim tmux htop tree jq httpie net-tools telnet nc lsof
            fi
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm vim neovim tmux htop tree jq httpie net-tools gnu-netcat lsof
            ;;
    esac
    
    print_success "开发工具安装完成"
}

# 配置防火墙
configure_firewall() {
    print_info "配置防火墙..."
    
    # 开放常用端口
    local ports=(22 80 443 8080 8081 8082 8083 8084 8085 8086 8087 5432 6379 27017)
    
    if command -v ufw &> /dev/null; then
        # Ubuntu/Debian使用ufw
        sudo ufw --force enable
        for port in "${ports[@]}"; do
            sudo ufw allow $port
        done
    elif command -v firewall-cmd &> /dev/null; then
        # CentOS/RHEL/Fedora使用firewalld
        sudo systemctl enable firewalld
        sudo systemctl start firewalld
        for port in "${ports[@]}"; do
            sudo firewall-cmd --permanent --add-port=${port}/tcp
        done
        sudo firewall-cmd --reload
    fi
    
    print_success "防火墙配置完成"
}

# 创建项目目录结构
create_project_structure() {
    print_info "创建项目目录结构..."
    
    local project_root="/opt/videocall-system"
    sudo mkdir -p $project_root/{logs,data,config,scripts,backups}
    sudo chown -R $USER:$USER $project_root
    
    # 创建systemd服务目录
    sudo mkdir -p /etc/systemd/system
    
    # 创建日志轮转配置
    sudo tee /etc/logrotate.d/videocall-system > /dev/null <<EOF
$project_root/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 $USER $USER
    postrotate
        systemctl reload videocall-* || true
    endscript
}
EOF
    
    print_success "项目目录结构创建完成"
}

# 创建构建脚本
create_build_scripts() {
    print_info "创建构建脚本..."
    
    # 后端构建脚本
    cat > build-backend.sh << 'EOF'
#!/bin/bash

set -e

BUILD_TYPE=${1:-release}
CLEAN=${2:-false}

echo "🚀 VideoCall System - Linux Backend Build Script"

# 设置Go环境
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:/usr/local/go/bin:$GOBIN

# 构建目录
BUILD_DIR="build-linux"

if [ "$CLEAN" = "true" ] && [ -d "$BUILD_DIR" ]; then
    echo "🧹 清理构建目录..."
    rm -rf $BUILD_DIR
fi

mkdir -p $BUILD_DIR
cd $BUILD_DIR

echo "📋 构建Go后端服务..."

# 构建各个微服务
services=(
    "user-service"
    "meeting-service"
    "signaling-service"
    "media-service"
    "ai-detection-service"
    "notification-service"
    "file-service"
    "gateway-service"
)

for service in "${services[@]}"; do
    echo "🔨 构建 $service..."
    
    if [ -d "../src/backend/services/$service" ]; then
        cd "../src/backend/services/$service"
        
        # 下载依赖
        go mod download
        go mod tidy
        
        # 构建
        if [ "$BUILD_TYPE" = "debug" ]; then
            go build -race -o "../../../../$BUILD_DIR/$service" .
        else
            go build -ldflags="-s -w" -o "../../../../$BUILD_DIR/$service" .
        fi
        
        cd "../../../../$BUILD_DIR"
        echo "✅ $service 构建完成"
    else
        echo "⚠️  $service 目录不存在，跳过"
    fi
done

echo "📦 构建AI检测服务..."
if [ -d "../src/ai-detection" ]; then
    cd "../src/ai-detection"
    
    # 创建虚拟环境
    python3 -m venv venv
    source venv/bin/activate
    
    # 安装依赖
    pip install -r requirements.txt
    
    # 复制到构建目录
    cp -r . "../../$BUILD_DIR/ai-detection/"
    
    cd "../../$BUILD_DIR"
    echo "✅ AI检测服务构建完成"
fi

echo "🎉 所有服务构建完成！"
echo "📁 输出目录: $(pwd)"

# 创建启动脚本
cat > start-all-services.sh << 'SCRIPT_EOF'
#!/bin/bash

echo "🚀 启动VideoCall System所有服务..."

# 启动数据库服务
sudo systemctl start postgresql redis-server mongod

# 启动后端服务
for service in user-service meeting-service signaling-service media-service ai-detection-service notification-service file-service; do
    if [ -f "./$service" ]; then
        echo "启动 $service..."
        nohup ./$service > logs/$service.log 2>&1 &
        echo $! > pids/$service.pid
    fi
done

# 启动网关服务（最后启动）
if [ -f "./gateway-service" ]; then
    echo "启动 gateway-service..."
    nohup ./gateway-service > logs/gateway-service.log 2>&1 &
    echo $! > pids/gateway-service.pid
fi

# 启动AI检测服务
if [ -d "./ai-detection" ]; then
    echo "启动 AI检测服务..."
    cd ai-detection
    source venv/bin/activate
    nohup python app.py > ../logs/ai-detection.log 2>&1 &
    echo $! > ../pids/ai-detection.pid
    cd ..
fi

echo "✅ 所有服务启动完成！"
echo "📊 查看状态: ./status.sh"
SCRIPT_EOF

chmod +x start-all-services.sh

# 创建状态检查脚本
cat > status.sh << 'STATUS_EOF'
#!/bin/bash

echo "📊 VideoCall System 服务状态"
echo "================================"

# 检查数据库服务
echo "🗄️  数据库服务:"
systemctl is-active postgresql redis-server mongod | while read status; do
    if [ "$status" = "active" ]; then
        echo "  ✅ 数据库服务运行正常"
    else
        echo "  ❌ 数据库服务异常"
    fi
done

# 检查后端服务
echo ""
echo "⚙️  后端服务:"
for pidfile in pids/*.pid; do
    if [ -f "$pidfile" ]; then
        service_name=$(basename "$pidfile" .pid)
        pid=$(cat "$pidfile")
        if kill -0 "$pid" 2>/dev/null; then
            echo "  ✅ $service_name (PID: $pid)"
        else
            echo "  ❌ $service_name (已停止)"
        fi
    fi
done

# 检查端口占用
echo ""
echo "🌐 端口占用:"
ports=(8080 8081 8082 8083 8084 8085 8086 8087 5432 6379 27017)
for port in "${ports[@]}"; do
    if netstat -tuln | grep -q ":$port "; then
        echo "  ✅ 端口 $port 已占用"
    else
        echo "  ⚠️  端口 $port 未占用"
    fi
done
STATUS_EOF

chmod +x status.sh

# 创建停止脚本
cat > stop-all-services.sh << 'STOP_EOF'
#!/bin/bash

echo "🛑 停止VideoCall System所有服务..."

# 停止后端服务
for pidfile in pids/*.pid; do
    if [ -f "$pidfile" ]; then
        service_name=$(basename "$pidfile" .pid)
        pid=$(cat "$pidfile")
        if kill -0 "$pid" 2>/dev/null; then
            echo "停止 $service_name (PID: $pid)..."
            kill "$pid"
            rm "$pidfile"
        fi
    fi
done

echo "✅ 所有服务已停止"
STOP_EOF

chmod +x stop-all-services.sh

# 创建必要目录
mkdir -p logs pids

EOF

    chmod +x build-backend.sh
    
    print_success "构建脚本创建完成"
}

# 主函数
main() {
    print_header "VideoCall System - Linux后端开发环境设置"
    echo "================================================"
    
    detect_distro
    
    # 检查参数
    INSTALL_ALL=false
    INSTALL_BASIC=false
    INSTALL_LANGUAGES=false
    INSTALL_DATABASES=false
    INSTALL_DOCKER=false
    INSTALL_TOOLS=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --all)
                INSTALL_ALL=true
                shift
                ;;
            --basic)
                INSTALL_BASIC=true
                shift
                ;;
            --languages)
                INSTALL_LANGUAGES=true
                shift
                ;;
            --databases)
                INSTALL_DATABASES=true
                shift
                ;;
            --docker)
                INSTALL_DOCKER=true
                shift
                ;;
            --tools)
                INSTALL_TOOLS=true
                shift
                ;;
            --help)
                echo "用法: $0 [选项]"
                echo ""
                echo "选项:"
                echo "  --all         安装所有组件"
                echo "  --basic       安装基础开发工具"
                echo "  --languages   安装编程语言环境"
                echo "  --databases   安装数据库"
                echo "  --docker      安装Docker"
                echo "  --tools       安装开发工具"
                echo "  --help        显示此帮助"
                exit 0
                ;;
            *)
                print_error "未知选项: $1"
                exit 1
                ;;
        esac
    done
    
    if [ "$INSTALL_ALL" = true ]; then
        INSTALL_BASIC=true
        INSTALL_LANGUAGES=true
        INSTALL_DATABASES=true
        INSTALL_DOCKER=true
        INSTALL_TOOLS=true
    fi
    
    # 如果没有指定任何选项，默认安装所有
    if [ "$INSTALL_BASIC" = false ] && [ "$INSTALL_LANGUAGES" = false ] && [ "$INSTALL_DATABASES" = false ] && [ "$INSTALL_DOCKER" = false ] && [ "$INSTALL_TOOLS" = false ]; then
        INSTALL_ALL=true
        INSTALL_BASIC=true
        INSTALL_LANGUAGES=true
        INSTALL_DATABASES=true
        INSTALL_DOCKER=true
        INSTALL_TOOLS=true
    fi
    
    update_package_manager
    
    if [ "$INSTALL_BASIC" = true ]; then
        install_basic_tools
    fi
    
    if [ "$INSTALL_LANGUAGES" = true ]; then
        install_go
        install_nodejs
        install_python
    fi
    
    if [ "$INSTALL_DATABASES" = true ]; then
        install_databases
    fi
    
    if [ "$INSTALL_DOCKER" = true ]; then
        install_docker
    fi
    
    if [ "$INSTALL_TOOLS" = true ]; then
        install_dev_tools
    fi
    
    configure_firewall
    create_project_structure
    create_build_scripts
    
    print_success "Linux后端开发环境设置完成！"
    print_info "请运行 'source ~/.bashrc' 或重新登录以使环境变量生效"
    print_info "然后运行 './build-backend.sh' 来构建后端服务"
}

main "$@"
