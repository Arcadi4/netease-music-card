require('dotenv').config();
const { Octokit } = require('@octokit/rest');
const { user_record, song_detail, user_account, user_detail } = require('NeteaseCloudMusicApi');
const axios = require('axios').default;
const fs = require('fs');
const path = require('path');

async function getBase64(url) {
    const response = await axios.get(url, { responseType: 'arraybuffer' });
    return Buffer.from(response.data, 'binary').toString('base64');
}

const {
    USER_ID,
    USER_TOKEN,
    GH_TOKEN,
    AUTHOR,
    REPO,
} = process.env;

// ─── Preflight Validation ──────────────────────────────────────────
const REQUIRED_ENVS = ['USER_ID', 'USER_TOKEN', 'GH_TOKEN', 'AUTHOR', 'REPO'];
const missingEnvs = REQUIRED_ENVS.filter(k => !process.env[k]);
if (missingEnvs.length > 0) {
    console.error(`[preflight] Missing required environment variables: ${missingEnvs.join(', ')}`);
    process.exit(1);
}

// ─── Shared Style Tokens ────────────────────────────────────────────
const STYLE = {
    width: 310,
    height: 490,
    borderRadius: "10px",
    shadow: "gray 0 0 10px",
    accent: "#BA0400",
    green: "#28C131",
    fontStack: "'PingFang SC', 'Helvetica Neue', 'Segoe UI', 'Microsoft YaHei', sans-serif",
};

// ─── Stats Derivation ────────────────────────────────────────────────
function safeWeekData(rawBody) {
    return rawBody?.weekData ?? [];
}

function safePlays(entry) {
    return entry?.playCount > 0 ? entry.playCount : (entry?.score ?? 0);
}

function deriveTopArtists(weekData, n = 5) {
    if (!Array.isArray(weekData) || weekData.length === 0) {
        return [];
    }

    const allPlayCountZero = weekData.every(entry => (entry?.playCount ?? 0) === 0);
    const getPlays = allPlayCountZero ? safePlays : (entry => entry?.playCount ?? 0);
    const artistMap = new Map();

    for (const entry of weekData) {
        const plays = getPlays(entry);
        const artists = entry?.song?.ar ?? [];
        for (const artist of artists) {
            if (artist?.id == null) {
                continue;
            }

            if (!artistMap.has(artist.id)) {
                artistMap.set(artist.id, {
                    id: artist.id,
                    name: artist.name,
                    plays: 0,
                });
            }
            artistMap.get(artist.id).plays += plays;
        }
    }

    return [...artistMap.values()]
        .sort((a, b) => b.plays - a.plays)
        .slice(0, n);
}

function deriveWeeklyOverview(weekData) {
    if (!Array.isArray(weekData) || weekData.length === 0) {
        return {
            totalPlays: 0,
            uniqueSongs: 0,
            uniqueArtists: 0,
            repeatIntensity: 0,
        };
    }

    const uniqueSongIds = new Set();
    const uniqueArtistIds = new Set();
    let totalPlays = 0;
    let maxPlayCount = 0;

    for (const entry of weekData) {
        const plays = safePlays(entry);
        totalPlays += plays;
        if (plays > maxPlayCount) {
            maxPlayCount = plays;
        }

        const songId = entry?.song?.id;
        if (songId != null) {
            uniqueSongIds.add(songId);
        }

        const artists = entry?.song?.ar ?? [];
        for (const artist of artists) {
            if (artist?.id != null) {
                uniqueArtistIds.add(artist.id);
            }
        }
    }

    return {
        totalPlays,
        uniqueSongs: uniqueSongIds.size,
        uniqueArtists: uniqueArtistIds.size,
        repeatIntensity: totalPlays > 0 ? (maxPlayCount / totalPlays * 100).toFixed(1) : 0,
    };
}

function deriveTopTracks(weekData, n = 5) {
    if (!Array.isArray(weekData) || weekData.length === 0) {
        return [];
    }

    const allPlayCountZero = weekData.every(entry => (entry?.playCount ?? 0) === 0);
    const getPlays = allPlayCountZero ? safePlays : (entry => entry?.playCount ?? 0);

    return weekData
        .map(entry => ({
            name: entry?.song?.name ?? '',
            artists: (entry?.song?.ar ?? []).map(artist => artist?.name).filter(Boolean).join(' / '),
            plays: getPlays(entry),
        }))
        .sort((a, b) => b.plays - a.plays)
        .slice(0, n);
}

