type ExecutionResult = {
  status: string
  stdout?: string
  stderr?: string
  time_ms?: number
  memory_kb?: number
}

type ResultPanelProps = {
  result: ExecutionResult | null
  isRunning: boolean
  errorMessage?: string
  title?: string
}

function getStatusClass(status: string) {
  const s = status.toLowerCase()
  if (s.includes("success")) return "status-success"
  if (s.includes("processing") || s.includes("queued")) return "status-progress"
  if (s.includes("error") || s.includes("failed")) return "status-error"
  return "text-muted"
}

export function ResultPanel({
  result,
  isRunning,
  errorMessage,
  title = "Execution Result",
}: ResultPanelProps) {
  const showIdle = !result && !isRunning && !errorMessage
  const statusText =
    errorMessage ? "Request Error" : result?.status ?? (isRunning ? "Processing" : "Idle")

  return (
    <section className="panel overflow-hidden">
      <div className="panel-head">
        <div>
          <h3 className="section-title">{title}</h3>
          <p className="section-caption">Program output and execution summary.</p>
        </div>
        <span className={`status-pill ${getStatusClass(statusText)}`}>
          {statusText}
        </span>
      </div>

      <div className="space-y-4 p-4">
        {showIdle && (
          <p className="text-sm text-muted">
            Run your code to see output, errors, and execution metrics.
          </p>
        )}

        {errorMessage && (
          <div className="notice notice-error">{errorMessage}</div>
        )}

        <div className="output-stack">
          <div className="output-card">
            <p className="output-label">stdout</p>
            <pre className="output-block">
              {result?.stdout?.trim() ? result.stdout : "No standard output."}
            </pre>
          </div>

          <div className="output-card">
            <p className="output-label">stderr</p>
            <pre className="output-block output-block-error">
              {result?.stderr?.trim() ? result.stderr : "No error output."}
            </pre>
          </div>
        </div>

        {result && (
          <div className="metrics-grid text-sm">
            <div className="metric-card">
              <p className="metric-label">Time</p>
              <p className="metric-value">{result.time_ms ?? "-"} ms</p>
            </div>
            <div className="metric-card">
              <p className="metric-label">Memory</p>
              <p className="metric-value">{result.memory_kb ?? "-"} KB</p>
            </div>
          </div>
        )}
      </div>
    </section>
  )
}