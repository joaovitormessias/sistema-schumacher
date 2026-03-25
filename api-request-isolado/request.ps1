param(
  [Parameter(Mandatory=$true)]
  [string]$Url,

  [ValidateSet('GET','POST','PUT','PATCH','DELETE')]
  [string]$Method = 'GET',

  [string]$BodyJson,

  [string]$HeadersJson,

  [int]$TimeoutSec = 30
)

$headers = @{}
if ($HeadersJson) {
  $parsedHeaders = $HeadersJson | ConvertFrom-Json -AsHashtable
  foreach ($k in $parsedHeaders.Keys) {
    $headers[$k] = [string]$parsedHeaders[$k]
  }
}

$params = @{
  Uri         = $Url
  Method      = $Method
  TimeoutSec  = $TimeoutSec
  ErrorAction = 'Stop'
}

if ($headers.Count -gt 0) {
  $params.Headers = $headers
}

if ($BodyJson) {
  $params.Body = $BodyJson
  if (-not $headers.ContainsKey('Content-Type')) {
    $params.ContentType = 'application/json'
  }
}

try {
  $response = Invoke-RestMethod @params
  $response | ConvertTo-Json -Depth 20
}
catch {
  Write-Error "Falha na requisicao: $($_.Exception.Message)"
  exit 1
}
