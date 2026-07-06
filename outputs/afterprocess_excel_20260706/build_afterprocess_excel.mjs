import fs from "node:fs/promises";
import path from "node:path";
import { SpreadsheetFile, Workbook } from "@oai/artifact-tool";

const repoRoot = path.resolve("../..");
const sourcePath = path.join(repoRoot, "mock", "AfterProcess_20260630_1700 1.json");
const outputDir = path.join(repoRoot, "outputs", "afterprocess_excel_20260706");
const outputPath = path.join(outputDir, "AfterProcess_20260630_1700.xlsx");
const previewPath = path.join(outputDir, "AfterProcess_20260630_1700_preview.png");

function columnName(index) {
  let n = index + 1;
  let name = "";
  while (n > 0) {
    const rem = (n - 1) % 26;
    name = String.fromCharCode(65 + rem) + name;
    n = Math.floor((n - 1) / 26);
  }
  return name;
}

function normalizeRows(headers, rows) {
  return rows.map((row) => {
    const normalized = Array.isArray(row) ? row.slice(0, headers.length) : [];
    while (normalized.length < headers.length) normalized.push("");
    return normalized.map((value) => (value == null ? "" : String(value)));
  });
}

const raw = await fs.readFile(sourcePath, "utf8");
const parsed = JSON.parse(raw);
const headers = parsed.headers.map((value) => String(value));
const rows = normalizeRows(headers, parsed.rows);

if (!headers.length) {
  throw new Error("No headers found in source JSON.");
}
if (!rows.length) {
  throw new Error("No rows found in source JSON.");
}

const workbook = Workbook.create();
const sheet = workbook.worksheets.add("AfterProcess");
sheet.showGridLines = false;

const lastCol = columnName(headers.length - 1);
const headerRow = 4;
const dataStartRow = headerRow + 1;
const dataEndRow = dataStartRow + rows.length - 1;
const tableRange = `A${headerRow}:${lastCol}${dataEndRow}`;

sheet.getRange(`A1:${lastCol}1`).merge();
sheet.getRange("A1").values = [["AfterProcess 20260630 1700"]];
sheet.getRange("A1").format = {
  fill: "#164E63",
  font: { bold: true, color: "#FFFFFF", size: 15 },
  horizontalAlignment: "left",
  verticalAlignment: "center",
};

sheet.getRange(`A2:${lastCol}2`).merge();
sheet.getRange("A2").values = [[`Source: ${path.relative(repoRoot, sourcePath)} | Rows: ${rows.length}`]];
sheet.getRange("A2").format = {
  fill: "#E0F2FE",
  font: { color: "#0F172A", size: 10 },
  horizontalAlignment: "left",
  verticalAlignment: "center",
};

sheet.getRange(`A${headerRow}:${lastCol}${headerRow}`).values = [headers];
sheet.getRange(`A${dataStartRow}:${lastCol}${dataEndRow}`).values = rows;

const table = sheet.tables.add(tableRange, true, "AfterProcessTable");
table.style = "TableStyleMedium4";
table.showFilterButton = true;
table.showBandedRows = true;

sheet.getRange(`A${headerRow}:${lastCol}${headerRow}`).format = {
  fill: "#0F766E",
  font: { bold: true, color: "#FFFFFF" },
  horizontalAlignment: "center",
  verticalAlignment: "center",
  wrapText: true,
};

sheet.getRange(`A${dataStartRow}:${lastCol}${dataEndRow}`).format = {
  font: { color: "#111827" },
  horizontalAlignment: "center",
  verticalAlignment: "center",
  wrapText: false,
  numberFormat: "@",
};

sheet.getRange(`A${dataStartRow}:A${dataEndRow}`).format.horizontalAlignment = "left";
sheet.getRange(`B${dataStartRow}:B${dataEndRow}`).format.numberFormat = "00";
sheet.getRange(`V${dataStartRow}:V${dataEndRow}`).format.horizontalAlignment = "left";
sheet.getRange(`A${headerRow}:${lastCol}${dataEndRow}`).format.borders = {
  preset: "inside",
  style: "thin",
  color: "#CBD5E1",
};
sheet.getRange(`A${headerRow}:${lastCol}${dataEndRow}`).format.borders = {
  preset: "outside",
  style: "medium",
  color: "#94A3B8",
};

const widths = [
  13, 7, 10, 10, 10, 10, 9, 13, 10, 10, 15, 13, 11, 13, 8, 8, 8, 8, 8, 8, 12, 22,
];
for (let i = 0; i < headers.length; i += 1) {
  const col = columnName(i);
  sheet.getRange(`${col}${headerRow}:${col}${dataEndRow}`).format.columnWidth = widths[i] ?? 12;
}

sheet.getRange("A1").format.rowHeight = 26;
sheet.getRange("A2").format.rowHeight = 22;
sheet.getRange(`A${headerRow}:${lastCol}${headerRow}`).format.rowHeight = 38;
sheet.getRange(`A${dataStartRow}:${lastCol}${dataEndRow}`).format.rowHeight = 21;
sheet.freezePanes.freezeRows(headerRow);

const check = await workbook.inspect({
  kind: "table",
  range: tableRange,
  include: "values,formulas",
  tableMaxRows: 6,
  tableMaxCols: 8,
  maxChars: 4000,
});
console.log(check.ndjson);

const errors = await workbook.inspect({
  kind: "match",
  searchTerm: "#REF!|#DIV/0!|#VALUE!|#NAME\\?|#N/A",
  options: { useRegex: true, maxResults: 300 },
  summary: "final formula error scan",
});
console.log(errors.ndjson);

const preview = await workbook.render({
  sheetName: "AfterProcess",
  autoCrop: "all",
  scale: 1,
  format: "png",
});
await fs.writeFile(previewPath, new Uint8Array(await preview.arrayBuffer()));

const output = await SpreadsheetFile.exportXlsx(workbook);
await output.save(outputPath);

console.log(JSON.stringify({ outputPath, previewPath, rows: rows.length, columns: headers.length }, null, 2));
