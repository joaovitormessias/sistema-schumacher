**Design System**

Este documento resume os tokens e o uso padrao do tema em `apps/app/src/styles/theme.css`.

**Colors**
```
--bg
--panel
--panel-muted
--surface-elevated
--text
--text-strong
--text-weak
--muted
--accent
--accent-strong
--border-light
--border-medium
--border-strong
--success
--warning
--danger
--info
```

**Typography**
```
--display-lg / --display-md
--heading-lg / --heading-md / --heading-sm
--body-lg / --body-md / --body-sm
```

**Spacing**
```
--space-0  --space-1  --space-2  --space-3
--space-4  --space-5  --space-6  --space-7
```

**Radius**
```
--radius-sm  --radius-md  --radius-lg  --radius-xl  --radius-full
```

**Shadows**
```
--shadow-sm  --shadow-md  --shadow-lg  --shadow-xl  --shadow-focus
```

**Component Tokens**
```
--input-height
--input-padding-x
--button-height
--table-row-padding
--card-padding
```

**Component Classes**
Use estas classes como base para manter consistencia visual.
```
.page
.section
.page-header
.form-grid
.form-actions
.full-span
.space-top-2
.toolbar
.table
.status-badge
.segmented
```

**Exemplo de uso**
```html
<section class="page">
  <div class="section">
    <div class="section-header">
      <div class="section-title">Titulo</div>
    </div>
    <div class="form-grid">
      <label class="form-field">
        <span class="form-label">Nome</span>
        <input class="input" />
      </label>
      <div class="form-actions">
        <button class="button">Salvar</button>
      </div>
    </div>
  </div>
</section>
```
