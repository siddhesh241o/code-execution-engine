interface InputPanelProps {
    value: string 
    onChange: (value: string) => void
    readOnly?: boolean 
    title?: string 
    rows?: number
}

export function InputPanel ({
    value,
    onChange,
    readOnly = false,
    title = "Program Input (stdin)",
    rows = 6,
}: InputPanelProps) {
    return (
        <div className="panel overflow-hidden">
            <div className="panel-head">
                <h3 className="text-sm font-semibold tracking-wide">{title}</h3>
            </div>
            <div className="p-4">
                <textarea 
                    className="input-area w-full resize-y font-mono leading-6"
                    value={value}
                    onChange={(e) => onChange(e.target.value)}
                    readOnly={readOnly}
                    spellCheck={false}
                    rows={rows}
                    placeholder="Example: 5 10"
                />
            </div>
        </div>
    )
}