import { useState } from 'react'
import { CodeEditor } from './components/CodeEditor'
import { InputPanel } from './components/InputPanel'
import { ResultPanel } from './components/ResultPanel'
import { executeCode, pollResult } from './api/execution'

type ExecutionResult = {
  status: string
  stdout?: string
  stderr?: string
  time_ms?: number
  memory_kb?: number
}

export default function App() {
  const [code, setCode] = useState(`print("Hello world")`)
  const [language, setLanguage] = useState<'python' | 'go' | 'cpp' >('python')
  const [programInput, setProgramInput] = useState("")
  const [result, setResult] = useState<ExecutionResult | null>(null)
  const [isRunning, setIsRunning] = useState(false)
  const [errorMessage, setErrorMessage] = useState("")

  const handleRunCode = async () => {
    setIsRunning(true)
    setErrorMessage("")
    setResult({ status: "Queued" })

    const backendLanguage = language;

    try {
      const { job_id } = await executeCode({
        code,
        language: backendLanguage,
        input: programInput,
      })

      const finalResult = await pollResult(job_id)
      setResult(finalResult)
    } catch (error: any) {
      setErrorMessage(error.message || "Something went wrong while running code.")
      setResult(null)
    } finally {
      setIsRunning(false)
    }
  }
  return (
    <main className="app-shell">
      <header className="mb-4 md:mb-6">
        <h1 className="page-title">Code Playground</h1>
        <p className="page-subtitle">Write, run, and iterate quickly.</p>
      </header>

      <section className="grid gap-4 lg:grid-cols-[1fr_320px]">
        <div className="grid gap-4">
          <CodeEditor
            value={code}
            onChange={setCode}
            language={language}
            onLanguageChange={setLanguage}
            title="Playground"
          />
          <InputPanel 
            value={programInput}
            onChange={setProgramInput}
          />
        </div>
        <div className="grid gap-4">
          <aside className="panel p-4">
            <h2 className="text-sm font-semibold tracking-wide">Actions</h2>
            <p className="mt-2 text-sm text-muted">Run Code</p>
            <button
              type="button"
              className="btn-primary mt-4 w-full disabled:opacity-60"
              onClick={handleRunCode}
              disabled={isRunning}
            >
              {isRunning ? "Running..." : "Run Code"}
            </button>
          </aside>
          <ResultPanel
            result={result}
            isRunning={isRunning}
            errorMessage={errorMessage}
          />
        </div>
      </section>
    </main>
  )
}