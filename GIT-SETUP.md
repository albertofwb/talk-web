# Git 仓库设置和推送指南

## 当前状态 ✅

- ✅ 代码已提交到本地 Git 仓库
- ✅ 提交信息完整
- ✅ 39 个文件，5716 行代码
- ⏳ 等待推送到远程仓库

## 推送到 GitHub

### 方式一：新建 GitHub 仓库（推荐）

#### 1. 在 GitHub 上创建新仓库

1. 访问 https://github.com/new
2. 输入仓库名：`talk-web`
3. 描述：`Voice chat web application with push-to-talk and STT`
4. 选择 **Public** 或 **Private**
5. **不要**勾选 "Initialize this repository with README"
6. 点击 "Create repository"

#### 2. 推送代码到 GitHub

```bash
# 添加远程仓库（替换为你的 GitHub 用户名）
git remote add origin https://github.com/YOUR_USERNAME/talk-web.git

# 或使用 SSH（如果已配置 SSH key）
git remote add origin git@github.com:YOUR_USERNAME/talk-web.git

# 推送代码
git push -u origin main
```

#### 3. 验证推送

访问你的 GitHub 仓库，应该能看到所有代码和 README。

### 方式二：使用 GitHub CLI（gh）

```bash
# 安装 gh（如果未安装）
# Ubuntu/Debian: sudo apt install gh
# macOS: brew install gh

# 登录 GitHub
gh auth login

# 创建仓库并推送
gh repo create talk-web --public --source=. --push

# 或创建私有仓库
gh repo create talk-web --private --source=. --push
```

### 方式三：推送到现有仓库

如果已有仓库：

```bash
# 添加远程仓库
git remote add origin <your-repo-url>

# 推送代码
git push -u origin main
```

## 推送到 GitLab

```bash
# 1. 在 GitLab 创建新项目
# 访问: https://gitlab.com/projects/new

# 2. 添加远程仓库
git remote add origin https://gitlab.com/YOUR_USERNAME/talk-web.git

# 3. 推送代码
git push -u origin main
```

## 推送到 Gitee（码云）

```bash
# 1. 在 Gitee 创建新仓库
# 访问: https://gitee.com/projects/new

# 2. 添加远程仓库
git remote add origin https://gitee.com/YOUR_USERNAME/talk-web.git

# 3. 推送代码
git push -u origin main
```

## 推送到自己的 Git 服务器

```bash
# 添加远程仓库
git remote add origin ssh://user@your-server.com/path/to/talk-web.git

# 推送代码
git push -u origin main
```

## 验证推送

```bash
# 查看远程仓库信息
git remote -v

# 查看提交历史
git log --oneline

# 查看仓库状态
git status
```

## 常见问题

### Q1: 推送时要求输入用户名密码

**A:** GitHub 已不再支持密码认证，需要使用以下方式之一：

1. **Personal Access Token (推荐)**
   ```bash
   # 1. 生成 Token: https://github.com/settings/tokens
   # 2. 推送时使用 Token 作为密码
   git push -u origin main
   # Username: your-username
   # Password: ghp_xxxxxxxxxxxxx (你的 token)
   ```

2. **SSH Key**
   ```bash
   # 生成 SSH key
   ssh-keygen -t ed25519 -C "your-email@example.com"

   # 复制公钥
   cat ~/.ssh/id_ed25519.pub

   # 添加到 GitHub: https://github.com/settings/keys

   # 使用 SSH URL
   git remote set-url origin git@github.com:YOUR_USERNAME/talk-web.git
   ```

### Q2: 推送被拒绝 (rejected)

```bash
# 如果远程有你本地没有的提交
git pull --rebase origin main
git push -u origin main
```

### Q3: 推送失败 - 权限不足

确保：
- GitHub 仓库存在
- 你有写入权限
- 使用正确的认证方式

### Q4: 想要修改远程仓库地址

```bash
# 查看当前远程仓库
git remote -v

# 修改远程仓库地址
git remote set-url origin <new-url>
```

## 后续开发流程

### 日常提交

```bash
# 1. 查看修改
git status

# 2. 添加修改的文件
git add .

# 3. 提交
git commit -m "feat: add new feature"

# 4. 推送
git push
```

### 分支开发

```bash
# 创建新分支
git checkout -b feature/new-feature

# 开发完成后推送
git push -u origin feature/new-feature

# 在 GitHub 上创建 Pull Request
```

### 拉取更新

```bash
# 拉取最新代码
git pull

# 或者
git fetch origin
git merge origin/main
```

## 推送检查清单

推送前确认：

- [ ] 代码已测试通过
- [ ] 没有敏感信息（密码、密钥等）
- [ ] .gitignore 配置正确
- [ ] README 文档完整
- [ ] 提交信息清晰

## 快速命令参考

```bash
# 查看状态
git status

# 添加所有文件
git add .

# 提交
git commit -m "your message"

# 推送
git push

# 拉取
git pull

# 查看日志
git log --oneline

# 查看远程仓库
git remote -v
```

## 需要帮助？

- GitHub 文档: https://docs.github.com
- Git 文档: https://git-scm.com/doc
- Pro Git 书籍: https://git-scm.com/book/zh/v2

---

准备好推送了吗？选择上面的任一方式开始吧！ 🚀
