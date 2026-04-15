import { useState } from 'react'
import { CodeEditor } from './components/CodeEditor'
import { InputPanel } from './components/InputPanel'
import { ResultPanel } from './components/ResultPanel'
type ExecutionResult = {
  status: string
  stdout?: string
  stderr?: string
  time_ms?: number
  memory_kb?: number
}

export default function App() {
  const [code, setCode] = useState(`print("Hello world")`)
  const [language, setLanguage] = useState<'python' | 'javascript' | 'cpp'>('python')
  const [programInput, setProgramInput] = useState("")
  const [result, setResult] = useState<ExecutionResult | null>(null)
  const [isRunning, setIsRunning] = useState(false)
  const [errorMessage, setErrorMessage] = useState("")
  const handleRunCode = async () => {
    setIsRunning(true)
    setErrorMessage("")
    setResult({ status: "Processing" })

    try {
      await new Promise((resolve) => setTimeout(resolve, 900))

      setResult({
        status: "Successfully Executed",
        stdout: `Language: ${language}\nInput: ${programInput || "(empty)"}\n\nCode received:\n${code}`,
        stderr: "",
        time_ms: 32,
        memory_kb: 4096,
      })
    } catch {
      setErrorMessage("Something went wrong while running code.")
      setResult(null)
    } finally {
      setIsRunning(false)
    }
  }
  return (
    <main className="app-shell">
      <header className="mb-4 md:mb-6">
        <h1 className="page-title">Code Playground</h1>
        <p className="page-subtitle">Write, run, and iterate quickly. Keep this layout and reuse the editor in future challenge screens.</p>
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
            <p className="mt-2 text-sm text-muted">This is your future run and output area. Keep it as a placeholder while you build editor behavior first.</p>
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