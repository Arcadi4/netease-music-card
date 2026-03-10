#!/usr/bin/env node
"use strict";
const fs = require("fs");
const path = require("path");

function checkHeightContract(svgString) {
    const svgHeightMatch = svgString.match(/<svg[^>]+height="(\d+)"/);
    const foHeightMatch = svgString.match(/<foreignObject[^>]+height="(\d+)"/);
    if (!svgHeightMatch || !foHeightMatch) {
        return { ok: false, error: "Could not parse height attributes" };
    }
    const svgH = parseInt(svgHeightMatch[1], 10);
    const foH = parseInt(foHeightMatch[1], 10);
    return {
        ok: svgH === foH,
        svgHeight: svgH,
        foreignObjectHeight: foH,
        error: svgH !== foH ? `MISMATCH: svg=${svgH} foreignObject=${foH}` : null,
    };
}

const out = [];
out.push("=== NEGATIVE CASE: Malformed SVG Height Mismatch Detection ===");
out.push("Generated: " + new Date().toISOString());
out.push("");

const malformedSvg = `<svg width="310" height="300" xmlns="http://www.w3.org/2000/svg">
    <foreignObject width="310" height="250">
        <div xmlns="http://www.w3.org/1999/xhtml">Content here</div>
    </foreignObject>
</svg>`;

const result = checkHeightContract(malformedSvg);
out.push("Test: SVG with height=300, foreignObject height=250");
out.push("  Expected: mismatch detected (ok === false)");
out.push("  Actual ok: " + result.ok);
out.push("  svgHeight: " + result.svgHeight);
out.push("  foreignObjectHeight: " + result.foreignObjectHeight);
out.push("  error: " + result.error);

if (result.ok === false && result.error && result.error.includes("MISMATCH")) {
    out.push("  RESULT: Mismatch correctly detected \u2713");
} else {
    out.push("  RESULT: FAIL - checker did not detect mismatch \u2717");
    process.exitCode = 1;
}
out.push("");

const wellFormedSvg = `<svg width="310" height="352" xmlns="http://www.w3.org/2000/svg">
    <foreignObject width="310" height="352">
        <div xmlns="http://www.w3.org/1999/xhtml">Content here</div>
    </foreignObject>
</svg>`;

const result2 = checkHeightContract(wellFormedSvg);
out.push("Test: SVG with height=352, foreignObject height=352 (well-formed)");
out.push("  Expected: no mismatch (ok === true)");
out.push("  Actual ok: " + result2.ok);
out.push("  svgHeight: " + result2.svgHeight);
out.push("  foreignObjectHeight: " + result2.foreignObjectHeight);

if (result2.ok === true) {
    out.push("  RESULT: Correctly accepted well-formed SVG \u2713");
} else {
    out.push("  RESULT: FAIL - false positive on well-formed SVG \u2717");
    process.exitCode = 1;
}
out.push("");

if (!process.exitCode) {
    out.push("=== NEGATIVE CASE VALIDATION PASSED ===");
} else {
    out.push("=== NEGATIVE CASE VALIDATION FAILED ===");
}

const report = out.join("\n") + "\n";
console.log(report);
fs.writeFileSync(path.join(__dirname, "task-7-decoded-contracts-error.txt"), report);