// ─── Duration Snapshot ──────────────────────────────────────────────
function loadDurationSnapshot(snapshotPath) {
    snapshotPath = snapshotPath || path.resolve(__dirname, 'duration-snapshot.json');
    try {
        const raw = fs.readFileSync(snapshotPath, 'utf8');
        const parsed = JSON.parse(raw);
        if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
            return parsed;
        }
        return {};
    } catch (e) {
        return {};
    }
}

function saveDurationSnapshot(snapshotPath, snapshot) {
    snapshotPath = snapshotPath || path.resolve(__dirname, 'duration-snapshot.json');
    fs.writeFileSync(snapshotPath, JSON.stringify(snapshot, null, 2), 'utf8');
}

function updateDurationSnapshot(snapshot, listenSongs) {
    const todayISO = new Date().toISOString().slice(0, 10);
    snapshot[todayISO] = listenSongs;
    return snapshot;
}

function deriveDailyDurations(snapshot, avgMinPerSong) {
    if (avgMinPerSong === undefined || avgMinPerSong === null) {
        avgMinPerSong = 3.5;
    }

    const dayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
    const today = new Date();
    const dow = today.getDay(); // 0=Sun, 1=Mon...6=Sat
    const monday = new Date(today);
    monday.setDate(today.getDate() - ((dow + 6) % 7));

    const result = [];
    for (let i = 0; i < 7; i++) {
        const d = new Date(monday);
        d.setDate(monday.getDate() + i);
        const dateStr = d.toISOString().slice(0, 10);

        const prevD = new Date(d);
        prevD.setDate(d.getDate() - 1);
        const prevDateStr = prevD.toISOString().slice(0, 10);

        let estimatedMinutes = null;
        if (snapshot[dateStr] !== undefined && snapshot[prevDateStr] !== undefined) {
            const delta = snapshot[dateStr] - snapshot[prevDateStr];
            if (delta >= 0) {
                estimatedMinutes = Math.round(delta * avgMinPerSong * 10) / 10;
            }
        }

        result.push({
            date: dateStr,
            day: dayNames[i],
            estimatedMinutes: estimatedMinutes,
        });
    }

    return result;
}

// ─── Data Fetching Layer ────────────────────────────────────────────
async function fetchData(cookie, userId) {
    let account;
    try {
        account = await user_account({ cookie });
    } catch (err) {
        throw new Error(`[auth] user_account request failed: ${err.message || err}`);
    }
    if (!account?.body?.profile) {
        throw new Error('[auth] Cookie appears invalid or expired — user_account returned no profile');
    }

    const avatarUrl = account.body.profile.avatarUrl + "?param=128y128";
    console.log(`[info] avatar: ${avatarUrl}`);

    const nickname = account.body.profile.nickname;

    const record = await user_record({
        cookie,
        uid: userId,
        type: 1,
    }).catch(error => {
        console.error(`[error] Failed to fetch user play record: ${error}`);
        return null;
    });

    if (!record?.body) {
        throw new Error('[auth] Could not retrieve play records — cookie may be expired');
    }

    const content = record.body;
    const weekData = safeWeekData(content);
    const songId = weekData[0].song.id + '';
    const songName = weekData[0].song.name;
    const songAuthorArray = weekData[0].song.ar;
    const playCount = weekData[0].playCount;

    const songAuthors = songAuthorArray.map(i => i.name).join(' / ');

    const songDetailResult = await song_detail({
        cookie,
        ids: songId,
    }).catch(error => {
        console.error(`[error] Failed to fetch song detail: ${error}`);
        return null;
    });

    const songCover = songDetailResult?.body?.songs?.[0]?.al?.picUrl
        ? songDetailResult.body.songs[0].al.picUrl + "?param=300y300"
        : '';

    console.log(`[info] song=${songName} artists=${songAuthors} plays=${playCount}`);

    let dailyDurations = [
        { date: null, day: 'Mon', estimatedMinutes: null },
        { date: null, day: 'Tue', estimatedMinutes: null },
        { date: null, day: 'Wed', estimatedMinutes: null },
        { date: null, day: 'Thu', estimatedMinutes: null },
        { date: null, day: 'Fri', estimatedMinutes: null },
        { date: null, day: 'Sat', estimatedMinutes: null },
        { date: null, day: 'Sun', estimatedMinutes: null },
    ];

    const snapshotPath = path.resolve(__dirname, 'duration-snapshot.json');
    try {
        const detailResult = await user_detail({ cookie, uid: userId });
        const listenSongs = detailResult.body.profile.listenSongs;
        console.log(`[info] total listen count: ${listenSongs}`);

        let snapshot = loadDurationSnapshot(snapshotPath);
        snapshot = updateDurationSnapshot(snapshot, listenSongs);
        saveDurationSnapshot(snapshotPath, snapshot);
        dailyDurations = deriveDailyDurations(snapshot);
    } catch (err) {
        console.error(`[error] Failed to update duration snapshot: ${err}`);
    }

    return {
        avatarUrl,
        nickname,
        songName,
        songAuthors,
        songCover,
        playCount,
        weekData,
        dailyDurations,
    };
}

