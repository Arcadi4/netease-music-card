<div align="center"><img src="https://github.com/Arcadi4/netease-music-card/blob/svg/card.svg"></div>

<div align="center"><h1>neteasemusic-github-profile</h1></div>

<div align="center">🎧 在 Github Profile 显示你这周在网易云音乐上最喜欢听的歌曲 🎵</div>

## 🚀使用方法：


### 🎒 `Fork` 一份此仓库或者自己新建一个仓库

### 1. 获取网易云音乐用户 `id`

![image](https://user-images.githubusercontent.com/31311826/133114645-1a27d063-971d-4ede-9775-52f8052ef655.png)

然后修改 [main.yml](https://github.com/Nthily/netease-music-card/blob/main/.github/workflows/main.yml#L21) 中的 `USER_ID`

### 2. 获取网易云音乐用户的 `TOKEN`
 * 打开网页控制台，找到 Application 下 Cookie 为 `MUSIC_U` 的值
![}QV)3FH9@L9LUJ({35JJI}M](https://user-images.githubusercontent.com/31311826/133136019-63bbf232-d8d0-469d-8a45-f46fffdbeaab.png)
 * 打开自己项目中的设置，找到 `Secrets` 新建一个名为 `USER_TOKEN` 的 `Secrets`
 ![image](https://user-images.githubusercontent.com/31311826/133136507-fb2b61f8-1c09-40b8-bb7e-90e3f43b2c55.png)
 * 将第一步获取到的值粘贴进去

### 3. 修改 `main.yml`
 将 [main.yml](https://github.com/Nthily/netease-music-card/blob/main/.github/workflows/main.yml#L24) 中的 `AUTHOR` 修改为自己的 Github 用户名即可

### 4. 引用图片

最后只需要在你的 github profile 仓库添加图片链接即可

`![card](https://github.com/你的 Github 用户名/netease-music-card/blob/main/card.svg)`

你也可以使用 [Jsdelivr](https://www.jsdelivr.com/?docs=gh) CDN 来引用图片

`![card](https://cdn.jsdelivr.net/gh/你的 Github 用户名/netease-music-card/card.svg)`

你也可以将这个图片部署到你的博客等地方 😋

## 💨 本地测试：

`Fork` 项目或者新建一份。

需要设置以下环境变量：
* `USER_ID` - 网易云音乐用户 ID
* `USER_TOKEN` - 网易云音乐 Cookie (MUSIC_U)
* `GH_TOKEN` - GitHub Personal Access Token
* `AUTHOR` - GitHub 用户名
* `REPO` - 仓库名称

运行程序：

```bash
go run ./cmd/cardgen
```

如果你想直接在仓库的 `workflow` 里面查看详细的输出，可以不需要本地环境变量

如果你想在本地测试网易云 API 并且查看，请设置相应的环境变量，并注意在 `push` 到仓库之前删除敏感信息

## ❤️ 灵感和帮助：

[spotify-github-profile](https://github.com/kittinan/spotify-github-profile)

[netease-music-box](https://github.com/Leecason/netease-music-box)

[NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi)

## 🤔 工作原理：

* 使用 [NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi) 获取听歌记录
* 基于 Github API 将 Go 程序处理好的 `svg` 写入到仓库中
* 使用 Github Actions 定期更新 `card.svg`

## 📄 开源协议

```
MIT License

Copyright (c) 2021 Nthily

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

