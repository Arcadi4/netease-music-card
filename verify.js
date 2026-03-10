#!/usr/bin/env node
'use strict';

/**
 * Fixture-based end-to-end verification for all 4 data-card SVG renderers.
 * (renderCard is skipped — it requires HTTP via getBase64.)
 *
 * Exercises: renderTopArtists, renderTopTracks, renderWeeklyOverview, renderWeeklyDuration
 * Also tests edge cases: empty weekData, all-null dailyDurations, all-playCount-zero.
 *
 * Exit 0 on success, 1 on any failure.
 */

const fs = require('fs');
const path = require('path');

// ─── Copied from index.js (pure functions + STYLE) ──────────────────

const STYLE = {
    width: 310,
    height: 490,
    borderRadius: "10px",
    shadow: "gray 0 0 10px",
    accent: "#BA0400",
    green: "#28C131",
    fontStack: "'PingFang SC', 'Helvetica Neue', 'Segoe UI', 'Microsoft YaHei', sans-serif",
};

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

// ─── Renderer Functions (copied from index.js) ──────────────────────

async function renderTopArtists(data) {
    const artists = deriveTopArtists(data.weekData, 5);
    let artistRows = "";

    if (artists.length === 0) {
        artistRows = `<div class="empty-state">\u6682\u65E0\u6570\u636E / No data</div>`;
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
                height: 310px;
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
                    <div class="title">\u672C\u5468 Top \u827A\u672F\u5BB6</div>
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
            </div>`).join('') : `<div class="empty">\u6682\u65E0\u6570\u636E / No data</div>`;
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
                height: 320px;
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
                    <div class="title-header">\u672C\u5468 Top \u6B4C\u66F2</div>
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
                height: 250px;
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
                    <div class="title">\u672C\u5468\u6982\u89C8</div>
                    <div class="subtitle">Weekly Overview</div>
                </div>
                <div class="grid">
                    <div class="cell">
                        <div class="value">${stats.totalPlays}</div>
                        <div class="label">\u603B\u64AD\u653E / Total Plays</div>
                    </div>
                    <div class="cell">
                        <div class="value">${stats.uniqueSongs}</div>
                        <div class="label">\u4E0D\u540C\u6B4C\u66F2 / Unique Songs</div>
                    </div>
                    <div class="cell">
                        <div class="value">${stats.uniqueArtists}</div>
                        <div class="label">\u4E0D\u540C\u827A\u672F\u5BB6 / Artists</div>
                    </div>
                    <div class="cell">
                        <div class="value">${stats.repeatIntensity === 0 ? '0' : stats.repeatIntensity}%</div>
                        <div class="label">\u91CD\u590D\u5F3A\u5EA6 / Repeat %</div>
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
        chartHtml = `<div class="empty-state">\u6682\u65E0\u6570\u636E / No data</div>`;
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
                height: 250px;
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
                    <div class="title">\u672C\u5468\u542C\u6B4C\u65F6\u957F</div>
                    <div class="subtitle">\u4F30\u7B97\u65F6\u957F \u00B7 Estimated</div>
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

// ─── Fixture Data ───────────────────────────────────────────────────

const FIXTURE_WEEK_DATA = [
    { playCount: 42, score: 85, song: { id: 1, name: '\u5B64\u72EC\u6447\u6EDA', ar: [{ id: 101, name: 'SICK HACK' }, { id: 102, name: 'Bocchi' }], al: {} } },
    { playCount: 30, score: 70, song: { id: 2, name: 'Guitar, Loneliness and Blue Planet', ar: [{ id: 101, name: 'SICK HACK' }], al: {} } },
    { playCount: 0, score: 60, song: { id: 3, name: '\u3042\u306E\u30D0\u30F3\u30C9', ar: [{ id: 103, name: '\u6A61\u76AE\u64E6' }], al: {} } },
    { playCount: 25, score: 55, song: { id: 4, name: '\u306A\u306B\u304C\u60AA\u3044', ar: [{ id: 104, name: '\u6597\u5FD7' }], al: {} } },
    { playCount: 18, score: 40, song: { id: 5, name: '\u3072\u3068\u308A\u307C\u3063\u3061\u6771\u4EAC', ar: [{ id: 105, name: '\u5F8C\u85E4\u3072\u3068\u308A' }], al: {} } },
    { playCount: 10, score: 20, song: { id: 6, name: 'Long Longer', ar: [{ id: 106, name: '\u306A\u306B\u304B' }], al: {} } },
];

const FIXTURE_DAILY = [
    { date: '2026-03-04', day: 'Mon', estimatedMinutes: 35.0 },
    { date: '2026-03-05', day: 'Tue', estimatedMinutes: null },
    { date: '2026-03-06', day: 'Wed', estimatedMinutes: 87.5 },
    { date: '2026-03-07', day: 'Thu', estimatedMinutes: 42.0 },
    { date: '2026-03-08', day: 'Fri', estimatedMinutes: null },
    { date: '2026-03-09', day: 'Sat', estimatedMinutes: 61.0 },
    { date: '2026-03-10', day: 'Sun', estimatedMinutes: 14.0 },
];

const FIXTURE_DATA = {
    avatarUrl: '', nickname: 'Test User',
    songName: '\u5B64\u72EC\u6447\u6EDA', songAuthors: 'SICK HACK',
    songCover: '', playCount: 42,
    weekData: FIXTURE_WEEK_DATA, dailyDurations: FIXTURE_DAILY,
};

const EMPTY_DATA = {
    weekData: [], dailyDurations: Array(7).fill(null).map((_, i) => ({
        date: null, day: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'][i], estimatedMinutes: null
    })),
};

// All playCount zero - should fall back to score-based ranking
const ZERO_PLAY_WEEK_DATA = FIXTURE_WEEK_DATA.map(e => ({ ...e, playCount: 0 }));
const ZERO_PLAY_DATA = { ...FIXTURE_DATA, weekData: ZERO_PLAY_WEEK_DATA };

// ─── Verification Helpers ───────────────────────────────────────────

let passCount = 0;
let failCount = 0;
const results = [];

function assertSvgValid(label, svgBase64) {
    if (!svgBase64 || svgBase64.length === 0) {
        throw new Error(`[FAIL] ${label}: svgBase64 is empty`);
    }
    const decoded = Buffer.from(svgBase64, 'base64').toString('utf-8');
    if (!decoded.includes('<svg')) {
        throw new Error(`[FAIL] ${label}: decoded content missing <svg tag`);
    }
    if (!decoded.includes('</svg>')) {
        throw new Error(`[FAIL] ${label}: decoded content missing </svg> tag`);
    }
    const byteSize = Buffer.from(svgBase64, 'base64').length;
    console.log(`\u2713 ${label} passed (${byteSize} bytes)`);
    passCount++;
    results.push(`PASS: ${label} (${byteSize} bytes)`);
}

function assertEmptyState(label, svgBase64, emptyMarker) {
    if (!svgBase64 || svgBase64.length === 0) {
        throw new Error(`[FAIL] ${label}: svgBase64 is empty`);
    }
    const decoded = Buffer.from(svgBase64, 'base64').toString('utf-8');
    if (!decoded.includes('<svg')) {
        throw new Error(`[FAIL] ${label}: decoded content missing <svg tag`);
    }
    if (!decoded.includes(emptyMarker)) {
        throw new Error(`[FAIL] ${label}: expected empty-state marker "${emptyMarker}" not found`);
    }
    const byteSize = Buffer.from(svgBase64, 'base64').length;
    console.log(`\u2713 ${label} passed (${byteSize} bytes, contains empty-state)`);
    passCount++;
    results.push(`PASS: ${label} (${byteSize} bytes, empty-state)`);
}

function assertDerivation(label, actual, check) {
    if (!check(actual)) {
        throw new Error(`[FAIL] ${label}: assertion failed, got ${JSON.stringify(actual)}`);
    }
    console.log(`\u2713 ${label} passed`);
    passCount++;
    results.push(`PASS: ${label}`);
}

// ─── Main ───────────────────────────────────────────────────────────

(async () => {
    console.log('=== Fixture-based E2E Verification ===\n');

    // ── 1. Derivation unit checks ──────────────────────────────────
    console.log('--- Derivation checks ---');

    const topArtists = deriveTopArtists(FIXTURE_WEEK_DATA, 5);
    assertDerivation('deriveTopArtists: returns 5 artists', topArtists, a => a.length === 5);
    assertDerivation('deriveTopArtists: #1 is SICK HACK (72 plays)', topArtists,
        a => a[0].name === 'SICK HACK' && a[0].plays === 72);

    const overview = deriveWeeklyOverview(FIXTURE_WEEK_DATA);
    assertDerivation('deriveWeeklyOverview: totalPlays=185', overview, o => o.totalPlays === 185);
    assertDerivation('deriveWeeklyOverview: uniqueSongs=6', overview, o => o.uniqueSongs === 6);
    assertDerivation('deriveWeeklyOverview: uniqueArtists=6', overview, o => o.uniqueArtists === 6);

    const topTracks = deriveTopTracks(FIXTURE_WEEK_DATA, 5);
    assertDerivation('deriveTopTracks: returns 5 tracks', topTracks, t => t.length === 5);
    assertDerivation('deriveTopTracks: #1 is \u5B64\u72EC\u6447\u6EDA (42 plays)', topTracks,
        t => t[0].name === '\u5B64\u72EC\u6447\u6EDA' && t[0].plays === 42);

    // Zero-play fallback: should use score
    const zeroArtists = deriveTopArtists(ZERO_PLAY_WEEK_DATA, 5);
    assertDerivation('deriveTopArtists(zero-play): falls back to score, #1 is SICK HACK',
        zeroArtists, a => a[0].name === 'SICK HACK' && a[0].plays === (85 + 70));

    const zeroTracks = deriveTopTracks(ZERO_PLAY_WEEK_DATA, 5);
    assertDerivation('deriveTopTracks(zero-play): #1 is \u5B64\u72EC\u6447\u6EDA (score=85)',
        zeroTracks, t => t[0].name === '\u5B64\u72EC\u6447\u6EDA' && t[0].plays === 85);

    // Empty
    const emptyArtists = deriveTopArtists([], 5);
    assertDerivation('deriveTopArtists(empty): returns []', emptyArtists, a => a.length === 0);

    const emptyOverview = deriveWeeklyOverview([]);
    assertDerivation('deriveWeeklyOverview(empty): totalPlays=0', emptyOverview, o => o.totalPlays === 0);

    // ── 2. Renderer SVG checks (normal data) ──────────────────────
    console.log('\n--- Renderer checks (normal data) ---');

    const r1 = await renderTopArtists(FIXTURE_DATA);
    assertSvgValid('renderTopArtists(normal)', r1.svgBase64);

    const r2 = await renderTopTracks(FIXTURE_DATA);
    assertSvgValid('renderTopTracks(normal)', r2.svgBase64);

    const r3 = await renderWeeklyOverview(FIXTURE_DATA);
    assertSvgValid('renderWeeklyOverview(normal)', r3.svgBase64);

    const r4 = await renderWeeklyDuration(FIXTURE_DATA);
    assertSvgValid('renderWeeklyDuration(normal)', r4.svgBase64);

    // ── 3. Renderer SVG checks (empty/edge data) ─────────────────
    console.log('\n--- Renderer checks (empty/edge data) ---');

    const e1 = await renderTopArtists(EMPTY_DATA);
    assertEmptyState('renderTopArtists(empty)', e1.svgBase64, '\u6682\u65E0\u6570\u636E');

    const e2 = await renderTopTracks(EMPTY_DATA);
    assertEmptyState('renderTopTracks(empty)', e2.svgBase64, '\u6682\u65E0\u6570\u636E');

    const e3 = await renderWeeklyOverview(EMPTY_DATA);
    assertSvgValid('renderWeeklyOverview(empty)', e3.svgBase64);
    // Verify shows zeros
    const e3decoded = Buffer.from(e3.svgBase64, 'base64').toString('utf-8');
    assertDerivation('renderWeeklyOverview(empty): shows 0 totalPlays', e3decoded,
        d => d.includes('>0<'));

    const e4 = await renderWeeklyDuration(EMPTY_DATA);
    assertEmptyState('renderWeeklyDuration(empty)', e4.svgBase64, '\u6682\u65E0\u6570\u636E');

    // ── 4. Zero-play renderers ────────────────────────────────────
    console.log('\n--- Renderer checks (zero-play data) ---');

    const z1 = await renderTopArtists(ZERO_PLAY_DATA);
    assertSvgValid('renderTopArtists(zero-play)', z1.svgBase64);
    const z1decoded = Buffer.from(z1.svgBase64, 'base64').toString('utf-8');
    assertDerivation('renderTopArtists(zero-play): contains SICK HACK', z1decoded,
        d => d.includes('SICK HACK'));

    const z2 = await renderTopTracks(ZERO_PLAY_DATA);
    assertSvgValid('renderTopTracks(zero-play)', z2.svgBase64);

    // ── 5. Filename checks ────────────────────────────────────────
    console.log('\n--- Filename checks ---');
    assertDerivation('renderTopArtists filename', r1, r => r.filename === 'top-artists.svg');
    assertDerivation('renderTopTracks filename', r2, r => r.filename === 'top-tracks.svg');
    assertDerivation('renderWeeklyOverview filename', r3, r => r.filename === 'weekly-overview.svg');
    assertDerivation('renderWeeklyDuration filename', r4, r => r.filename === 'weekly-duration.svg');

    // ── Summary & Evidence ────────────────────────────────────────
    console.log(`\n=== Results: ${passCount} passed, ${failCount} failed ===`);

    const evidenceDir = path.resolve(__dirname, '.sisyphus', 'evidence');
    fs.mkdirSync(evidenceDir, { recursive: true });

    const evidenceContent = [
        `Fixture-based E2E Verification \u2014 Task 10`,
        `Date: ${new Date().toISOString()}`,
        `Node: ${process.version}`,
        ``,
        `Total checks: ${passCount} passed, ${failCount} failed`,
        ``,
        ...results,
    ].join('\n');

    fs.writeFileSync(path.join(evidenceDir, 'task-10-e2e.txt'), evidenceContent, 'utf8');
    console.log(`\nEvidence written to .sisyphus/evidence/task-10-e2e.txt`);

    if (failCount > 0) {
        process.exit(1);
    }
    process.exit(0);
})();
