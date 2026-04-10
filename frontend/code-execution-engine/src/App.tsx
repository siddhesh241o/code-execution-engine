import { useState } from 'react'
import { CodeEditor } from './components/CodeEditor'
import { InputPanel } from './components/InputPanel'
export default function App() {
  const [code, setCode] = useState(`print("Hello world")`)
  const [language, setLanguage] = useState<'python' | 'javascript' | 'cpp'>('python')
  const [programInput, setProgramInput] = useState("")
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

        <aside className="panel p-4">
          <h2 className="text-sm font-semibold tracking-wide">Actions</h2>
          <p className="mt-2 text-sm text-slate-300">This is your future run and output area. Keep it as a placeholder while you build editor behavior first.</p>
          <button type="button" className="btn-primary mt-4 w-full">Run Code</button>
        </aside>
      </section>
    </main>
  )
}