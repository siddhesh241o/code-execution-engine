import Editor from '@monaco-editor/react'
import type { editor } from 'monaco-editor'

type Language = 'python' | 'cpp' | 'javascript'

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
    javascript: 'javascript',
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
                <h2 className="text-sm font-semibold tracking-wide">{title}</h2>
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
                theme="vs-dark"
                language={monacoLanguage}
                value={value}
                options={{ ...editorOptions, readOnly }}
                onChange={(updatedValue) => onChange(updatedValue ?? '')}
            />
        </div>
    )
}

