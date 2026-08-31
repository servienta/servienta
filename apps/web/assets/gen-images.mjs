// Rasterize brand SVGs to PNG at build time (committed as static assets).
import { Resvg } from "@resvg/resvg-js";
import { readFileSync, writeFileSync } from "node:fs";

function render(svgPath, outPath, width) {
  const svg = readFileSync(svgPath, "utf8");
  const r = new Resvg(svg, {
    fitTo: { mode: "width", value: width },
    font: { loadSystemFonts: true },
  });
  writeFileSync(outPath, r.render().asPng());
  console.log(`${outPath} (${width}px)`);
}

// Social card at full 1200×630.
render("assets/og.svg", "public/og.png", 1200);
// Favicon PNGs from the mark.
render("public/favicon.svg", "public/favicon-32.png", 32);
render("public/favicon.svg", "public/apple-touch-icon.png", 180);
render("public/favicon.svg", "public/favicon-512.png", 512);
