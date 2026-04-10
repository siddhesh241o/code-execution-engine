type Language = 'python' | 'cpp' | 'javascript'

interface CodeEditorProps {
    value: string 
    onChange: (value: string) => void 
    language: Language
    onLanguageChange: (language: Language) => void 
    readOnly?: boolean 
    title?: string
}

export function CodeEditor({
    value,
    onChange, 
    language, 
    onLanguageChange,
    readOnly = false, 
    title = 'Code Editor'
}: CodeEditorProps) {
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
            <textarea
            className="editor-area"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            readOnly={readOnly}
            spellCheck={false}
            placeholder='Write your code here...'
            />
        </div>
    )
}

