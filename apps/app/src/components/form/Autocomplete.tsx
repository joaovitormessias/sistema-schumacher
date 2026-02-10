import { useMemo, useState, type KeyboardEvent } from "react";

export type AutocompleteOption = {
  id: string;
  label: string;
  meta?: string;
};

type AutocompleteProps = {
  value: string;
  options: AutocompleteOption[];
  placeholder?: string;
  ariaLabel: string;
  onInputChange: (value: string) => void;
  onOptionSelect?: (option: AutocompleteOption) => void;
};

function highlightLabel(label: string, query: string) {
  const term = query.trim();
  if (!term) return label;
  const index = label.toLowerCase().indexOf(term.toLowerCase());
  if (index < 0) return label;
  const before = label.slice(0, index);
  const match = label.slice(index, index + term.length);
  const after = label.slice(index + term.length);
  return (
    <>
      {before}
      <mark>{match}</mark>
      {after}
    </>
  );
}

export default function Autocomplete({
  value,
  options,
  placeholder,
  ariaLabel,
  onInputChange,
  onOptionSelect,
}: AutocompleteProps) {
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);

  const filtered = useMemo(() => {
    const term = value.trim().toLowerCase();
    if (!term) return options.slice(0, 8);
    return options
      .filter((option) =>
        `${option.label} ${option.meta ?? ""}`.toLowerCase().includes(term)
      )
      .slice(0, 8);
  }, [options, value]);

  const commitSelection = (option: AutocompleteOption) => {
    onInputChange(option.label);
    onOptionSelect?.(option);
    setOpen(false);
    setActiveIndex(-1);
  };

  const onKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (!open && event.key === "ArrowDown" && filtered.length > 0) {
      event.preventDefault();
      setOpen(true);
      setActiveIndex(0);
      return;
    }

    if (!open) return;

    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActiveIndex((prev) => Math.min(prev + 1, filtered.length - 1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActiveIndex((prev) => Math.max(prev - 1, 0));
    } else if (event.key === "Enter") {
      const selected = filtered[activeIndex];
      if (!selected) return;
      event.preventDefault();
      commitSelection(selected);
    } else if (event.key === "Escape") {
      event.preventDefault();
      setOpen(false);
      setActiveIndex(-1);
    }
  };

  return (
    <div className="autocomplete">
      <input
        className="input input-delightful"
        value={value}
        placeholder={placeholder}
        aria-label={ariaLabel}
        role="combobox"
        aria-expanded={open && filtered.length > 0}
        aria-autocomplete="list"
        aria-controls="autocomplete-options"
        onFocus={() => setOpen(true)}
        onBlur={() => {
          // Delay allows click selection on options.
          window.setTimeout(() => setOpen(false), 120);
        }}
        onChange={(event) => {
          onInputChange(event.target.value);
          setOpen(true);
          setActiveIndex(0);
        }}
        onKeyDown={onKeyDown}
      />
      {open && filtered.length > 0 ? (
        <ul className="autocomplete-list" id="autocomplete-options" role="listbox">
          {filtered.map((option, index) => (
            <li key={option.id} role="option" aria-selected={activeIndex === index}>
              <button
                type="button"
                className={`autocomplete-option${activeIndex === index ? " active" : ""}`}
                onMouseEnter={() => setActiveIndex(index)}
                onClick={() => commitSelection(option)}
              >
                <span>{highlightLabel(option.label, value)}</span>
                {option.meta ? <small>{option.meta}</small> : null}
              </button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}
