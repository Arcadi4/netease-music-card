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

    const avatarUrl = account.body.profile.avatarUrl;
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

    const songPicUrl = songDetail.body.songs[0].al.picUrl;

    console.log(`歌曲名：${songName}\n歌曲作者：${songAuthors}\n歌曲封面：${songPicUrl}\n播放次数：${playCount}`);

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
                <img class="neteasecloud" src="data:image/x-icon;base64,AAABAAEAICAAAAEAIACoEAAAFgAAACgAAAAgAAAAQAAAAAEAIAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGwDdExsA3WMbAN2kGwDd0hsA3fMbAN3/GwDd/xsA3fMbAN3SGwDdpBsA3WMbAN0TAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGwDdLhsA3acbAN36GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3fobAN2nGwDdLgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGwDdDRsA3ZcbAN3+GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3+GwDdlxsA3Q0AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABsA3SsbAN3bGwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd2xsA3SsAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAbAN04GwDd8RsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd8RsA3TgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGwDdKxsA3fEbAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd8RsA3SsAAAAAAAAAAAAAAAAAAAAAAAAAABsA3Q0bAN3bGwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f9JM+T/mo7w/9PO+P/w7v3/9vX+/+fk+/+8tPX/eGjr/yYM3/8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd2xsA3Q0AAAAAAAAAAAAAAAAAAAAAGwDdlxsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f80HOH/urL1////////////////////////////////////////////8vD9/3pq6/8cAd3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDdlwAAAAAAAAAAAAAAABsA3S4bAN3+GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/TTjk//Du/f///////////87I+P+Ede3/W0jn/0455f9kUuj/nJDw/+7s/P///////////7Oq9P8gBt7/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3+GwDdLgAAAAAAAAAAGwDdpxsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/0Mt4//y8f3//////+rn/P9cSef/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/JQve/5yQ8P///////////6mf8v8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN2nAAAAABsA3RMbAN36GwDd/xsA3f8bAN3/GwDd/xsA3f8fBd7/2dX5///////h3fr/Mxvh/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/3tr6////////////2BN5/8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3fobAN0TGwDdYxsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/3ho6///////+/r+/0Er4/8bAN3/GwDd/xsA3f8xGeD/fGzr/4x+7v9pV+n/Iwne/xsA3f8bAN3/GwDd/7at9P//////z8n4/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3WMbAN2kGwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/1tH5//////+ekvD/GwDd/xsA3f8bAN3/b17q//j3/v/////////////////i3/v/Qizj/xsA3f8bAN3/RjDj////////////Mxvh/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDdpBsA3dIbAN3/GwDd/xsA3f8bAN3/GwDd/zMb4f///////////0Yw4/8bAN3/GwDd/0045P/9/f///////9TP+f+pn/L/6eb8///////c2Pr/IAbe/xsA3f8bAN3/7+39//////9dSuf/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3SGwDd8xsA3f8bAN3/GwDd/xsA3f8bAN3/W0jn///////y8P3/GwDd/xsA3f8bAN3/ubH1//////+yqfT/GwDd/xsA3f8uFeD/7ev8//////9bSOf/GwDd/xsA3f/h3fr//////2hW6P8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3fMbAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f9sW+n//////9vX+v8bAN3/GwDd/xsA3f/r6fz//////1A75f8bAN3/GwDd/xsA3f+3r/T//////39w7P8bAN3/HgPd//X0/v//////WUXm/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/19M5///////5eL7/xsA3f8bAN3/GwDd/+7s/P//////Szbk/xsA3f8bAN3/GwDd/8fA9///////c2Lq/xsA3f9YROb///////39//8sE+D/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd8xsA3f8bAN3/GwDd/xsA3f8bAN3/Qizj////////////LBPg/xsA3f8bAN3/wrv2//////+Vie//GwDd/xsA3f8jCd7/+Pf+//////9FL+P/IAbe/9DK+P//////vbX1/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3fMbAN3SGwDd/xsA3f8bAN3/GwDd/xsA3f8eA93/7Or8//////93Z+v/GwDd/xsA3f9lU+j///////v7/v9pV+n/GwDd/1xJ5///////6uf8/ykQ3/+3rvT///////z8//9GMOP/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd0hsA3aQbAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f+ajvD//////+Pg+/8kCt7/GwDd/xsA3f+0q/T////////////Jw/f/yML3///////u7Pz/8e/9////////////eWnr/xsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/xsA3f8bAN2kGwDdYxsA3f8bAN3/GwDd/xsA3f8bAN3/GwDd/zEZ4P/z8v3//////6+l8/8cAd3/GwDd/yAG3v+ajvD/+/r+////////////////////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" />
                <a class="intro">这周听了多达 ${playCount} 次</a>
            </div>
            <p class="song">${songName}</p>
            <p class="singer">${songAuthors}</p>
            <img class="cover" src="data:image/jpg;base64,${await getBase64(songPicUrl)}" />
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
                }
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
            message: 'update SVG periodically - [Skip Github Action]',
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

