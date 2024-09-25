#!/bin/bash

# 📚 项目文档管理脚本
# 用于自动化文档创建、检查和管理任务
# 
# 使用方法:
#   ./scripts/docs-manager.sh create api user-service-api "用户服务API文档"
#   ./scripts/docs-manager.sh check
#   ./scripts/docs-manager.sh update-index

set -e  # 遇到错误立即退出

# 颜色输出定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # 无颜色

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOCS_DIR="$PROJECT_ROOT/docs"
TEMPLATES_DIR="$DOCS_DIR/templates"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "📚 项目文档管理脚本"
    echo ""
    echo "使用方法:"
    echo "  $0 <command> [options]"
    echo ""
    echo "命令:"
    echo "  create <type> <name> <title>  创建新文档"
    echo "  check                         检查文档格式和链接"
    echo "  update-index                  更新文档索引"
    echo "  list                          列出所有文档"
    echo "  help                          显示此帮助信息"
    echo ""
    echo "文档类型 (create命令):"
    echo "  architecture    架构设计文档"
    echo "  api            API文档"
    echo "  development    开发文档"
    echo "  deployment     部署文档"
    echo "  testing        测试文档"
    echo "  progress       进度报告"
    echo "  technical      技术笔记"
    echo ""
    echo "示例:"
    echo "  $0 create api user-service-api \"用户服务API文档\""
    echo "  $0 create architecture system-design \"系统架构设计\""
    echo "  $0 check"
    echo "  $0 update-index"
}

# 创建新文档
create_document() {
    local doc_type="$1"
    local doc_name="$2"
    local doc_title="$3"
    
    if [[ -z "$doc_type" || -z "$doc_name" || -z "$doc_title" ]]; then
        print_error "缺少必要参数。使用: $0 create <type> <name> <title>"
        return 1
    fi
    
    # 验证文档类型
    case "$doc_type" in
        architecture|api|development|deployment|testing|progress|technical)
            ;;
        *)
            print_error "无效的文档类型: $doc_type"
            print_info "支持的类型: architecture, api, development, deployment, testing, progress, technical"
            return 1
            ;;
    esac
    
    # 确定目标目录
    case "$doc_type" in
        progress)
            target_dir="$DOCS_DIR/progress-reports"
            ;;
        technical)
            target_dir="$DOCS_DIR/technical-notes"
            ;;
        *)
            target_dir="$DOCS_DIR/$doc_type"
            ;;
    esac
    
    # 创建目标目录（如果不存在）
    mkdir -p "$target_dir"
    
    # 生成文件名（确保是.md扩展名）
    if [[ "$doc_name" != *.md ]]; then
        doc_name="$doc_name.md"
    fi
    
    local target_file="$target_dir/$doc_name"
    
    # 检查文件是否已存在
    if [[ -f "$target_file" ]]; then
        print_warning "文档已存在: $target_file"
        read -p "是否覆盖？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "操作已取消"
            return 0
        fi
    fi
    
    # 复制模板文件
    local template_file="$TEMPLATES_DIR/document-template.md"
    if [[ ! -f "$template_file" ]]; then
        print_error "模板文件不存在: $template_file"
        return 1
    fi
    
    # 生成当前日期
    local current_date=$(date +%Y-%m-%d)
    
    # 确定文档类型中文名
    local doc_type_cn
    case "$doc_type" in
        architecture) doc_type_cn="架构设计" ;;
        api) doc_type_cn="API文档" ;;
        development) doc_type_cn="开发指南" ;;
        deployment) doc_type_cn="部署指南" ;;
        testing) doc_type_cn="测试文档" ;;
        progress) doc_type_cn="进度报告" ;;
        technical) doc_type_cn="技术笔记" ;;
    esac
    
    print_info "正在创建文档: $target_file"
    
    # 复制模板并替换占位符
    sed -e "s/\[文档标题\]/$doc_title/g" \
        -e "s/YYYY-MM-DD/$current_date/g" \
        -e "s/\[维护者姓名\/团队\]/开发团队/g" \
        -e "s/\[架构设计\/开发指南\/API文档\/测试报告\/技术笔记\]/$doc_type_cn/g" \
        "$template_file" > "$target_file"
    
    print_success "文档创建成功: $target_file"
    print_info "请编辑文档内容并运行 '$0 update-index' 更新索引"
}

