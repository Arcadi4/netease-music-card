<div align="center">

<table>
 <td>
  <image height=200 src="https://github.com/Arcadi4/netease-music-card/blob/svg/card.svg">
 </td>
 <td>
  <image height=200 src="https://github.com/Arcadi4/netease-music-card/blob/svg/top-artists.svg">
 </td>
 <td>
  <image height=200 src="https://github.com/Arcadi4/netease-music-card/blob/svg/top-tracks.svg">
 </td>
   <td>
  <image height=200 src="https://github.com/Arcadi4/netease-music-card/blob/svg/weekly-overview.svg">
 </td>
</table>

<h1>Netease Music Cards</h1>

🎧 在 Github Profile 显示你这周在网易云音乐上最喜欢听的歌曲 🎵

</div>

## 🚀 使用方法

### 1. Fork 本仓库

注意只 fork `main` 分支。

### 2. 获取网易云音乐用户ID

进入网页端网易云音乐用户主页，URL 中 `https://music.163.com/#/user/home?id=123456789` 的 `123456789` 就是用户 ID 了。

<image width=360 src="https://user-images.githubusercontent.com/31311826/133114645-1a27d063-971d-4ede-9775-52f8052ef655.png">

然后在你 fork 的仓库中打开 `Settings > Security > Secrets and Variables > Actions`

<image width=360 src="assets/image/settings.png">

添加一个名为 `NETEASE_USER_ID` 的 `Variable`，值为你刚才获取到的用户 ID。

<image width=360 src="assets/image/variable.png">

### 3. 获取网易云音乐用户TOKEN (Cookie)

访问 [网易云音乐](https://music.163.com/) 网页端，登录你的账号。

打开网页控制台（通常为`f12`），找到 Application 下 Cookie 为 `MUSIC_U` 的值:

<image width=360 src="https://user-images.githubusercontent.com/31311826/133136019-63bbf232-d8d0-469d-8a45-f46fffdbeaab.png"/>


回到 `Secrets and Variables` 页面，添加一个名为 `NETEASE_USER_TOKEN` 的 `Secret`，值为你刚才获取到的 `MUSIC_U` 的值。

<image width=360 src="assets/image/secret.png">

### 4. 引用图片

图片会生成在 `.github/workflows/main.yml` 中指定的分支（默认为 `svg`）中。

最后只需要在你的 github profile 仓库添加图片链接即可。

`![card](https://github.com/GitHub用户名/netease-music-card/blob/svg/card.svg)`

你也可以将这个图片部署到你的博客等地方 😋

## 💨 本地测试

```bash
git clone https://github.com/Arcadi4/netease-music-card.git
```

需要设置以下环境变量：

- `NETEASE_USER_ID` - 网易云音乐用户 ID
- `NETEASE_USER_TOKEN` - 网易云音乐 Cookie (MUSIC_U)
- `OUTPUT_BRANCH` - 输出分支，默认为 `svg`（可选）
- `GITHUB_TOKEN` - GitHub Token（可选，不设置则跳过发布）

运行程序：

```bash
go run ./cmd/cardgen
```

## ❤️ 灵感和帮助

本项目基于 [12og3r/netease-music-card](https://github.com/12og3r/netease-music-card) 以go重构，并添加更多样式。

## 🤔 工作原理

- 抓取网易云音乐用户数据，使用 Go 模板生成 `svg` 图片
- 使用 GitHub Actions 定时更新图片
