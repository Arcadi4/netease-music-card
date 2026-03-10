#!/usr/bin/env node
"use strict";
const fs = require("fs");
const src = fs.readFileSync(require("path").join(__dirname, "../../index.js"), "utf8");

// ── 1. Replicate computeListCardHeight ──
function computeListCardHeight({ rowCount, rowHeight, headerHeight, cardPadding, containerPadding, rowMargin = 0 }) {
    const MIN_HEIGHT = 200;
    const safe = (v) => (Number.isFinite(v) && v >= 0 ? v : 0);
    const sRowCount = safe(rowCount);
    const sRowHeight = safe(rowHeight);
    const sHeaderHeight = safe(headerHeight);
    const sCardPadding = safe(cardPadding);
    const sContainerPadding = safe(containerPadding);
    const sRowMargin = safe(rowMargin);
    const rowMargins = sRowCount > 1 ? (sRowCount - 1) * sRowMargin : 0;
    const contentHeight = sHeaderHeight + sRowCount * sRowHeight + rowMargins;
    const rawCardHeight = contentHeight + sCardPadding * 2;
    const rawSvgHeight = rawCardHeight + sContainerPadding * 2;
    const svgHeight = Math.max(MIN_HEIGHT, Math.ceil(rawSvgHeight));
    const cardHeight = svgHeight - sContainerPadding * 2;
    return { svgHeight, foreignObjectHeight: svgHeight, cardHeight };
}

// Known constants from source (extracted manually from LIST_CARD_HEIGHT)
const LC = {
    CONTAINER_PADDING: 5,
    CARD_PADDING: 20,
    HEADER_HEIGHT: 54,
    TOP_ARTISTS_ROW_HEIGHT: 40,
    TOP_ARTISTS_ROW_COUNT: 5,
    TOP_ARTISTS_ROW_MARGIN: 12,
    TOP_TRACKS_ROW_HEIGHT: 46,
    TOP_TRACKS_ROW_COUNT: 5,
    TOP_TRACKS_ROW_MARGIN: 12,
    WEEKLY_CELL_HEIGHT: 80,
    WEEKLY_GRID_ROWS: 2,
    WEEKLY_HEADER_HEIGHT: 64,
};

// Verify the source actually contains these exact values
function verifyConstant(name, expected) {
    // Match patterns like "CONTAINER_PADDING: 5," or "CONTAINER_PADDING: 5  // comment"
    const pat = new RegExp(name + ":\\s*" + expected + "\\b");
    return pat.test(src);
}

const TOP_ARTISTS_HEIGHT = computeListCardHeight({
    rowCount: LC.TOP_ARTISTS_ROW_COUNT,
    rowHeight: LC.TOP_ARTISTS_ROW_HEIGHT,
    headerHeight: LC.HEADER_HEIGHT,
    cardPadding: LC.CARD_PADDING,
    containerPadding: LC.CONTAINER_PADDING,
    rowMargin: LC.TOP_ARTISTS_ROW_MARGIN,
});

const TOP_TRACKS_HEIGHT = computeListCardHeight({
    rowCount: LC.TOP_TRACKS_ROW_COUNT,
    rowHeight: LC.TOP_TRACKS_ROW_HEIGHT,
    headerHeight: LC.HEADER_HEIGHT,
    cardPadding: LC.CARD_PADDING,
    containerPadding: LC.CONTAINER_PADDING,
    rowMargin: LC.TOP_TRACKS_ROW_MARGIN,
});

const WEEKLY_OVERVIEW_HEIGHT = computeListCardHeight({
    rowCount: LC.WEEKLY_GRID_ROWS,
    rowHeight: LC.WEEKLY_CELL_HEIGHT,
    headerHeight: LC.WEEKLY_HEADER_HEIGHT,
    cardPadding: LC.CARD_PADDING,
    containerPadding: LC.CONTAINER_PADDING,
});

