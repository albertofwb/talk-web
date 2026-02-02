#!/bin/bash

echo "🚀 talk-web - 推送到 GitHub"
echo "============================"
echo ""

# 检查是否已配置远程仓库
if git remote | grep -q origin; then
    echo "✓ 已配置远程仓库:"
    git remote -v
    echo ""
    read -p "是否要推送到此仓库? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "取消推送"
        exit 0
    fi
else
    echo "⚠️  尚未配置远程仓库"
    echo ""
    echo "请选择操作:"
    echo "1) 我已在 GitHub 创建仓库，输入仓库 URL"
    echo "2) 使用 GitHub CLI (gh) 自动创建并推送"
    echo "3) 取消，稍后手动配置"
    echo ""
    read -p "请选择 (1/2/3): " -n 1 -r choice
    echo ""

    case $choice in
        1)
            echo ""
            echo "请输入 GitHub 仓库 URL:"
            echo "HTTPS 示例: https://github.com/username/talk-web.git"
            echo "SSH 示例: git@github.com:username/talk-web.git"
            echo ""
            read -p "URL: " repo_url

            if [ -z "$repo_url" ]; then
                echo "❌ URL 不能为空"
                exit 1
            fi

            git remote add origin "$repo_url"
            echo "✓ 已添加远程仓库: $repo_url"
            ;;
        2)
            if ! command -v gh &> /dev/null; then
                echo "❌ 未安装 GitHub CLI (gh)"
                echo ""
                echo "安装方法:"
                echo "  Ubuntu/Debian: sudo apt install gh"
                echo "  macOS: brew install gh"
                echo ""
                exit 1
            fi

            echo ""
            echo "请选择仓库类型:"
            read -p "创建公开仓库? (Y/n): " -n 1 -r
            echo ""

            if [[ $REPLY =~ ^[Nn]$ ]]; then
                visibility="--private"
                echo "创建私有仓库..."
            else
                visibility="--public"
                echo "创建公开仓库..."
            fi

            gh repo create talk-web $visibility --source=. --push

            if [ $? -eq 0 ]; then
                echo ""
                echo "✅ 仓库创建并推送成功！"
                gh repo view --web
                exit 0
            else
                echo "❌ 创建失败，请检查 gh 是否已登录"
                echo "运行: gh auth login"
                exit 1
            fi
            ;;
        3)
            echo "取消操作"
            echo ""
            echo "手动配置方法:"
            echo "  git remote add origin <your-repo-url>"
            echo "  git push -u origin main"
            exit 0
            ;;
        *)
            echo "无效选择"
            exit 1
            ;;
    esac
fi

# 推送到远程仓库
echo ""
echo "开始推送..."
echo ""

git push -u origin main

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ 推送成功！"
    echo ""
    echo "仓库信息:"
    git remote -v
    echo ""
    echo "最新提交:"
    git log --oneline -3
else
    echo ""
    echo "❌ 推送失败"
    echo ""
    echo "可能的原因:"
    echo "1. 需要认证 - 使用 Personal Access Token 或 SSH Key"
    echo "2. 远程仓库不存在"
    echo "3. 没有推送权限"
    echo ""
    echo "详细说明请查看: GIT-SETUP.md"
fi
