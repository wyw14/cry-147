const page = document.body.dataset.page;
const endpoint = `/api/${page}`;
const table = document.querySelector("tbody");
const status = document.querySelector(".status");

const columns = {
  operations: ["run_id", "tray_id", "stage", "message", "updated_at"],
  equipment: ["channel", "step", "set_current", "set_voltage", "protected", "updated_at"],
  interlocks: ["run_id", "channel", "reason", "latched", "updated_at"],
  incidents: ["run_id", "tray_id", "summary", "severity", "open"]
};

function valueCell(key, value) {
  if (key === "stage" || key === "severity" || key === "step") {
    return `<span class="pill ${value === "critical" ? "critical" : ""}">${value ?? "-"}</span>`;
  }
  if (typeof value === "boolean") {
    return `<span class="pill ${value ? "critical" : "ok"}">${value ? "yes" : "no"}</span>`;
  }
  if (key.endsWith("_at") && value) return new Date(value).toLocaleString();
  return value ?? "-";
}

function render(rows) {
  table.innerHTML = "";
  if (!rows.length) {
    table.innerHTML = `<tr><td class="empty" colspan="${columns[page].length}">No active records</td></tr>`;
    return;
  }
  for (const row of rows) {
    const tr = document.createElement("tr");
    tr.innerHTML = columns[page].map(key => `<td>${valueCell(key, row[key])}</td>`).join("");
    table.appendChild(tr);
  }
}

async function refresh() {
  status.textContent = "Refreshing";
  try {
    const response = await fetch(endpoint, { headers: { Accept: "application/json" } });
    if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
    const rows = await response.json();
    render(rows);
    status.textContent = `${rows.length} records · ${new Date().toLocaleTimeString()}`;
    const count = document.querySelector("[data-count]");
    if (count) count.textContent = rows.length;
  } catch (error) {
    status.textContent = error.message;
  }
}

document.querySelector("[data-refresh]")?.addEventListener("click", refresh);
refresh();
setInterval(refresh, 15000);
