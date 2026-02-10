import { useRef, useState, type DragEvent } from "react";

type FileUploadProps = {
  value?: File | null;
  onChange: (file: File | null) => void;
  accept?: string;
  capture?: "user" | "environment";
  disabled?: boolean;
  label?: string;
  hint?: string;
};

export default function FileUpload({
  value,
  onChange,
  accept,
  capture,
  disabled,
  label = "Selecionar arquivo",
  hint,
}: FileUploadProps) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [dragging, setDragging] = useState(false);

  const onDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    if (disabled) return;
    setDragging(false);
    const file = event.dataTransfer.files?.[0] ?? null;
    onChange(file);
  };

  return (
    <div
      className={`file-upload${dragging ? " dragging" : ""}${disabled ? " is-disabled" : ""}`}
      onDragOver={(event) => {
        event.preventDefault();
        if (!disabled) setDragging(true);
      }}
      onDragLeave={() => setDragging(false)}
      onDrop={onDrop}
    >
      <input
        ref={inputRef}
        className="file-upload-input"
        type="file"
        accept={accept}
        capture={capture}
        disabled={disabled}
        onChange={(event) => onChange(event.target.files?.[0] ?? null)}
      />
      <div className="file-upload-content">
        <div className="file-upload-title">{label}</div>
        <div className="file-upload-subtitle">
          {value ? value.name : hint ?? "Arraste aqui ou toque para selecionar"}
        </div>
      </div>
      <div className="file-upload-actions">
        <button
          type="button"
          className="button secondary sm"
          disabled={disabled}
          onClick={() => inputRef.current?.click()}
        >
          {value ? "Trocar" : "Escolher"}
        </button>
        {value ? (
          <button type="button" className="button ghost sm" onClick={() => onChange(null)} disabled={disabled}>
            Remover
          </button>
        ) : null}
      </div>
    </div>
  );
}