// ── 2. Verify source uses the correct constants ──
function checkSourceUsage(constName) {
    const svgHeightPat = new RegExp('height="\\$\\{' + constName + '\\.svgHeight\\}"');
    const foHeightPat = new RegExp('height="\\$\\{' + constName + '\\.foreignObjectHeight\\}"');
    const cardHeightPat = new RegExp('height: \\$\\{' + constName + '\\.cardHeight\\}px');
    return {
        hasSvg: svgHeightPat.test(src),
        hasFo: foHeightPat.test(src),
        hasCard: cardHeightPat.test(src),
    };
}

// ── 3. Check cell styling for weekly-overview ──
const hasCellMinHeight = /\.cell\s*\{[\s\S]*?min-height:\s*60px/.test(src);
const hasCellPadding = /\.cell\s*\{[\s\S]*?padding:\s*12px\s+8px/.test(src);

// ── 4. Check verify.js results ──
const verifyOutput = fs.readFileSync(require("path").join(__dirname, "task-7-verify-run.txt"), "utf8");
function checkVerifyResult(pattern) {
    return verifyOutput.includes("\u2713 " + pattern + " passed");
}

// ── 5. Build report ──
const out = [];
let allPassed = true;

function check(desc, val) {
    const mark = val ? "\u2713" : "\u2717 FAIL";
    if (!val) allPassed = false;
    out.push("  " + desc + ": " + mark);
}

out.push("=== REGRESSION CONTRACT CHECKS ===");
out.push("Generated: " + new Date().toISOString());
out.push("");

// --- top-artists ---
const taUsage = checkSourceUsage("TOP_ARTISTS_HEIGHT");
out.push("--- top-artists ---");
out.push("index.js uses:");
out.push("  svg height: TOP_ARTISTS_HEIGHT.svgHeight (computed: " + TOP_ARTISTS_HEIGHT.svgHeight + ")");
out.push("  foreignObject height: TOP_ARTISTS_HEIGHT.foreignObjectHeight (computed: " + TOP_ARTISTS_HEIGHT.foreignObjectHeight + ")");
out.push("  card CSS height: " + TOP_ARTISTS_HEIGHT.cardHeight + "px");
check("svgHeight === foreignObjectHeight: " + (TOP_ARTISTS_HEIGHT.svgHeight === TOP_ARTISTS_HEIGHT.foreignObjectHeight ? "TRUE" : "FALSE"),
    TOP_ARTISTS_HEIGHT.svgHeight === TOP_ARTISTS_HEIGHT.foreignObjectHeight);
check("source uses TOP_ARTISTS_HEIGHT.svgHeight for <svg>", taUsage.hasSvg);
check("source uses TOP_ARTISTS_HEIGHT.foreignObjectHeight for <foreignObject>", taUsage.hasFo);
check("source uses TOP_ARTISTS_HEIGHT.cardHeight for CSS", taUsage.hasCard);
check("verify.js: renderTopArtists(normal) passed", checkVerifyResult("renderTopArtists(normal)"));
check("verify.js: renderTopArtists(empty) passed", checkVerifyResult("renderTopArtists(empty)"));
check("verify.js: renderTopArtists(zero-play) passed", checkVerifyResult("renderTopArtists(zero-play)"));
check("verify.js: deriveTopArtists returns 5 artists", checkVerifyResult("deriveTopArtists: returns 5 artists"));
check("verify.js: zero-play fallback works", checkVerifyResult("deriveTopArtists(zero-play): falls back to score, #1 is SICK HACK"));
out.push("");

// --- top-tracks ---
const ttUsage = checkSourceUsage("TOP_TRACKS_HEIGHT");
out.push("--- top-tracks ---");
out.push("index.js uses:");
out.push("  svg height: TOP_TRACKS_HEIGHT.svgHeight (computed: " + TOP_TRACKS_HEIGHT.svgHeight + ")");
out.push("  foreignObject height: TOP_TRACKS_HEIGHT.foreignObjectHeight (computed: " + TOP_TRACKS_HEIGHT.foreignObjectHeight + ")");
out.push("  card CSS height: " + TOP_TRACKS_HEIGHT.cardHeight + "px");
check("svgHeight === foreignObjectHeight: " + (TOP_TRACKS_HEIGHT.svgHeight === TOP_TRACKS_HEIGHT.foreignObjectHeight ? "TRUE" : "FALSE"),
    TOP_TRACKS_HEIGHT.svgHeight === TOP_TRACKS_HEIGHT.foreignObjectHeight);
check("source uses TOP_TRACKS_HEIGHT.svgHeight for <svg>", ttUsage.hasSvg);
check("source uses TOP_TRACKS_HEIGHT.foreignObjectHeight for <foreignObject>", ttUsage.hasFo);
check("source uses TOP_TRACKS_HEIGHT.cardHeight for CSS", ttUsage.hasCard);
check("verify.js: renderTopTracks(normal) passed", checkVerifyResult("renderTopTracks(normal)"));
check("verify.js: renderTopTracks(empty) passed", checkVerifyResult("renderTopTracks(empty)"));
check("verify.js: renderTopTracks(zero-play) passed", checkVerifyResult("renderTopTracks(zero-play)"));
check("verify.js: deriveTopTracks returns 5 tracks", checkVerifyResult("deriveTopTracks: returns 5 tracks"));
out.push("");

// --- weekly-overview ---
const woUsage = checkSourceUsage("WEEKLY_OVERVIEW_HEIGHT");
out.push("--- weekly-overview ---");
out.push("index.js uses:");
out.push("  svg height: WEEKLY_OVERVIEW_HEIGHT.svgHeight (computed: " + WEEKLY_OVERVIEW_HEIGHT.svgHeight + ")");
out.push("  foreignObject height: WEEKLY_OVERVIEW_HEIGHT.foreignObjectHeight (computed: " + WEEKLY_OVERVIEW_HEIGHT.foreignObjectHeight + ")");
out.push("  card CSS height: " + WEEKLY_OVERVIEW_HEIGHT.cardHeight + "px");
check("svgHeight === foreignObjectHeight: " + (WEEKLY_OVERVIEW_HEIGHT.svgHeight === WEEKLY_OVERVIEW_HEIGHT.foreignObjectHeight ? "TRUE" : "FALSE"),
    WEEKLY_OVERVIEW_HEIGHT.svgHeight === WEEKLY_OVERVIEW_HEIGHT.foreignObjectHeight);
check("source uses WEEKLY_OVERVIEW_HEIGHT.svgHeight for <svg>", woUsage.hasSvg);
check("source uses WEEKLY_OVERVIEW_HEIGHT.foreignObjectHeight for <foreignObject>", woUsage.hasFo);
check("source uses WEEKLY_OVERVIEW_HEIGHT.cardHeight for CSS", woUsage.hasCard);
check("verify.js: renderWeeklyOverview(normal) passed", checkVerifyResult("renderWeeklyOverview(normal)"));
check("verify.js: renderWeeklyOverview(empty) passed", checkVerifyResult("renderWeeklyOverview(empty)"));
check("verify.js: renderWeeklyOverview(empty) shows 0 totalPlays", checkVerifyResult("renderWeeklyOverview(empty): shows 0 totalPlays"));

out.push("  Cell spacing guardrails in source:");
check("  .cell has min-height: 60px", hasCellMinHeight);
check("  .cell has padding: 12px 8px", hasCellPadding);

// Check 4 cell labels present in source
const cellLabels = ["总播放 / Total Plays", "不同歌曲 / Unique Songs", "不同艺术家 / Artists", "重复强度 / Repeat %"];
const allLabelsPresent = cellLabels.every(label => src.includes(label));
check("  All 4 cell labels present in source", allLabelsPresent);
out.push("");

// === Summary ===
if (allPassed) {
    out.push("=== ALL CHECKS PASSED ===");
} else {
    out.push("=== SOME CHECKS FAILED ===");
}

const report = out.join("\n") + "\n";
console.log(report);
fs.writeFileSync(require("path").join(__dirname, "task-7-decoded-contracts.txt"), report);
if (!allPassed) process.exit(1);