// ─── Card Renderer (card.svg) ───────────────────────────────────────
async function renderCard(data) {
    const { avatarUrl, nickname, songName, songAuthors, songCover, playCount } = data;

    var svgContent = "";
    try {
        svgContent = Buffer.from(
            `<svg width="${STYLE.width}" height="${STYLE.height}" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
    <foreignObject width="${STYLE.width}" height="${STYLE.height}">
        <div xmlns="http://www.w3.org/1999/xhtml" class="container" style="padding: 5px;">
        <style>
            * {
                box-sizing: border-box;
                color:black;
                font-size: 0;
                font-family: ${STYLE.fontStack} !important;
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
                border-radius: ${STYLE.borderRadius};
                text-align: center;
                box-shadow: ${STYLE.shadow};
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
                background: ${STYLE.green};
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
                color: ${STYLE.accent};
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
            }

            .bars {
                position: relative;
                margin-top: 3px;
                height: 30px;
            }

            .bar {
                background: red;
                bottom: 0;
                height: 3px;
                position: absolute;
                width: 3px;
                animation: sound 0ms -1000ms linear infinite alternate;
            }

            @keyframes sound {
                0% {
                    opacity: .35;
                    height: 3px;
                }
                100% {
                    opacity: 1;
                    height: 20px;
                }
            }

            .bar:nth-child(1)  { left: 1px; animation-duration: 412ms; }.bar:nth-child(2)  { left: 5px; animation-duration: 413ms; }.bar:nth-child(3)  { left: 9px; animation-duration: 466ms; }.bar:nth-child(4)  { left: 13px; animation-duration: 452ms; }.bar:nth-child(5)  { left: 17px; animation-duration: 423ms; }.bar:nth-child(6)  { left: 21px; animation-duration: 486ms; }.bar:nth-child(7)  { left: 25px; animation-duration: 416ms; }.bar:nth-child(8)  { left: 29px; animation-duration: 380ms; }.bar:nth-child(9)  { left: 33px; animation-duration: 405ms; }.bar:nth-child(10)  { left: 37px; animation-duration: 447ms; }.bar:nth-child(11)  { left: 41px; animation-duration: 409ms; }.bar:nth-child(12)  { left: 45px; animation-duration: 490ms; }.bar:nth-child(13)  { left: 49px; animation-duration: 392ms; }.bar:nth-child(14)  { left: 53px; animation-duration: 481ms; }.bar:nth-child(15)  { left: 57px; animation-duration: 373ms; }.bar:nth-child(16)  { left: 61px; animation-duration: 444ms; }.bar:nth-child(17)  { left: 65px; animation-duration: 397ms; }.bar:nth-child(18)  { left: 69px; animation-duration: 464ms; }.bar:nth-child(19)  { left: 73px; animation-duration: 396ms; }.bar:nth-child(20)  { left: 77px; animation-duration: 397ms; }.bar:nth-child(21)  { left: 81px; animation-duration: 370ms; }.bar:nth-child(22)  { left: 85px; animation-duration: 486ms; }.bar:nth-child(23)  { left: 89px; animation-duration: 484ms; }.bar:nth-child(24)  { left: 93px; animation-duration: 430ms; }.bar:nth-child(25)  { left: 97px; animation-duration: 402ms; }.bar:nth-child(26)  { left: 101px; animation-duration: 479ms; }.bar:nth-child(27)  { left: 105px; animation-duration: 417ms; }.bar:nth-child(28)  { left: 109px; animation-duration: 391ms; }.bar:nth-child(29)  { left: 113px; animation-duration: 433ms; }.bar:nth-child(30)  { left: 117px; animation-duration: 369ms; }.bar:nth-child(31)  { left: 121px; animation-duration: 439ms; }.bar:nth-child(32)  { left: 125px; animation-duration: 387ms; }.bar:nth-child(33)  { left: 129px; animation-duration: 423ms; }.bar:nth-child(34)  { left: 133px; animation-duration: 462ms; }.bar:nth-child(35)  { left: 137px; animation-duration: 435ms; }.bar:nth-child(36)  { left: 141px; animation-duration: 480ms; }.bar:nth-child(37)  { left: 145px; animation-duration: 429ms; }.bar:nth-child(38)  { left: 149px; animation-duration: 467ms; }.bar:nth-child(39)  { left: 153px; animation-duration: 429ms; }.bar:nth-child(40)  { left: 157px; animation-duration: 469ms; }.bar:nth-child(41)  { left: 161px; animation-duration: 372ms; }.bar:nth-child(42)  { left: 165px; animation-duration: 481ms; }.bar:nth-child(43)  { left: 169px; animation-duration: 408ms; }.bar:nth-child(44)  { left: 173px; animation-duration: 457ms; }.bar:nth-child(45)  { left: 177px; animation-duration: 465ms; }.bar:nth-child(46)  { left: 181px; animation-duration: 353ms; }.bar:nth-child(47)  { left: 185px; animation-duration: 430ms; }.bar:nth-child(48)  { left: 189px; animation-duration: 401ms; }.bar:nth-child(49)  { left: 193px; animation-duration: 356ms; }.bar:nth-child(50)  { left: 197px; animation-duration: 383ms; }.bar:nth-child(51)  { left: 201px; animation-duration: 491ms; }.bar:nth-child(52)  { left: 205px; animation-duration: 476ms; }.bar:nth-child(53)  { left: 209px; animation-duration: 396ms; }.bar:nth-child(54)  { left: 213px; animation-duration: 373ms; }.bar:nth-child(55)  { left: 217px; animation-duration: 362ms; }.bar:nth-child(56)  { left: 221px; animation-duration: 409ms; }.bar:nth-child(57)  { left: 225px; animation-duration: 375ms; }.bar:nth-child(58)  { left: 229px; animation-duration: 427ms; }.bar:nth-child(59)  { left: 233px; animation-duration: 368ms; }.bar:nth-child(60)  { left: 237px; animation-duration: 412ms; }.bar:nth-child(61)  { left: 241px; animation-duration: 381ms; }.bar:nth-child(62)  { left: 245px; animation-duration: 478ms; }.bar:nth-child(63)  { left: 249px; animation-duration: 449ms; }.bar:nth-child(64)  { left: 253px; animation-duration: 426ms; }.bar:nth-child(65)  { left: 257px; animation-duration: 481ms; }.bar:nth-child(66)  { left: 261px; animation-duration: 363ms; }.bar:nth-child(67)  { left: 265px; animation-duration: 365ms; }.bar:nth-child(68)  { left: 269px; animation-duration: 467ms; }.bar:nth-child(69)  { left: 273px; animation-duration: 424ms; }.bar:nth-child(70)  { left: 277px; animation-duration: 432ms; }.bar:nth-child(71)  { left: 281px; animation-duration: 464ms; }.bar:nth-child(72)  { left: 285px; animation-duration: 453ms; }.bar:nth-child(73)  { left: 289px; animation-duration: 427ms; }.bar:nth-child(74)  { left: 293px; animation-duration: 452ms; }.bar:nth-child(75)  { left: 297px; animation-duration: 447ms; }
    
        </style>
            <div class="card">
                <div class="user">
                    <img class="avatar" src="data:image/jpg;base64,${await getBase64(avatarUrl)}"/>
                    <a class="username">${nickname}</a>
                    <a class="button"></a>
                    <div class="clear"></div>
                </div>
                <div class="hello">
                    <img class="neteasecloud" src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAABoVBMVEUAAADdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABvdABviJz7pXW7nS17eBiDugo/////62t7hHDTgEyzufYvdARzlOk/50tf+9vfteojpWGr97vDgGTH3xszjKkHviZXteIb86OrrbHzra3v509jqX3D0rbbfDijiIjnwkp7ynqjqYnP98vPzpa/wjpr++vv87O7pWmv74OPeCiT0q7T3w8n3wsj97/HraXndAx786uzrZ3foU2X++/vpV2nnSVz85+rfECn0rrf//PzjMEbjLEL2u8LeCSP+9/jjL0X4ytD1tb3nTF/74uXkNkv3wMfmRFj//f3pW2z619v86evlO1D0r7fscH/+9PXmRVnnSFv98PL1sbn0qbLgFS786+363eHoVmjhGzPkOE35z9Tyn6n85un62Nz97e/nSl350dbqXm/73+LraHjjK0Hufoz4yc/eBR/51dnnTWDjLUP98fLeCyXwkJz4yM7tdYTlOU7oUmT0qrP1srrranrkM0n4ztP+9fb75Of1tLzfDCZ1sXtGAAAAEHRSTlMAE2Ok0vMup/oNl/4r2zjxygE/KQAAAAFiS0dEFnzRqBkAAAAHdElNRQflCQ0QCDMeCNJTAAABwElEQVQ4y41T9VtCUQx9lAiiTsWjiIndhdiN3d2Kjd3dHX+1F+59DzA+PD9s+9727bazTZIUqNQarY5Ip9WoVdJPhOhDSUGoPuSb22AMowCEGQ3+flM4/UC4yeePiKRfEBmh/B/oj4qOMfMI8YYh8P1YAHHxPAuvwyi7LAkJVkoEkpKRwr8Yvf3J9aemATZKQTplZNpEL55u9cKfBWTnEOUizy+fnvEn+MkHCgqJioqBktJChTGVpOaWrQzlVjLbK+CBo1KOUEsabthRVU1UA9TWwVEPNIgAjaT16sYmNBO1oNVJbWjv6ERXNw/QSjqv7gEsRL3IJOrrx4B5EEM8QCdxPYwRJh0YZXIMpTSOCQv3iIBJTDE5jRkmZzFHjfNYEAE8hQuLTC5hmWhlFS6iNayLFLxIN8Cq2pjA5tY2dtisdrEnihRt7qOYyYN+RsJhD1E34BZtCqKOUHHs4SvnxNnHdAFOZaIE1WfnWHMrI8gFLmSq5WFdXqG9jZvXN8CWMixl3Bts2rd3rvuHGwfwaPaNW1mY+CcILD5b/RbGb+Ve7K9pb+8fR5/yYht+XVrfWpv+u/bBDyf46QU/3r/P/wt/IFj7qdvKMgAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAyMS0wOS0xM1QxNjowODo1MSswMDowMKszm9EAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMjEtMDktMTNUMTY6MDg6NTErMDA6MDDabiNtAAAAAElFTkSuQmCC" />
                    <a class="intro">这周听了多达 ${playCount} 次</a>
                </div>
                <p class="song">${songName}</p>
                <p class="singer">${songAuthors}</p>
                <div class="bars">
                    <div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div>
                </div>
                <img class="cover" src="data:image/jpg;base64,${await getBase64(songCover)}" />
            </div>
        </div>
    </foreignObject>
</svg>
`
        ).toString("base64");
    } catch (err) {
        console.error(`处理 SVG 时发生了错误：${err}`);
    }

    return { filename: "card.svg", svgBase64: svgContent };
}

