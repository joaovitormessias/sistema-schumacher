# API Request Isolado

Esta pasta foi criada separada do restante do projeto para executar requisicoes de API sem impactar o codigo principal.

## Arquivo
- `request.ps1`: script PowerShell para chamar qualquer endpoint HTTP.

## Exemplos de uso

### GET simples
```powershell
./request.ps1 -Url "https://jsonplaceholder.typicode.com/posts/1"
```

### POST com headers + JSON
```powershell
./request.ps1 `
  -Url "https://httpbin.org/post" `
  -Method POST `
  -HeadersJson '{"Authorization":"Bearer SEU_TOKEN"}' `
  -BodyJson '{"nome":"Geinfo"}'
```

## Observacoes
- O script imprime a resposta em JSON formatado.
- Se quiser, eu ja executo uma chamada real para voce; so me passe URL, metodo, headers e body.
