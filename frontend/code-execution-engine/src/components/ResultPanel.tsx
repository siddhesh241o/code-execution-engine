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
        <h3 className="text-sm font-semibold tracking-wide">{title}</h3>
        <span className={`text-xs font-medium ${getStatusClass(statusText)}`}>
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
          <div className="error-box">
            {errorMessage}
          </div>
        )}

        {(result?.stdout ?? "").length > 0 && (
          <div>
            <p className="output-label">stdout</p>
            <pre className="output-block">
              {result?.stdout}
            </pre>
          </div>
        )}

        {(result?.stderr ?? "").length > 0 && (
          <div>
            <p className="output-label">stderr</p>
            <pre className="output-block output-block-error">
              {result?.stderr}
            </pre>
          </div>
        )}
        
        {result && (
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div className="metric-card">
              <p className="text-xs text-muted">Time</p>
              <p>{result.time_ms ?? "-"} ms</p>
            </div>
            <div className="metric-card">
              <p className="text-xs text-muted">Memory</p>
              <p>{result.memory_kb ?? "-"} KB</p>
            </div>
          </div>
        )}
      </div>
    </section>
  )
}