// ─── Renderer Registry ──────────────────────────────────────────────
async function renderTopArtists(data) {
    const artists = deriveTopArtists(data.weekData, 5);
    let artistRows = "";

    if (artists.length === 0) {
        artistRows = `<div class="empty-state">暂无数据 / No data</div>`;
    } else {
        artists.forEach((artist, index) => {
            const rank = index + 1;
            artistRows += `
                <div class="row">
                    <div class="rank">${rank}</div>
                    <div class="info">
                        <div class="name">${artist.name}</div>
                        <div class="plays">${artist.plays} plays</div>
                    </div>
                </div>
            `;
        });
    }

    var svgContent = "";
    try {
        svgContent = Buffer.from(
            `<svg width="${STYLE.width}" height="320" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
    <foreignObject width="${STYLE.width}" height="320">
        <div xmlns="http://www.w3.org/1999/xhtml" class="container" style="padding: 5px;">
        <style>
            * {
                box-sizing: border-box;
                font-family: ${STYLE.fontStack} !important;
                margin: 0;
                padding: 0;
            }
            .card {
                background: white;
                border-radius: ${STYLE.borderRadius};
                box-shadow: ${STYLE.shadow};
                overflow: hidden;
                padding: 20px;
                height: full;
                display: flex;
                flex-direction: column;
            }
            .header {
                margin-bottom: 15px;
            }
            .title {
                font-size: 18px;
                font-weight: bold;
                color: #333;
            }
            .subtitle {
                font-size: 12px;
                color: #888;
                margin-top: 2px;
            }
            .row {
                display: flex;
                align-items: center;
                margin-bottom: 12px;
            }
            .row:last-child {
                margin-bottom: 0;
            }
            .rank {
                background: ${STYLE.accent};
                color: white;
                font-size: 14px;
                font-weight: bold;
                width: 24px;
                height: 24px;
                border-radius: 4px;
                display: flex;
                align-items: center;
                justify-content: center;
                margin-right: 12px;
                flex-shrink: 0;
            }
            .info {
                flex-grow: 1;
                overflow: hidden;
            }
            .name {
                font-size: 14px;
                color: #333;
                font-weight: 500;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }
            .plays {
                font-size: 12px;
                color: #888;
                margin-top: 2px;
            }
            .empty-state {
                flex-grow: 1;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 14px;
                color: #888;
            }
        </style>
            <div class="card">
                <div class="header">
                    <div class="title">本周 Top 艺术家</div>
                    <div class="subtitle">Weekly Top Artists</div>
                </div>
                <div class="list">
                    ${artistRows}
                </div>
            </div>
        </div>
    </foreignObject>
</svg>`
        ).toString("base64");
    } catch (err) {
        console.error(`top-artists SVG error: ${err}`);
    }

    return { filename: "top-artists.svg", svgBase64: svgContent };
}