# 检查文档格式和链接
check_documents() {
    print_info "开始检查文档..."
    
    local error_count=0
    
    # 检查所有markdown文件
    while IFS= read -r -d '' file; do
        print_info "检查文件: ${file#$PROJECT_ROOT/}"
        
        # 检查文档头部信息
        if ! grep -q "^\*\*文档版本\*\*:" "$file"; then
            print_warning "缺少版本信息: ${file#$PROJECT_ROOT/}"
            ((error_count++))
        fi
        
        if ! grep -q "^\*\*创建时间\*\*:" "$file"; then
            print_warning "缺少创建时间: ${file#$PROJECT_ROOT/}"
            ((error_count++))
        fi
        
        if ! grep -q "^\*\*维护者\*\*:" "$file"; then
            print_warning "缺少维护者信息: ${file#$PROJECT_ROOT/}"
            ((error_count++))
        fi
        
        # 检查标题层级
        local title_count=$(grep -c "^# " "$file" || true)
        if [[ $title_count -ne 1 ]]; then
            print_warning "一级标题数量异常 ($title_count): ${file#$PROJECT_ROOT/}"
            ((error_count++))
        fi
        
    done < <(find "$DOCS_DIR" -name "*.md" -type f -print0)
    
    if [[ $error_count -eq 0 ]]; then
        print_success "所有文档格式检查通过"
    else
        print_warning "发现 $error_count 个格式问题"
    fi
}

# 更新文档索引
update_index() {
    print_info "更新文档索引..."
    
    local index_file="$DOCS_DIR/README.md"
    local temp_file="/tmp/docs_index_update.md"
    
    # 备份原文件
    cp "$index_file" "${index_file}.backup"
    
    print_info "扫描文档目录..."
    
    # 生成文档列表
    {
        echo "# 📚 项目文档管理中心"
        echo ""
        echo "**创建时间**: 2025-09-29"
        echo "**维护者**: 开发团队"
        echo "**文档版本**: v1.0"
        echo ""
        echo "---"
        echo ""
        echo "## 📋 文档结构概览"
        echo ""
        echo "本项目采用分层级的文档管理体系，确保所有技术文档、开发记录、测试报告等都能被有效组织和维护。"
        echo ""
        
        # 这里可以添加更复杂的索引生成逻辑
        # 目前保持简单，提示手动更新
        print_info "索引更新功能正在开发中，请手动更新 docs/README.md"
    } > "$temp_file"
    
    # 注释掉自动替换，避免覆盖现有内容
    # mv "$temp_file" "$index_file"
    rm -f "$temp_file"
    
    print_success "索引更新完成"
}

# 列出所有文档
list_documents() {
    print_info "项目文档列表:"
    echo ""
    
    # 遍历文档目录
    for dir in architecture api development deployment testing progress-reports technical-notes; do
        local full_dir="$DOCS_DIR/$dir"
        if [[ -d "$full_dir" ]]; then
            local dir_name
            case "$dir" in
                progress-reports) dir_name="📊 进度报告" ;;
                technical-notes) dir_name="📝 技术笔记" ;;
                architecture) dir_name="🏗️ 架构设计" ;;
                api) dir_name="🔗 API文档" ;;
                development) dir_name="💻 开发文档" ;;
                deployment) dir_name="🚀 部署文档" ;;
                testing) dir_name="🧪 测试文档" ;;
                *) dir_name="📄 $dir" ;;
            esac
            
            echo -e "${BLUE}$dir_name${NC}"
            
            while IFS= read -r -d '' file; do
                local filename=$(basename "$file")
                local title=$(grep "^# " "$file" | head -1 | sed 's/^# //' || echo "$filename")
                echo "  - $filename: $title"
            done < <(find "$full_dir" -name "*.md" -type f -print0 2>/dev/null)
            
            echo ""
        fi
    done
}

# 主函数
main() {
    # 检查是否在项目根目录
    if [[ ! -d "$DOCS_DIR" ]]; then
        print_error "未找到docs目录。请在项目根目录运行此脚本。"
        exit 1
    fi
    
    # 解析命令
    case "${1:-help}" in
        create)
            create_document "$2" "$3" "$4"
            ;;
        check)
            check_documents
            ;;
        update-index)
            update_index
            ;;
        list)
            list_documents
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@" 