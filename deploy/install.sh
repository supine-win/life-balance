#!/bin/bash
# LifeRecorder 一键安装脚本
set -e

echo "🎬 LifeRecorder 安装脚本"
echo "=========================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: Docker 未安装${NC}"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# 检查 Docker Compose
if ! docker compose version &> /dev/null && ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}错误: Docker Compose 未安装${NC}"
    echo "请先安装 Docker Compose"
    exit 1
fi

echo -e "${GREEN}✓ Docker 和 Docker Compose 已安装${NC}"

# 创建 .env 文件（如果不存在）
if [ ! -f .env ]; then
    echo ""
    echo "请配置环境变量："
    read -p "JWT 密钥 (回车使用随机值): " JWT_SECRET
    if [ -z "$JWT_SECRET" ]; then
        JWT_SECRET=$(openssl rand -hex 32)
    fi
    read -p "高德地图 API Key (可选): " AMAP_API_KEY
    read -p "LLM API URL (可选): " LLM_API_URL
    read -p "LLM API Key (可选): " LLM_API_KEY

    cat > .env << EOF
JWT_SECRET=$JWT_SECRET
AMAP_API_KEY=$AMAP_API_KEY
GEOCODING_API=${AMAP_API_KEY:+amap}
LLM_API_URL=$LLM_API_URL
LLM_API_KEY=$LLM_API_KEY
EOF
    echo -e "${GREEN}✓ .env 文件已创建${NC}"
fi

# 构建并启动
echo ""
echo -e "${YELLOW}正在构建 Docker 镜像...${NC}"
docker compose -f deploy/docker-compose.yml build

echo ""
echo -e "${YELLOW}正在启动服务...${NC}"
docker compose -f deploy/docker-compose.yml up -d

# 等待服务就绪
echo ""
echo -e "${YELLOW}等待服务就绪...${NC}"
sleep 10

# 检查服务状态
if curl -s http://localhost/api/v1/system/health | grep -q "healthy"; then
    echo ""
    echo -e "${GREEN}🎉 LifeRecorder 安装成功！${NC}"
    echo ""
    echo "访问地址: http://localhost"
    echo "管理后台: http://localhost/admin"
    echo "API 文档: http://localhost/api/v1/system/health"
    echo ""
    echo "首次使用请访问 http://localhost/setup 完成初始化"
    echo ""
    echo "常用命令："
    echo "  查看日志: docker compose -f deploy/docker-compose.yml logs -f"
    echo "  停止服务: docker compose -f deploy/docker-compose.yml down"
    echo "  重启服务: docker compose -f deploy/docker-compose.yml restart"
else
    echo -e "${YELLOW}服务可能还在启动中，请稍后访问 http://localhost${NC}"
    echo "查看日志: docker compose -f deploy/docker-compose.yml logs -f"
fi