async function renderTopTracks(data) {
    const tracks = deriveTopTracks(data.weekData, 5);
    var svgContent = "";
    try {
        const rows = tracks.length > 0 ? tracks.map((t, i) => `
            <div class="row">
                <div class="rank">${i + 1}</div>
                <div class="info">
                    <div class="title">${t.name}</div>
                    <div class="sub">${t.artists}</div>
                </div>
                <div class="plays">${t.plays}</div>
            </div>`).join('') : `<div class="empty">暂无数据 / No data</div>`;
        svgContent = Buffer.from(`<svg width="${STYLE.width}" height="330" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
    <foreignObject width="${STYLE.width}" height="330">
        <div xmlns="http://www.w3.org/1999/xhtml" class="container" style="padding: 5px;">
        <style>
            * {
                box-sizing: border-box;
                font-family: ${STYLE.fontStack} !important;
                margin: 0;
                padding: 0;
            }
            .card {
                background: white;
                border-radius: ${STYLE.borderRadius};
                box-shadow: ${STYLE.shadow};
                overflow: hidden;
                padding: 20px;
                height: full;
                display: flex;
                flex-direction: column;
            }
            .header {
                margin-bottom: 15px;
            }
            .title-header {
                font-size: 18px;
                font-weight: bold;
                color: #333;
            }
            .subtitle {
                font-size: 12px;
                color: #888;
                margin-top: 2px;
            }
            .row {
                display: flex;
                align-items: center;
                margin-bottom: 12px;
            }
            .row:last-child {
                margin-bottom: 0;
            }
            .rank {
                background: ${STYLE.accent};
                color: white;
                font-size: 14px;
                font-weight: bold;
                width: 24px;
                height: 24px;
                border-radius: 4px;
                display: flex;
                align-items: center;
                justify-content: center;
                margin-right: 12px;
                flex-shrink: 0;
            }
            .info {
                flex-grow: 1;
                overflow: hidden;
            }
            .title {
                font-size: 14px;
                color: #333;
                font-weight: 500;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }
            .sub {
                font-size: 12px;
                color: #888;
                margin-top: 2px;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }
            .plays {
                font-size: 12px;
                color: #888;
                margin-left: 10px;
                flex-shrink: 0;
                background: #f0f0f0;
                padding: 2px 6px;
                border-radius: 10px;
            }
            .empty {
                flex-grow: 1;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 14px;
                color: #888;
            }
        </style>
            <div class="card">
                <div class="header">
                    <div class="title-header">本周 Top 歌曲</div>
                    <div class="subtitle">Weekly Top Tracks</div>
                </div>
                <div class="list">
                    ${rows}
                </div>
            </div>
        </div>
    </foreignObject>
</svg>`).toString("base64");
    } catch (err) { console.error(`top-tracks SVG error: ${err}`); }
    return { filename: "top-tracks.svg", svgBase64: svgContent };
}

