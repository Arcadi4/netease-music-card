require('dotenv').config();
const { Octokit } = require('@octokit/rest');
const { user_record, song_detail, user_account } = require('NeteaseCloudMusicApi');
const axios = require('axios').default;

async function getBase64(url) {
    const response = await axios.get(url, { responseType: 'arraybuffer' });
    return Buffer.from(response.data, 'binary').toString('base64');
}

const {
    USER_ID,
    USER_TOKEN,
    GH_TOKEN,
} = process.env;

(async () => {
    const account = await user_account({
        cookie: `MUSIC_U=${USER_TOKEN}`,
    })

    const avatarUrl = account.body.profile.avatarUrl + "?param=32y32"; // 压缩
    console.log(`个人头像: ${avatarUrl}`);

    /*
      获取歌单记录
    */
   
    const record = await user_record({
        cookie: `MUSIC_U=${USER_TOKEN}`,
        uid: USER_ID,
        type: 1,
    }).catch(error => console.error(`无法获取用户播放记录 \n${error}`));

    const content = record.body;
    const songId = content.weekData[0].song.id + '';
    const songName = content.weekData[0].song.name;
    const songAuthorArray = content.weekData[0].song.ar;
    const playCount = content.weekData[0].playCount;

    const songAuthors = songAuthorArray.map(i => i.name).join(' / ');

    const songDetail = await song_detail({
        cookie: `MUSIC_U=${USER_TOKEN}`,
        ids: songId,
    }).catch(error => console.error(`无法获取歌曲信息 \n${error}`));

    const songCover = songDetail.body.songs[0].al.picUrl + "?param=300y300";

    console.log(`歌曲名：${songName}\n歌曲作者：${songAuthors}\n歌曲封面：${songCover}\n播放次数：${playCount}`);

    var svgContent = "";
    try {
        svgContent = Buffer.from(
`<svg width="320" height="480" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
<foreignObject width="320" height="480">
    <div xmlns="http://www.w3.org/1999/xhtml" class="container" style="padding: 5px;">
    <style>
        * {
            box-sizing: border-box;
            image-rendering: crisp-edges;
            image-rendering: -webkit-optimize-contrast;
            color:black;
            font-size: 0;
        }

        html, body {
            margin: 0;
            padding: 3px;
        }

        p {
            margin: 0;
            padding: 0;
        }

        img {
            margin: 0;
            padding: 0;
        }

        .clear {
            clear: both;
        }

        .card {
            display: inline-block;
            background: white;
            border-radius: 10px;
            text-align: center;
            box-shadow: gray 0 0 10px;
            overflow: hidden;
        }

        .user {
            text-align: left;
            margin: 10px;
        }

        .avatar {
            width: 32px;
            height: 32px;
            border-radius: 100%;
            vertical-align: middle;
        }

        .username {
            line-height: 32px;
            vertical-align: middle;
            font-size: 16px;
            margin-left: 5px;
        }

        .button {
            background: #28C131;
            width: 16px;
            height: 16px;
            border-radius: 100%;
            vertical-align: middle;
            float: right;
            margin-top: 8px;
            margin-right: 3px;
        }

        .hello {
            margin: 10px;
            margin-top: 5px;
        }

        .neteasecloud {
            width: 28px;
            height: 28px;
            vertical-align: middle;
        }

        .intro {
            vertical-align: middle;
            color: #BA0400;
            font-size: 16px;
            margin-left: 10px;
        }

        .song {
            margin: 0 auto;
            width: 260px;
            font-size: 20px;
            margin-top: 10px;
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;
        }

        .singer {
            margin: 0 auto;
            width: 260px;
            margin-top: 5px;
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;
            opacity: 0.5;
            font-size: 16px;
        }

        .cover {
            width: 300px;
            height: 300px;
            margin-top: 20px;
        }
    </style>
        <div class="card">
            <div class="user">
                <img class="avatar" src="data:image/jpg;base64,${await getBase64(avatarUrl)}"/>
                <a class="username">Nthily</a>
                <a class="button"></a>
                <div class="clear"></div>
            </div>
            <div class="hello">
                <img class="neteasecloud" src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAABoVBMVEUAAADdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABviJz7pXW7nS17eBiDugo/////62t7hHDTgEyzufYvdARzlOk/50tf+9vfteojpWGr97vDgGTH3xszjKkHviZXteIb86OrrbHzra3v509jqX3D0rbbfDijiIjnwkp7ynqjqYnP98vPzpa/wjpr++vv87O7pWmv74OPeCiT0q7T3w8n3wsj97/HraXndAx786uzrZ3foU2X++/vpV2nnSVz85+rfECn0rrf//PzjMEbjLEL2u8LeCSP+9/jjL0X4ytD1tb3nTF/74uXkNkv3wMfmRFj//f3pW2z619v86evlO1D0r7fscH/+9PXmRVnnSFv98PL1sbn0qbLgFS786+363eHoVmjhGzPkOE35z9Tyn6n85un62Nz97e/nSl350dbqXm/73+LraHjjK0Hufoz4yc/eBR/51dnnTWDjLUP98fLeCyXwkJz4yM7tdYTlOU7oUmT0qrP1srrranrkM0n4ztP+9fb75Of1tLzfDCZ1sXtGAAAAEHRSTlMAE2Ok0vMup/oNl/4r2zjxygE/KQAAAAFiS0dEFnzRqBkAAAAHdElNRQflCQ0QCDMeCNJTAAABwElEQVQ4y41T9VtCUQx9lAiiTsWjiIndhdiN3d2Kjd3dHX+1F+59DzA+PD9s+9727razTZIUqNQarY5Ip9WoVdJPhOhDSUGoPuSb22AMowCEGQ3+flM4/UC4yeePiKRfEBmh/B/oj4qOMfMI8YYh8P1YAHHxPAuvwyi7LAkJVkoEkpKRwr8Yvf3J9aemATZKQTplZNpEL55u9cKfBWTnEOUizy+fnvEn+MkHCgqJioqBktJChTGVpOaWrQzlVjLbK+CBo1KOUEsabthRVU1UA9TWwVEPNIgAjaT16sYmNBO1oNVJbWjv6ERXNw/QSjqv7gEsRL3IJOrrx4B5EEM8QCdxPYwRJh0YZXIMpTSOCQv3iIBJTDE5jRkmZzFHjfNYEAE8hQuLTC5hmWhlFS6iNayLFLxIN8Cq2pjA5tY2dtisdrEnihRt7qOYyYN+RsJhD1E34BZtCqKOUHHs4SvnxNnHdAFOZaIE1WfnWHMrI8gFLmSq5WFdXqG9jZvXN8CWMixl3Bts2rd3rvuHGwfwaPaNW1mY+CcILD5b/RbGb+Ve7K9pb+8fR5/yYht+XVrfWpv+u/bBDyf46QU/3r/P/wt/IFj7qdvKMgAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAyMS0wOS0xM1QxNjowODo1MSswMDowMKszm9EAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMjEtMDktMTNUMTY6MDg6NTErMDA6MDDabiNtAAAAAElFTkSuQmCC" />
                <a class="intro">这周听了多达 ${playCount} 次</a>
            </div>
            <p class="song">${songName}</p>
            <p class="singer">${songAuthors}</p>
            <img class="cover" src="data:image/jpg;base64,${await getBase64(songCover)}" />
        </div>
    </div>
</foreignObject>
</svg>
`
        ).toString("base64");
    } catch(err) {
        console.error(`处理 SVG 时发生了错误：${err}`);
    }

    try {
        const octokit = new Octokit({
            auth: GH_TOKEN,
        });

        const {
            data: { sha: svgSha }
        } = await octokit.git.createBlob({
            owner: "Nthily",
            repo: "neteasemusic-github-profile",
            content: svgContent,
            encoding: "base64"
        });

        /*const {
            data: { sha: avatarSha }
        } = await octokit.git.createBlob({
            owner: "Nthily",
            repo: "neteasemusic-github-profile",
            content: await getBase64(avatarUrl),
            encoding: "base64"
        });
        const {
            data: { sha: coverSha }
        } = await octokit.git.createBlob({
            owner: "Nthily",
            repo: "neteasemusic-github-profile",
            content: await getBase64(songCover),
            encoding: "base64"
        });*/

        const commits = await octokit.repos.listCommits({
            owner: "Nthily",
            repo: "neteasemusic-github-profile",
        });
        const lastSha = commits.data[0].sha;
        const {
            data: { sha: treeSHA }
        } =  await octokit.git.createTree({
            owner: "Nthily",
            repo: "neteasemusic-github-profile",
            tree: [
                {
                    mode: '100644',
                    path: "music_card.svg",
                    type: "blob",
                    sha: svgSha
                }/*,
                {
                    mode: '100644',
                    path: "cache/avatar.jpg",
                    type: "blob",
                    sha: avatarSha
                },
                {
                    mode: '100644',
                    path: "cache/cover.jpg",
                    type: "blob",
                    sha: coverSha
                }*/
            ],
            base_tree: lastSha,
        });
        const {
            data: { sha: newSHA }
        } =  await octokit.git.createCommit({
            owner: "Nthily",
            repo: "neteasemusic-github-profile",
            author: {
                name: "github-actions[bot]",
                email: "41898282+github-actions[bot]@users.noreply.github.com",
            },
            committer: {
                name: "github-actions[bot]",
                email: "41898282+github-actions[bot]@users.noreply.github.com",
            },
            tree: treeSHA,
            message: 'Update SVG periodically',
            parents: [ lastSha ],
        });
        const result = await octokit.git.updateRef({
            owner: "Nthily",
            repo: "neteasemusic-github-profile",
            ref: "heads/main",
            sha: newSHA,
        });
        console.log(result);
    } catch(err) {
        console.error(`上传 SVG 时发生了错误：${err}`);
    }

})();

