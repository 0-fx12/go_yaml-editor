#!/bin/bash

# YAML配置管理器Kubernetes部署脚本
# 这个脚本用于在Kubernetes集群中部署完整的应用栈

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查kubectl是否可用
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found. Please install kubectl first."
        exit 1
    fi

    if ! kubectl cluster-info &> /dev/null; then
        log_error "Unable to connect to Kubernetes cluster."
        exit 1
    fi

    log_success "kubectl is available and connected to cluster"
}

# 检查Docker是否可用
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Please install Docker first."
        exit 1
    fi
    log_success "Docker is available"
}

# 构建Docker镜像
build_image() {
    local image_name="yaml-config-manager"
    local tag=${1:-"latest"}

    log_info "Building Docker image: ${image_name}:${tag}"
    docker build -t "${image_name}:${tag}" .

    log_info "Pushing image to registry (if needed)"
    # 如果你使用私有registry，需要取消注释下面的行
    # docker tag "${image_name}:${tag}" "your-registry.com/${image_name}:${tag}"
    # docker push "your-registry.com/${image_name}:${tag}"

    log_success "Docker image built successfully"
}

# 部署到Kubernetes
deploy_to_k8s() {
    log_info "Starting deployment to Kubernetes..."

    # 创建命名空间
    log_info "Creating namespace..."
    kubectl apply -f k8s/k8s-namespace.yaml

    # 部署MongoDB
    log_info "Deploying MongoDB..."
    kubectl apply -f k8s/k8s-mongodb.yaml

    # 等待MongoDB就绪
    log_info "Waiting for MongoDB to be ready..."
    kubectl wait --for=condition=ready pod -l app=mongodb -n yaml-config-manager --timeout=300s

    # 创建ConfigMap
    log_info "Creating ConfigMaps..."
    kubectl apply -f k8s/k8s-configmap.yaml

    # 创建额外的配置文件ConfigMap（如果需要）
    if [ -f "sample_config.yaml" ]; then
        kubectl create configmap app-config-files \
            --from-file=sample_config.yaml \
            -n yaml-config-manager --dry-run=client -o yaml | kubectl apply -f -
    fi

    # 部署应用
    log_info "Deploying application..."
    kubectl apply -f k8s/k8s-app-deployment.yaml

    # 等待应用就绪
    log_info "Waiting for application to be ready..."
    kubectl wait --for=condition=ready pod -l app=yaml-config-manager -n yaml-config-manager --timeout=300s

    # 部署Ingress
    log_info "Deploying Ingress..."
    kubectl apply -f k8s/k8s-ingress.yaml

    log_success "Deployment completed successfully!"
}

# 获取部署状态
get_status() {
    log_info "Getting deployment status..."

    echo "=== Pods Status ==="
    kubectl get pods -n yaml-config-manager

    echo ""
    echo "=== Services Status ==="
    kubectl get svc -n yaml-config-manager

    echo ""
    echo "=== Ingress Status ==="
    kubectl get ingress -n yaml-config-manager

    echo ""
    echo "=== Pod Logs (App) ==="
    kubectl logs -l app=yaml-config-manager -n yaml-config-manager --tail=20
}

# 清理部署
cleanup() {
    log_warning "This will delete all resources in the yaml-config-manager namespace"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleaning up deployment..."
        kubectl delete namespace yaml-config-manager --ignore-not-found=true
        log_success "Cleanup completed"
    else
        log_info "Cleanup cancelled"
    fi
}

# 显示帮助信息
show_help() {
    echo "YAML配置管理器 - Kubernetes部署脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -b, --build          构建Docker镜像"
    echo "  -d, --deploy         部署到Kubernetes"
    echo "  -s, --status         查看部署状态"
    echo "  -c, --cleanup        清理部署"
    echo "  -h, --help           显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 --build --deploy    # 构建并部署"
    echo "  $0 --status           # 查看状态"
    echo "  $0 --cleanup          # 清理部署"
}

# 主函数
main() {
    local build=false
    local deploy=false
    local status=false
    local cleanup=false

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -b|--build)
                build=true
                shift
                ;;
            -d|--deploy)
                deploy=true
                shift
                ;;
            -s|--status)
                status=true
                shift
                ;;
            -c|--cleanup)
                cleanup=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 如果没有参数，显示帮助
    if [[ "$build" == "false" && "$deploy" == "false" && "$status" == "false" && "$cleanup" == "false" ]]; then
        show_help
        exit 0
    fi

    # 检查依赖
    check_kubectl

    # 执行操作
    if [[ "$build" == "true" ]]; then
        check_docker
        build_image
    fi

    if [[ "$deploy" == "true" ]]; then
        deploy_to_k8s
    fi

    if [[ "$status" == "true" ]]; then
        get_status
    fi

    if [[ "$cleanup" == "true" ]]; then
        cleanup
    fi
}

# 运行主函数
main "$@"