async function renderWeeklyOverview(data) {
    var svgContent = "";
    try {
        const stats = deriveWeeklyOverview(data.weekData);
        svgContent = Buffer.from(
            `<svg width="${STYLE.width}" height="260" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
    <foreignObject width="${STYLE.width}" height="260">
        <div xmlns="http://www.w3.org/1999/xhtml" style="padding: 5px;">
        <style>
            * {
                box-sizing: border-box;
                font-family: ${STYLE.fontStack} !important;
                margin: 0;
                padding: 0;
            }
            .card {
                background: white;
                border-radius: ${STYLE.borderRadius};
                box-shadow: ${STYLE.shadow};
                width: 300px;
                height: full;
                display: flex;
                flex-direction: column;
                overflow: hidden;
            }
            .header {
                padding: 15px;
                border-bottom: 1px solid #eee;
                text-align: center;
            }
            .title {
                font-size: 18px;
                font-weight: bold;
                color: #333;
            }
            .subtitle {
                font-size: 12px;
                color: #999;
                margin-top: 2px;
            }
            .grid {
                display: grid;
                grid-template-columns: 1fr 1fr;
                grid-template-rows: 1fr 1fr;
                flex: 1;
            }
            .cell {
                display: flex;
                flex-direction: column;
                justify-content: center;
                align-items: center;
                border-right: 1px solid #f5f5f5;
                border-bottom: 1px solid #f5f5f5;
            }
            .cell:nth-child(2n) {
                border-right: none;
            }
            .cell:nth-child(n+3) {
                border-bottom: none;
            }
            .value {
                font-size: 28px;
                font-weight: bold;
                color: ${STYLE.accent};
            }
            .label {
                font-size: 11px;
                color: #666;
                margin-top: 4px;
            }
        </style>
            <div class="card">
                <div class="header">
                    <div class="title">本周概览</div>
                    <div class="subtitle">Weekly Overview</div>
                </div>
                <div class="grid">
                    <div class="cell">
                        <div class="value">${stats.totalPlays}</div>
                        <div class="label">总播放 / Total Plays</div>
                    </div>
                    <div class="cell">
                        <div class="value">${stats.uniqueSongs}</div>
                        <div class="label">不同歌曲 / Unique Songs</div>
                    </div>
                    <div class="cell">
                        <div class="value">${stats.uniqueArtists}</div>
                        <div class="label">不同艺术家 / Artists</div>
                    </div>
                    <div class="cell">
                        <div class="value">${stats.repeatIntensity === 0 ? '0' : stats.repeatIntensity}%</div>
                        <div class="label">重复强度 / Repeat %</div>
                    </div>
                </div>
            </div>
        </div>
    </foreignObject>
</svg>`
        ).toString("base64");
    } catch (err) {
        console.error(`weekly-overview SVG error: ${err}`);
    }
    return { filename: "weekly-overview.svg", svgBase64: svgContent };
}

