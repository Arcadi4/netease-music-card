# 网易云 WEAPI 逆向文档（基于本仓库实测）

## 1. 文档范围

本文件只覆盖**本项目实际调用并已实测**的开放接口（3 个）：

1. 用户账户信息 `user_account`
2. 用户播放记录 `user_record`（周榜 + 全时段）
3. 歌曲详情 `song_detail`

说明：

- 这些接口由 `NeteaseCloudMusicApi` 封装调用，项目入口见 `index.js`。
- 本文的“可靠性”来自仓库内真实联调抓包结果，不是静态推测。

## 2. 证据来源与样本

- 调用代码：`index.js`
- SDK 封装定义：
  - `node_modules/NeteaseCloudMusicApi/module/user_account.js`
  - `node_modules/NeteaseCloudMusicApi/module/user_record.js`
  - `node_modules/NeteaseCloudMusicApi/module/song_detail.js`
- WEAPI 路径改写/加密逻辑：`node_modules/NeteaseCloudMusicApi/util/request.js`
- 抓包产物：`artifacts/api-request-captures/`
- 可靠性统计：`artifacts/api-request-captures/reliability-summary.json`

当前样本：

- 采样轮次：6
- 总请求数：24
- 成功数：24
- 总成功率：100%

## 3. 协议与鉴权

### 3.1 鉴权方式

- 鉴权 Cookie：`MUSIC_U=<token>`
- 项目中由环境变量注入：`USER_TOKEN`（兼容 `NETEASE_USER_TOKEN`）

### 3.2 请求协议

- 协议：HTTPS
- 方法：`POST`
- `Content-Type`：`application/x-www-form-urlencoded`
- 发送体字段：
  - `params`（加密后的业务参数）
  - `encSecKey`（加密密钥）

### 3.3 路径改写（关键）

在 `request.js` 中，`crypto: 'weapi'` 会执行：

- 业务参数加密：`encrypt.weapi(data)`
- URL 改写：`url = url.replace(/\w*api/, 'weapi')`

因此：

- SDK 中声明的 `/api/...` 端点，实发请求会变成 `/weapi/...`。

## 4. 开放接口清单（本项目实测）

| 接口标识 | SDK 方法       | SDK 声明 URL                                  | 实发 URL（抓包）                                | 请求方法 | 可靠性（6轮样本）                           |
|----------|----------------|-----------------------------------------------|-------------------------------------------------|----------|---------------------------------------------|
| 账户信息 | `user_account` | `https://music.163.com/api/nuser/account/get` | `https://music.163.com/weapi/nuser/account/get` | `POST`   | 6/6 成功，100%，avg 955.5ms，p95 1174ms     |
| 播放记录 | `user_record`  | `https://music.163.com/weapi/v1/play/record`  | `https://music.163.com/weapi/v1/play/record`    | `POST`   | 12/12 成功，100%，avg 2246.75ms，p95 2450ms |
| 歌曲详情 | `song_detail`  | `https://music.163.com/api/v3/song/detail`    | `https://music.163.com/weapi/v3/song/detail`    | `POST`   | 6/6 成功，100%，avg 836.33ms，p95 859ms     |

> 备注：`user_record` 同一个端点按 `type` 返回不同结构（`weekData` / `allData`）。

## 5. 数据格式（Schema）

以下 schema 为抓包实测结构（字段省略非核心部分）。

### 5.1 `user_account`

```json
{
  "code": 200,
  "account": {
    "id": 1696843520,
    "userName": "string",
    "vipType": 11
  },
  "profile": {
    "userId": 1696843520,
    "nickname": "string",
    "avatarUrl": "string"
  }
}
```

关键类型：

- `code: number`
- `account.id: number`
- `profile.nickname: string`
- `profile.avatarUrl: string`

### 5.2 `user_record`（`type=1`，最近一周）

```json
{
  "code": 200,
  "weekData": [
    {
      "playCount": 15,
      "song": {
        "id": 3355479297,
        "name": "string",
        "ar": [{ "id": 12390232, "name": "string" }],
        "al": { "picUrl": "string" }
      }
    }
  ]
}
```

关键类型：

- `code: number`
- `weekData: array`
- `weekData[].playCount: number`
- `weekData[].song.id: number`
- `weekData[].song.name: string`
- `weekData[].song.ar: array`

### 5.3 `user_record`（`type=0`，所有时间）

```json
{
  "code": 200,
  "allData": [
    {
      "playCount": 123,
      "song": {
        "id": 1990571322,
        "name": "string",
        "ar": [{ "id": 123, "name": "string" }],
        "al": { "picUrl": "string" }
      }
    }
  ]
}
```

关键类型：

- `code: number`
- `allData: array`
- `allData[].playCount: number`
- `allData[].song.id: number`

### 5.4 `song_detail`

```json
{
  "code": 200,
  "songs": [
    {
      "id": 3355479297,
      "name": "string",
      "al": { "picUrl": "string" }
    }
  ],
  "privileges": [
    {
      "id": 3355479297,
      "fee": 0,
      "maxbr": 999000
    }
  ]
}
```

关键类型：

- `code: number`
- `songs: array`
- `songs[].id: number`
- `songs[].name: string`
- `songs[].al.picUrl: string`
- `privileges: array`

## 6. 业务字段映射（本项目实际使用）

`index.js` 中实际消费字段：

- 头像：`account.body.profile.avatarUrl + "?param=128y128"`
- 周榜歌曲：
  - `record.body.weekData[0].song.id`
  - `record.body.weekData[0].song.name`
  - `record.body.weekData[0].song.ar[].name`
  - `record.body.weekData[0].playCount`
- 封面：`songDetail.body.songs[0].al.picUrl + "?param=300y300"`

输出并非 JSON API：

- 最终将上述字段渲染为 `card.svg`。

## 7. 可靠性评级说明

本文件采用如下评级：

- `A`：成功率 >= 99% 且样本 >= 5
- `B`：成功率 >= 95%
- `C`：成功率 < 95% 或样本不足

当前接口评级（基于 6 轮样本）：

- `user_account`：`A`
- `user_record`：`A`
- `song_detail`：`A`

限制：

- 当前样本来自单账号、单网络环境。
- 若 `MUSIC_U` 失效、账号风控、上游策略变化，可靠性会下降。

## 8. 复现与更新文档数据

抓取最新请求/响应：

```bash
npm run capture:requests
```

抓取并保留完整敏感头（仅本地调试）：

```bash
CAPTURE_INCLUDE_SECRETS=1 npm run capture:requests
```

主要产物：

- `artifacts/api-request-captures/latest.json`（完整请求+响应）
- `artifacts/api-request-captures/latest-responses.json`（明文响应聚合）
- `artifacts/api-request-captures/latest.http`（请求回放模板）
- `artifacts/api-request-captures/reliability-summary.json`（可靠性统计）
