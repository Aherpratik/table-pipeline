import { useEffect, useMemo, useState } from "react";
import {
  RefreshCw,
  Database,
  Filter,
  Search,
  Table2
} from "lucide-react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  LineChart,
  Line,
} from "recharts";

type ColumnInfo = {
  name: string;
  type: string;
};

type DatasetResponse = {
  table_name: string;
  columns: ColumnInfo[];
  rows: Record<string, any>[];
  row_count: number;
  numeric_columns: string[];
  text_columns: string[];
};

function panelStyle(): React.CSSProperties {
  return {
    background: "#ffffff",
    border: "1px solid #e2e8f0",
    borderRadius: 18,
    boxShadow: "0 8px 24px rgba(15, 23, 42, 0.06)",
  };
}

function inputStyle(withIcon = false): React.CSSProperties {
  return {
    width: "100%",
    background: "#f8fafc",
    border: "1px solid #e2e8f0",
    borderRadius: 12,
    padding: withIcon ? "12px 14px 12px 38px" : "12px 14px",
    fontSize: 14,
    outline: "none",
    color: "#0f172a",
    boxSizing: "border-box",
  };
}

function MetricCard({
  label,
  value,
  sublabel,
}: {
  label: string;
  value: string | number;
  sublabel: string;
}) {
  return (
    <div style={panelStyle()}>
      <div style={{ padding: 20 }}>
        <div style={{ fontSize: 14, fontWeight: 600, color: "#64748b" }}>{label}</div>
        <div style={{ marginTop: 10, fontSize: 36, fontWeight: 700, color: "#0f172a" }}>{value}</div>
        <div
          style={{
            marginTop: 4,
            fontSize: 11,
            textTransform: "uppercase",
            letterSpacing: "0.08em",
            color: "#94a3b8",
          }}
        >
          {sublabel}
        </div>
      </div>
    </div>
  );
}

function Panel({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle: string;
  children: React.ReactNode;
}) {
  return (
    <div style={panelStyle()}>
      <div style={{ padding: "20px 24px", borderBottom: "1px solid #e2e8f0" }}>
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700, color: "#0f172a" }}>{title}</h2>
        <p style={{ marginTop: 6, fontSize: 14, color: "#64748b" }}>{subtitle}</p>
      </div>
      <div style={{ padding: 24 }}>{children}</div>
    </div>
  );
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div
      style={{
        border: "1px solid #e2e8f0",
        background: "#f8fafc",
        borderRadius: 12,
        padding: "14px 16px",
      }}
    >
      <div
        style={{
          fontSize: 11,
          fontWeight: 700,
          textTransform: "uppercase",
          letterSpacing: "0.12em",
          color: "#94a3b8",
        }}
      >
        {label}
      </div>
      <div style={{ marginTop: 8, fontSize: 14, color: "#334155", wordBreak: "break-word" }}>{value}</div>
    </div>
  );
}