async function renderWeeklyDuration(data) {
    const durations = data.dailyDurations || [];
    const hasData = durations.some(d => d.estimatedMinutes !== null);

    let chartHtml = "";

    if (!hasData) {
        chartHtml = `<div class="empty-state">暂无数据 / No data</div>`;
    } else {
        const maxMinutes = Math.max(...durations.map(d => d.estimatedMinutes || 0));
        const MAX_BAR_HEIGHT = 100;

        const cols = durations.map(d => {
            if (d.estimatedMinutes === null) {
                return `
                <div class="bar-col">
                    <div class="bar-label-top"></div>
                    <div class="bar-wrap">
                        <div class="bar empty" style="height: 10px;"></div>
                    </div>
                    <div class="bar-day">${d.day}</div>
                </div>`;
            } else {
                const height = maxMinutes > 0 ? (d.estimatedMinutes / maxMinutes) * MAX_BAR_HEIGHT : 10;
                const minHeight = Math.max(height, 10);
                return `
                <div class="bar-col">
                    <div class="bar-label-top">${Math.round(d.estimatedMinutes)}m</div>
                    <div class="bar-wrap">
                        <div class="bar filled" style="height: ${minHeight}px;"></div>
                    </div>
                    <div class="bar-day">${d.day}</div>
                </div>`;
            }
        }).join('');

        chartHtml = `<div class="chart">${cols}</div>`;
    }

    var svgContent = "";
    try {
        svgContent = Buffer.from(
            `<svg width="${STYLE.width}" height="260" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
    <foreignObject width="${STYLE.width}" height="260">
        <div xmlns="http://www.w3.org/1999/xhtml" class="container" style="padding: 5px;">
        <style>
            * {
                box-sizing: border-box;
                font-family: ${STYLE.fontStack} !important;
                margin: 0;
                padding: 0;
            }
            .card {
                background: white;
                border-radius: ${STYLE.borderRadius};
                box-shadow: ${STYLE.shadow};
                overflow: hidden;
                padding: 20px;
                height: full;
                display: flex;
                flex-direction: column;
            }
            .header {
                margin-bottom: 15px;
                text-align: center;
            }
            .title {
                font-size: 18px;
                font-weight: bold;
                color: #333;
            }
            .subtitle {
                font-size: 12px;
                color: #888;
                margin-top: 2px;
            }
            .chart {
                flex-grow: 1;
                display: flex;
                align-items: flex-end;
                justify-content: space-between;
                padding: 0 10px;
            }
            .bar-col {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: flex-end;
                height: 100%;
                width: 30px;
            }
            .bar-label-top {
                font-size: 10px;
                color: #666;
                margin-bottom: 4px;
                height: 12px;
            }
            .bar-wrap {
                width: 100%;
                display: flex;
                align-items: flex-end;
                justify-content: center;
                height: 100px;
            }
            .bar {
                width: 16px;
                border-radius: 3px 3px 0 0;
            }
            .bar.filled {
                background: ${STYLE.accent};
            }
            .bar.empty {
                background: #e0e0e0;
                height: 10px !important;
            }
            .bar-day {
                font-size: 11px;
                color: #888;
                margin-top: 8px;
            }
            .empty-state {
                flex-grow: 1;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 14px;
                color: #888;
            }
        </style>
            <div class="card">
                <div class="header">
                    <div class="title">本周听歌时长</div>
                    <div class="subtitle">估算时长 · Estimated</div>
                </div>
                ${chartHtml}
            </div>
        </div>
    </foreignObject>
</svg>`
        ).toString("base64");
    } catch (err) {
        console.error(`weekly-duration SVG error: ${err}`);
    }

    return { filename: "weekly-duration.svg", svgBase64: svgContent };
}

