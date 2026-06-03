import Editor from '@monaco-editor/react'
import type { editor } from 'monaco-editor'

type Language = 'python' | 'cpp' | 'go'

interface CodeEditorProps {
    value: string 
    onChange: (value: string) => void 
    language: Language
    onLanguageChange: (language: Language) => void 
    readOnly?: boolean 
    title?: string
}

const monacoLanguageMap: Record<Language, string> = {
    python: 'python',
    cpp: 'cpp',
    go: 'go',
}

const editorOptions: editor.IStandaloneEditorConstructionOptions = {
    lineNumbers: 'on',
    autoIndent: 'full',
    autoClosingBrackets: 'always',
    bracketPairColorization: { enabled: true },
    guides: { bracketPairs: true },
}

export function CodeEditor({
    value,
    onChange, 
    language, 
    onLanguageChange,
    readOnly = false, 
    title = 'Code Editor'
}: CodeEditorProps) {
    const monacoLanguage = monacoLanguageMap[language]

    return (
        <div className="panel overflow-hidden">
            <div className="panel-head">
                <div>
                    <h2 className="section-title">{title}</h2>
                    <p className="section-caption">Select a language and write code.</p>
                </div>
                <select 
                className="field"
                value={language}
                onChange={(e) => onLanguageChange(e.target.value as Language)}
                >
                    <option value="python">Python</option>
                    <option value="cpp">C++</option>
                    <option value="javascript">Javascript</option>
                </select>
            </div>
            <Editor
                height="420px"
                theme="vs-light"
                language={monacoLanguage}
                value={value}
                options={{ ...editorOptions, readOnly }}
                onChange={(updatedValue) => onChange(updatedValue ?? '')}
            />
        </div>
    )
}