export default function App() {
  const [dataset, setDataset] = useState<DatasetResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState("Loading dataset...");
  const [search, setSearch] = useState("");
  const [chartX, setChartX] = useState("");
  const [chartY, setChartY] = useState("");

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await fetch("http://localhost:8080/api/dataset");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data: DatasetResponse = await res.json();

      setDataset(data);
      setStatus(`Connected to ${data.table_name} · ${data.row_count} rows loaded`);

      const preferredX =
        data.text_columns.includes("railway_station")
          ? "railway_station"
          : data.text_columns[0] || "";

      const preferredY = data.numeric_columns[0] || "";

      setChartX(preferredX);
      setChartY(preferredY);
    } catch (err) {
      console.error(err);
      setStatus("Failed to load dataset");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const filteredRows = useMemo(() => {
    if (!dataset) return [];
    if (!search) return dataset.rows;

    return dataset.rows.filter((row) =>
      Object.values(row).some((value) =>
        String(value ?? "")
          .toLowerCase()
          .includes(search.toLowerCase())
      )
    );
  }, [dataset, search]);

  const stats = useMemo(() => {
    if (!dataset) {
      return { rows: 0, columns: 0, numeric: 0, text: 0 };
    }
    return {
      rows: dataset.row_count,
      columns: dataset.columns.length,
      numeric: dataset.numeric_columns.length,
      text: dataset.text_columns.length,
    };
  }, [dataset]);

  const chartData = useMemo(() => {
    if (!dataset || !chartX || !chartY) return [];

    return filteredRows
      .filter((row) => row[chartX] != null && row[chartY] != null)
      .map((row) => ({
        label: String(row[chartX]),
        value: Number(row[chartY]),
      }))
      .filter((r) => Number.isFinite(r.value))
      .sort((a, b) => b.value - a.value)
      .slice(0, 8);
  }, [dataset, filteredRows, chartX, chartY]);

  const trendData = useMemo(() => {
    return chartData.map((item, index) => ({
      name: `${index + 1}`,
      value: item.value,
    }));
  }, [chartData]);

  const previewColumns = useMemo(() => {
    if (!dataset) return [];
    const preferred = [
      "country",
      "city",
      "railway_station",
      "passengers_millions_per_year",
      "platforms",
      "sourced_data_year",
    ];
    const ordered = preferred.filter((p) => dataset.columns.some((c) => c.name === p));
    const rest = dataset.columns.map((c) => c.name).filter((name) => !ordered.includes(name) && name !== "row_hash");
    return [...ordered, ...rest].slice(0, 6);
  }, [dataset]);

  const previewRows = filteredRows.slice(0, 12);

  return (
    <div
      style={{
        minHeight: "100vh",
        background: "#f4f7fb",
        color: "#0f172a",
        fontFamily: "Inter, Arial, sans-serif",
      }}
    >
      <div style={{ display: "grid", gridTemplateColumns: "88px 1fr", minHeight: "100vh" }}>
        <aside
          style={{
            background: "#123f91",
            color: "white",
            padding: "18px 10px",
            borderRight: "1px solid rgba(255,255,255,0.08)",
          }}
        >
          <div style={{ display: "flex", flexDirection: "column", alignItems: "center", height: "100%" }}>
            <div
              style={{
                width: 46,
                height: 46,
                borderRadius: 14,
                background: "white",
                color: "#123f91",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontWeight: 800,
                marginBottom: 18,
              }}
            >
              ▦
            </div>

            {["Dashboard", "Datasets", "Streams", "Visuals", "Schema", "Monitoring", "Settings"].map(
              (item, index) => (
                <div
                  key={item}
                  style={{
                    width: "100%",
                    textAlign: "center",
                    padding: "12px 6px",
                    borderRadius: 12,
                    marginBottom: 8,
                    fontSize: 11,
                    fontWeight: 700,
                    background: index === 0 ? "white" : "transparent",
                    color: index === 0 ? "#123f91" : "#dbeafe",
                    boxShadow: index === 0 ? "0 4px 12px rgba(0,0,0,0.12)" : "none",
                  }}
                >
                  {item}
                </div>
              )
            )}
          </div>
        </aside>

        <main>
          <header
            style={{
              background: "white",
              borderBottom: "1px solid #e2e8f0",
              padding: "24px 32px",
            }}
          >
            <div style={{ display: "flex", justifyContent: "space-between", gap: 16, alignItems: "center", flexWrap: "wrap" }}>
              <div>
                <div
                  style={{
                    fontSize: 12,
                    fontWeight: 700,
                    letterSpacing: "0.18em",
                    textTransform: "uppercase",
                    color: "#94a3b8",
                  }}
                >
                  Control Panel
                </div>
                <h1 style={{ margin: "6px 0 0 0", fontSize: 38, fontWeight: 700, color: "#0f172a" }}>
                  Adaptive Dataset Dashboard
                </h1>
                <p style={{ marginTop: 8, color: "#64748b", fontSize: 14 }}>{status}</p>
              </div>

              <div style={{ display: "flex", gap: 12, alignItems: "center", flexWrap: "wrap" }}>
                <div
                  style={{
                    display: "inline-flex",
                    alignItems: "center",
                    gap: 8,
                    background: "#f8fafc",
                    border: "1px solid #e2e8f0",
                    borderRadius: 12,
                    padding: "10px 14px",
                    color: "#475569",
                    fontSize: 14,
                  }}
                >
                  <Database size={16} />
                  {dataset?.table_name ?? "No dataset"}
                </div>

                <button
                  onClick={loadData}
                  style={{
                    display: "inline-flex",
                    alignItems: "center",
                    gap: 8,
                    background: "#2563eb",
                    color: "white",
                    border: "none",
                    borderRadius: 12,
                    padding: "11px 16px",
                    fontSize: 14,
                    fontWeight: 600,
                    cursor: "pointer",
                  }}
                >
                  <RefreshCw size={16} />
                  {loading ? "Refreshing..." : "Refresh Data"}
                </button>
              </div>
            </div>
          </header>

          <div style={{ padding: 32 }}>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "minmax(0, 1fr) 300px",
                gap: 24,
                alignItems: "start",
              }}
              >
              
              <section
                style={{
                  display: "flex",
                  flexDirection: "column",
                  gap: 24,
                  
                  minWidth: 0,
                }}
              >
                <div style={{ display: "grid", gridTemplateColumns: "repeat(4, minmax(160px, 1fr))", gap: 16 }}>
                  <MetricCard label="Rows Loaded" value={stats.rows} sublabel="Records" />
                  <MetricCard label="Columns" value={stats.columns} sublabel="Detected schema" />
                  <MetricCard label="Numeric Fields" value={stats.numeric} sublabel="Chart-ready" />
                  <MetricCard label="Text Fields" value={stats.text} sublabel="Category-ready" />
                </div>

                <div style={panelStyle()}>
                  <div style={{ padding: "20px 24px", borderBottom: "1px solid #e2e8f0" }}>
                    <div style={{ display: "flex", justifyContent: "space-between", gap: 16, alignItems: "center" }}>
                      <div>
                        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700, color: "#0f172a" }}>Dataset Explorer</h2>
                        <p style={{ marginTop: 6, fontSize: 14, color: "#64748b" }}>
                          
                        </p>
                      </div>
                      <div
                        style={{
                          display: "inline-flex",
                          alignItems: "center",
                          gap: 8,
                          background: "#f8fafc",
                          borderRadius: 10,
                          padding: "8px 12px",
                          fontSize: 12,
                          fontWeight: 600,
                          color: "#475569",
                        }}
                      >
                        <Filter size={14} />
                        Dynamic Controls
                      </div>
                    </div>
                  </div>

                  <div style={{ padding: 24, display: "grid", gridTemplateColumns: "minmax(220px, 1.6fr) minmax(180px, 1fr) minmax(180px, 1fr)", gap: 16 }}>
                    <div style={{ position: "relative" }}>
                      <Search
                        size={16}
                        style={{
                          position: "absolute",
                          left: 12,
                          top: "50%",
                          transform: "translateY(-50%)",
                          color: "#94a3b8",
                        }}
                      />
                      <input
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        placeholder="Search across all fields"
                        style={inputStyle(true)}
                      />
                    </div>

                    <select value={chartX} onChange={(e) => setChartX(e.target.value)} style={inputStyle(false)}>
                      <option value="">Category field</option>
                      {dataset?.text_columns.map((col) => (
                        <option key={col} value={col}>
                          {col}
                        </option>
                      ))}
                    </select>

                    <select value={chartY} onChange={(e) => setChartY(e.target.value)} style={inputStyle(false)}>
                      <option value="">Numeric field</option>
                      {dataset?.numeric_columns.map((col) => (
                        <option key={col} value={col}>
                          {col}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>

                <div 
                style={{ 
                  display: "grid",
                  gridTemplateColumns: "minmax(0, 1.8fr) minmax(280px, 0.9fr)", 
                  gap: 24, 
                  alignItems: "start" 
                  }}
                  >
                  <Panel
                    title={`Top ${chartX || "Category"} by ${chartY || "Metric"}`}
                    subtitle="Bar visualization"
                  >
                    <div style={{ width: "100%", height: 340 }}>
                      <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={chartData} margin={{ top: 10, right: 12, left: -12, bottom: 58 }}>
                          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                          <XAxis
                            dataKey="label"
                            angle={-20}
                            textAnchor="end"
                            interval={0}
                            height={90}
                            tick={{ fill: "#475569", fontSize: 11 }}
                          />
                          <YAxis tick={{ fill: "#475569", fontSize: 12 }} />
                          <Tooltip />
                          <Bar dataKey="value" fill="#2563eb" radius={[6, 6, 0, 0]} />
                        </BarChart>
                      </ResponsiveContainer>
                    </div>
                  </Panel>

                  <Panel title="Trend Snapshot" subtitle="Simple line preview from selected metric">
                    <div style={{ width: "100%", height: 340 }}>
                      <ResponsiveContainer width="100%" height="100%">
                        <LineChart data={trendData} margin={{ top: 10, right: 10, left: -12, bottom: 10 }}>
                          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                          <XAxis dataKey="name" tick={{ fill: "#475569", fontSize: 12 }} />
                          <YAxis tick={{ fill: "#475569", fontSize: 12 }} />
                          <Tooltip />
                          <Line type="monotone" dataKey="value" stroke="#10b981" strokeWidth={3} dot={{ r: 3 }} />
                        </LineChart>
                      </ResponsiveContainer>
                    </div>
                  </Panel>
                </div>

                <div style={panelStyle()}>
                  <div
                    style={{
                      padding: "20px 24px",
                      borderBottom: "1px solid #e2e8f0",
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                      gap: 16,
                    }}
                  >
                    <div>
                      <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700, color: "#0f172a" }}> Table View</h2>
                      <p style={{ marginTop: 6, fontSize: 14, color: "#64748b" }}>
                        Columns and rows adapt automatically to the current source
                      </p>
                    </div>
                    <div
                      style={{
                        display: "inline-flex",
                        alignItems: "center",
                        gap: 8,
                        background: "#f8fafc",
                        borderRadius: 10,
                        padding: "8px 12px",
                        fontSize: 12,
                        fontWeight: 600,
                        color: "#475569",
                      }}
                    >
                      <Table2 size={14} />
                      {previewRows.length} shown
                    </div>
                  </div>

                  <div style={{ overflowX: "auto" }}>
                    <table style={{ minWidth: "100%", borderCollapse: "collapse", fontSize: 14 }}>
                      <thead style={{ background: "#f8fafc", color: "#475569" }}>
                        <tr>
                          {previewColumns.map((col) => (
                            <th
                              key={col}
                              style={{
                                borderBottom: "1px solid #e2e8f0",
                                padding: "14px 18px",
                                textAlign: "left",
                                fontWeight: 700,
                              }}
                            >
                              {col}
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {previewRows.map((row, idx) => (
                          <tr key={idx} style={{ background: idx % 2 === 0 ? "white" : "#f8fafc80" }}>
                            {previewColumns.map((col) => (
                              <td
                                key={col}
                                style={{
                                  borderBottom: "1px solid #f1f5f9",
                                  padding: "14px 18px",
                                  color: "#334155",
                                }}
                              >
                                {row[col] == null || row[col] === "" ? "—" : String(row[col])}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              </section>

              <section style={{ display: "flex", flexDirection: "column", gap: 24, width: "300px", minWidth: "300px", }}>
                <Panel title="Current Dataset" subtitle="Live metadata from backend">
                  <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                    <InfoCard label="Table Name" value={dataset?.table_name ?? "—"} />
                    <InfoCard label="Rows Loaded" value={String(filteredRows.length)} />
                    <InfoCard label="Schema Columns" value={dataset?.columns.filter((c) => c.name !== "row_hash").map((c) => c.name).join(", ")} />
                    <InfoCard label="Numeric Columns" value={dataset?.numeric_columns.join(", ") || "—"} />
                  </div>
                </Panel>

                <Panel title="Schema Overview" subtitle="Field types discovered by inference">
                  <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                    {dataset?.columns.filter((col) => col.name !== "row_hash").slice(0, 10).map((col) => (
                      <div
                        key={col.name}
                        style={{
                          display: "flex",
                          justifyContent: "space-between",
                          alignItems: "center",
                          border: "1px solid #e2e8f0",
                          background: "#f8fafc",
                          borderRadius: 12,
                          padding: "12px 14px",
                        }}
                      >
                        <div style={{ fontSize: 14, fontWeight: 600, color: "#334155" }}>{col.name}</div>
                        <div
                          style={{
                            background: "white",
                            padding: "6px 10px",
                            borderRadius: 10,
                            fontSize: 11,
                            fontWeight: 700,
                            color: "#64748b",
                            textTransform: "uppercase",
                          }}
                        >
                          {col.type}
                        </div>
                      </div>
                    ))}
                  </div>
                </Panel>

                <Panel title="System Notes" subtitle="Presentation-friendly summary">
                  <ul style={{ margin: 0, paddingLeft: 18, color: "#475569", fontSize: 14, lineHeight: 1.8 }}>
                    <li>UI adapts automatically when the backend dataset or schema changes.</li>
                    <li>Visualizations are generated from detected text and numeric columns.</li>
                    <li>Table headers are rendered dynamically using live schema metadata.</li>
                    <li>This layout is designed for cleaner project demos and stakeholder reviews.</li>
                  </ul>
                </Panel>
              </section>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}