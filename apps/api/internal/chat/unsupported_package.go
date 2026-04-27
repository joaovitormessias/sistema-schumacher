package chat

import (
	"strings"
)

const unsupportedPackageSupportPhone = "+55 49 9886-2222"

type unsupportedPackageQuery struct {
	StateCode string
	StateName string
	Location  string
}

type brazilStateMention struct {
	Code       string
	Name       string
	FoldedName string
}

var unsupportedPackageStates = []brazilStateMention{
	{Code: "AC", Name: "Acre", FoldedName: "acre"},
	{Code: "AL", Name: "Alagoas", FoldedName: "alagoas"},
	{Code: "AP", Name: "Amapa", FoldedName: "amapa"},
	{Code: "AM", Name: "Amazonas", FoldedName: "amazonas"},
	{Code: "BA", Name: "Bahia", FoldedName: "bahia"},
	{Code: "CE", Name: "Ceara", FoldedName: "ceara"},
	{Code: "DF", Name: "Distrito Federal", FoldedName: "distrito federal"},
	{Code: "ES", Name: "Espirito Santo", FoldedName: "espirito santo"},
	{Code: "GO", Name: "Goias", FoldedName: "goias"},
	{Code: "MT", Name: "Mato Grosso", FoldedName: "mato grosso"},
	{Code: "MS", Name: "Mato Grosso do Sul", FoldedName: "mato grosso do sul"},
	{Code: "MG", Name: "Minas Gerais", FoldedName: "minas gerais"},
	{Code: "PA", Name: "Para", FoldedName: ""},
	{Code: "PB", Name: "Paraiba", FoldedName: "paraiba"},
	{Code: "PR", Name: "Parana", FoldedName: "parana"},
	{Code: "PE", Name: "Pernambuco", FoldedName: "pernambuco"},
	{Code: "PI", Name: "Piaui", FoldedName: "piaui"},
	{Code: "RJ", Name: "Rio de Janeiro", FoldedName: "rio de janeiro"},
	{Code: "RN", Name: "Rio Grande do Norte", FoldedName: "rio grande do norte"},
	{Code: "RS", Name: "Rio Grande do Sul", FoldedName: "rio grande do sul"},
	{Code: "RO", Name: "Rondonia", FoldedName: "rondonia"},
	{Code: "RR", Name: "Roraima", FoldedName: "roraima"},
	{Code: "SP", Name: "Sao Paulo", FoldedName: "sao paulo"},
	{Code: "SE", Name: "Sergipe", FoldedName: "sergipe"},
	{Code: "TO", Name: "Tocantins", FoldedName: "tocantins"},
}

func inferUnsupportedPackageQuery(text string) (unsupportedPackageQuery, bool) {
	body := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if body == "" {
		return unsupportedPackageQuery{}, false
	}

	locations := extractCanonicalLocations(body)
	if !looksLikeUnsupportedPackageTravelQuery(body, len(locations)) {
		return unsupportedPackageQuery{}, false
	}

	for _, location := range locations {
		if query, ok := unsupportedQueryFromLocation(location); ok {
			return query, true
		}
	}

	folded := foldChatText(body)
	for _, state := range unsupportedPackageStates {
		if state.FoldedName != "" && strings.Contains(folded, " "+state.FoldedName+" ") {
			return unsupportedPackageQuery{StateCode: state.Code, StateName: state.Name}, true
		}
		if strings.Contains(folded, " "+strings.ToLower(state.Code)+" ") {
			return unsupportedPackageQuery{StateCode: state.Code, StateName: state.Name}, true
		}
	}
	return unsupportedPackageQuery{}, false
}

func looksLikeUnsupportedPackageTravelQuery(text string, locationCount int) bool {
	if looksLikeAvailabilityIntent(text, locationCount) {
		return true
	}
	folded := foldChatText(text)
	patterns := []string{
		" quero ir ",
		" ir para ",
		" viajar para ",
		" viagem para ",
		" onibus para ",
		" passagem para ",
	}
	for _, pattern := range patterns {
		if strings.Contains(folded, pattern) {
			return true
		}
	}
	return false
}

func unsupportedQueryFromLocation(location string) (unsupportedPackageQuery, bool) {
	parts := strings.Split(strings.TrimSpace(location), "/")
	if len(parts) != 2 {
		return unsupportedPackageQuery{}, false
	}
	code := strings.ToUpper(strings.TrimSpace(parts[1]))
	if code == "" || code == "SC" || code == "MA" {
		return unsupportedPackageQuery{}, false
	}
	for _, state := range unsupportedPackageStates {
		if state.Code == code {
			return unsupportedPackageQuery{
				StateCode: code,
				StateName: state.Name,
				Location:  strings.TrimSpace(location),
			}, true
		}
	}
	return unsupportedPackageQuery{}, false
}

func buildUnsupportedPackageDraftRun(query unsupportedPackageQuery) RunAgentResult {
	reply := buildUnsupportedPackageReply(query)
	return RunAgentResult{
		ReplyText: reply,
		Model:     "deterministic_unsupported_package",
		RequestPayload: map[string]interface{}{
			"mode":       "UNSUPPORTED_PACKAGE_ROUTE",
			"state_code": query.StateCode,
			"state_name": query.StateName,
			"location":   query.Location,
		},
		ResponsePayload: map[string]interface{}{
			"reply_text": reply,
		},
	}
}

func buildUnsupportedPackageReply(query unsupportedPackageQuery) string {
	target := strings.TrimSpace(query.Location)
	if target == "" {
		target = strings.TrimSpace(query.StateName)
	}
	if target == "" {
		target = strings.TrimSpace(query.StateCode)
	}
	if target != "" {
		return "No momento nao ha passagens disponiveis para " + target + " nos pacotes atendidos pela Schumacher. Para verificar outra possibilidade, entre em contato com o suporte pelo " + unsupportedPackageSupportPhone + "."
	}
	return "No momento nao ha passagens disponiveis para essa rota nos pacotes atendidos pela Schumacher. Para verificar outra possibilidade, entre em contato com o suporte pelo " + unsupportedPackageSupportPhone + "."
}