async function generateAllCards(data) {
    const outputs = [];

    if (data._fallback) {
        console.log('[info] Fallback mode — skipping card.svg (requires image fetching)');
    } else {
        const card = await renderCard(data);
        outputs.push(card);
    }

    outputs.push(await renderTopArtists(data));
    outputs.push(await renderTopTracks(data));
    outputs.push(await renderWeeklyOverview(data));
    outputs.push(await renderWeeklyDuration(data));

    return outputs;
}

// ─── Git Commit Pipeline ────────────────────────────────────────────
async function commitAll(outputs) {
    try {
        const octokit = new Octokit({
            auth: GH_TOKEN,
        });

        const blobEntries = [];
        for (const output of outputs) {
            const {
                data: { sha }
            } = await octokit.git.createBlob({
                owner: AUTHOR,
                repo: REPO,
                content: output.svgBase64,
                encoding: "base64"
            });
            blobEntries.push({
                mode: '100644',
                path: output.filename,
                type: "blob",
                sha,
            });
        }

        const snapshotPath = path.resolve(__dirname, 'duration-snapshot.json');
        let snapshotContent = '{}';
        try {
            snapshotContent = fs.readFileSync(snapshotPath, 'utf8');
        } catch (e) { }

        const {
            data: { sha: snapshotSha }
        } = await octokit.git.createBlob({
            owner: AUTHOR,
            repo: REPO,
            content: snapshotContent,
            encoding: "utf-8"
        });
        blobEntries.push({
            mode: '100644',
            path: 'duration-snapshot.json',
            type: "blob",
            sha: snapshotSha,
        });

        const commits = await octokit.repos.listCommits({
            owner: AUTHOR,
            repo: REPO,
        });
        const lastSha = commits.data[0].sha;
        const {
            data: { sha: treeSHA }
        } = await octokit.git.createTree({
            owner: AUTHOR,
            repo: REPO,
            tree: blobEntries,
            base_tree: lastSha,
        });
        const {
            data: { sha: newSHA }
        } = await octokit.git.createCommit({
            owner: AUTHOR,
            repo: REPO,
            author: {
                name: "github-actions[bot]",
                email: "41898282+github-actions[bot]@users.noreply.github.com",
            },
            committer: {
                name: "github-actions[bot]",
                email: "41898282+github-actions[bot]@users.noreply.github.com",
            },
            tree: treeSHA,
            message: 'Update music cards and duration snapshot',
            parents: [lastSha],
        });
        const result = await octokit.git.updateRef({
            owner: AUTHOR,
            repo: REPO,
            ref: "heads/main",
            sha: newSHA,
        });
        console.log(`[info] Ref updated: ${result.data?.ref ?? 'unknown'} -> ${newSHA}`);
    } catch (err) {
        console.error(`上传 SVG 时发生了错误：${err}`);
    }
}

// ─── Main ───────────────────────────────────────────────────────────
(async () => {
    const cookie = `MUSIC_U=${USER_TOKEN}`;
    let data;
    try {
        data = await fetchData(cookie, USER_ID);
    } catch (err) {
        console.error(`[auth] Cookie appears invalid or expired — falling back to empty data`);
        console.error(`[auth] Detail: ${err.message || err}`);
        data = {
            _fallback: true,
            avatarUrl: '',
            nickname: '\u2013',
            songName: '\u6682\u65E0\u6570\u636E',
            songAuthors: '\u2013',
            songCover: '',
            playCount: 0,
            weekData: [],
            dailyDurations: Array(7).fill(null).map((_, i) => ({
                date: null,
                day: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'][i],
                estimatedMinutes: null,
            })),
        };
    }
    const outputs = await generateAllCards(data);
    await commitAll(outputs);
})();
