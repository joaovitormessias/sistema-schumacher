package chat

import (
	"fmt"
	"strings"
	"time"
)

const defaultAgentSystemPrompt = `Voce e o atendente Shabas da Schumacher Tur.
Responda sempre em Portugues do Brasil.
Use mensagens curtas, diretas, cordiais e sem markdown ou emojis.
Use a mensagem atual do cliente como fonte principal do turno.
Use o historico recente para manter estado e nao repetir perguntas.
Faca no maximo uma pergunta por resposta.
Pergunte so o que for estritamente necessario para o proximo passo.
Nunca exponha IDs internos, classificacoes internas ou raciocinio interno.
Quando houver RESULTADO DE FERRAMENTA no contexto, use-o como fonte de verdade operacional.
Nao invente datas, horarios, preco, rota, pagamento ou disponibilidade operacional.
Se faltarem varios dados criticos, faca uma unica pergunta combo curta.
Se faltar apenas um dado critico, pergunte apenas esse dado.
Se a consulta for aberta, responda primeiro e refine depois.
Se a consulta ampla for sobre Santa Catarina ou SC, responda direto com a tabela publica:
Fraiburgo R$ 950; Monte Carlo R$ 950; Videira R$ 950; Campos Novos R$ 1000; Chapeco R$ 1100; Concordia R$ 1100; Ipumirim R$ 1100; Petrolandia R$ 1100; Ituporanga R$ 1100; Seara R$ 1100.
Depois dessa tabela, nao faca pergunta adicional no mesmo turno.
Se a consulta ampla for sobre Maranhao ou MA, nao informe valor unico antes; peca a cidade de saida em Santa Catarina.
Se o cliente fizer follow-up curto como "pra quando?", "quais datas?" ou "tem vaga quando?", reutilize a rota, o destino ou o package ja inferidos e nao reabra a coleta de origem/destino.
Se o contexto ainda estiver so no nivel do estado e houver busca de datas, use o package correspondente antes de perguntar cidade.
Ao responder datas, priorize ate 5 datas futuras e nao mostre IDs internos nem contagem de assentos.
Nao inverta origem e destino so porque o cliente citou Santa Catarina ou Maranhao.
Se a viagem estiver indo para Santa Catarina e a origem ainda faltar, a pergunta correta e sobre a cidade de saida no Maranhao.
Se a viagem estiver indo para Maranhao e a origem ainda faltar, a pergunta correta e sobre a cidade de saida em Santa Catarina.
Antes de coletar nome, documento ou criar reserva, garanta que a rota ja esteja definida em nivel de cidade.
Na reserva, aceite nome completo + documento digitados ou foto legivel do documento.
Para foto: RG e CNH pedem frente e verso; certidao pede ao menos a frente.
Na extracao por foto, use a hierarquia CPF > RG > matricula da certidao > numero da CNH como fallback.
Depois da extracao, confirme nome completo + tipo + numero do documento antes de criar a reserva.
Crianca de colo com ate 5 anos entra no cadastro, mas nao entra na cobranca.
Se a rota ou cidade estiver fora do pacote atendido, oriente contato humano no numero +55 49 9886-2222.`

func buildAgentSystemPrompt() string {
	return defaultAgentSystemPrompt
}

func buildAgentUserPrompt(session Session, memory map[string]interface{}, tools agentToolContext) string {
	var builder strings.Builder
	now := time.Now().UTC().Format(time.RFC3339)

	builder.WriteString("CONTEXTO DO ATENDIMENTO\n")
	builder.WriteString(fmt.Sprintf("- Telefone: %s\n", strings.TrimSpace(session.CustomerPhone)))
	builder.WriteString(fmt.Sprintf("- Nome: %s\n", strings.TrimSpace(session.CustomerName)))
	builder.WriteString(fmt.Sprintf("- Canal: %s\n", strings.TrimSpace(session.Channel)))
	builder.WriteString(fmt.Sprintf("- Data/Hora UTC: %s\n", now))
	currentTurn := strings.TrimSpace(asString(memory["current_turn_body"]))
	builder.WriteString(fmt.Sprintf("- Mensagem atual do cliente: %q\n", currentTurn))

	recentMessages := normalizeRecentMemoryMessages(memory["recent_messages"])
	if len(recentMessages) > 0 {
		builder.WriteString("\nHISTORICO RECENTE\n")
		for _, message := range recentMessages {
			direction := strings.TrimSpace(asString(message["direction"]))
			body := strings.TrimSpace(asString(message["body"]))
			if body == "" {
				continue
			}
			builder.WriteString(fmt.Sprintf("- %s: %s\n", direction, body))
		}
	}

	if last := strings.TrimSpace(asString(memory["last_customer_message"])); last != "" {
		builder.WriteString(fmt.Sprintf("\nULTIMA MENSAGEM DO CLIENTE NO HISTORICO: %q\n", last))
	}
	if last := strings.TrimSpace(asString(memory["last_assistant_message"])); last != "" {
		builder.WriteString(fmt.Sprintf("ULTIMA RESPOSTA DO ATENDIMENTO: %q\n", last))
	}

	context := derivePromptConversationContext(currentTurn, recentMessages)
	if context.PackageName != "" || context.Origin != "" || context.Destination != "" || context.RouteDirection != "" || context.ShouldRespondWithSCTable || context.ShortDateFollowUp {
		builder.WriteString("\nCONTEXTO DERIVADO\n")
		if context.Origin != "" {
			builder.WriteString(fmt.Sprintf("- Origem inferida: %s\n", context.Origin))
		}
		if context.Destination != "" {
			builder.WriteString(fmt.Sprintf("- Destino inferido: %s\n", context.Destination))
		}
		if context.PackageName != "" {
			builder.WriteString(fmt.Sprintf("- Package inferido: %s\n", context.PackageName))
		}
		if context.RouteDirection != "" {
			builder.WriteString(fmt.Sprintf("- Direcao inferida: %s\n", context.RouteDirection))
		}
		if context.ShortDateFollowUp {
			builder.WriteString("- Caso atual: follow-up curto sobre datas/disponibilidade. Reutilize o contexto acima e nao volte a perguntar origem ou destino se isso ja estiver implicito.\n")
		}
		if context.ShouldRespondWithSCTable {
			builder.WriteString("- Caso atual: consulta ampla sobre Santa Catarina. Resposta esperada neste turno: devolver a tabela publica de cidades e valores e encerrar sem pergunta adicional.\n")
		}
		if context.ShouldAskSCOriginForMA {
			builder.WriteString("- Caso atual: consulta ampla sobre Maranhao. Pergunta permitida neste turno: pedir apenas a cidade de saida em Santa Catarina.\n")
		}
		if context.RouteDirection == "TO_SC" {
			builder.WriteString("- Guardrail de direcao: se ainda faltar a origem para essa viagem, a pergunta correta e sobre a cidade de saida no Maranhao.\n")
		}
		if context.RouteDirection == "TO_MA" {
			builder.WriteString("- Guardrail de direcao: se ainda faltar a origem para essa viagem, a pergunta correta e sobre a cidade de saida em Santa Catarina.\n")
		}
	}

	if tools.Availability != nil {
		builder.WriteString("\nRESULTADO DE FERRAMENTA\n")
		filter := tools.Availability.Filter
		builder.WriteString(fmt.Sprintf("- Tool: %s\n", toolNameAvailabilitySearch))
		builder.WriteString(fmt.Sprintf("- Origem confirmada: %s\n", strings.TrimSpace(filter.Origin)))
		builder.WriteString(fmt.Sprintf("- Destino confirmado: %s\n", strings.TrimSpace(filter.Destination)))
		if filter.PackageName != "" {
			builder.WriteString(fmt.Sprintf("- Package consultado: %s\n", strings.TrimSpace(filter.PackageName)))
		}
		if filter.TripDate != nil {
			builder.WriteString(fmt.Sprintf("- Data consultada: %s\n", filter.TripDate.UTC().Format("2006-01-02")))
		}
		builder.WriteString(fmt.Sprintf("- Quantidade consultada: %d\n", filter.Qty))
		if len(tools.Availability.Results) == 0 {
			builder.WriteString("- Resultado: nenhuma viagem encontrada com esse filtro.\n")
			builder.WriteString("- Se responder ao cliente, diga que nao encontrou disponibilidade com esse filtro e peca outra data ou cidade.\n")
		} else {
			for index, item := range tools.Availability.Results {
				builder.WriteString(fmt.Sprintf(
					"- Opcao %d: %s -> %s | data %s | saida %s | R$ %.2f %s | pacote %s\n",
					index+1,
					strings.TrimSpace(item.OriginDisplayName),
					strings.TrimSpace(item.DestinationDisplayName),
					strings.TrimSpace(item.TripDate),
					strings.TrimSpace(item.OriginDepartTime),
					item.Price,
					strings.TrimSpace(item.Currency),
					strings.TrimSpace(item.PackageName),
				))
			}
			builder.WriteString("- Use apenas os itens acima para falar de data, horario, preco e disponibilidade.\n")
			builder.WriteString("- Se a busca foi por package ou follow-up curto sobre datas, priorize listar ate 5 datas futuras e nao reabra a coleta de origem/destino nesse turno.\n")
		}
	}

	if tools.Pricing != nil {
		builder.WriteString("\nRESULTADO DE FERRAMENTA\n")
		builder.WriteString(fmt.Sprintf("- Tool: %s\n", toolNamePricingQuote))
		builder.WriteString(fmt.Sprintf("- Fare mode: %s\n", strings.TrimSpace(tools.Pricing.Filter.FareMode)))
		if len(tools.Pricing.Results) == 0 {
			builder.WriteString("- Resultado: nenhuma cotacao retornada para os trechos consultados.\n")
			builder.WriteString("- Se responder ao cliente, diga que nao conseguiu confirmar o valor e peca ajuda humana.\n")
		} else {
			for index, item := range tools.Pricing.Results {
				builder.WriteString(fmt.Sprintf(
					"- Cotacao %d: %s -> %s | data %s | saida %s | base R$ %.2f | calculado R$ %.2f | final R$ %.2f %s\n",
					index+1,
					strings.TrimSpace(item.OriginDisplayName),
					strings.TrimSpace(item.DestinationDisplayName),
					strings.TrimSpace(item.TripDate),
					strings.TrimSpace(item.OriginDepartTime),
					item.BaseAmount,
					item.CalcAmount,
					item.FinalAmount,
					strings.TrimSpace(item.Currency),
				))
			}
			builder.WriteString("- Use os valores da cotacao como fonte de verdade para falar de preco.\n")
		}
	}

	if tools.Booking != nil {
		builder.WriteString("\nRESULTADO DE FERRAMENTA\n")
		filter := tools.Booking.Filter
		builder.WriteString(fmt.Sprintf("- Tool: %s\n", toolNameBookingLookup))
		if filter.BookingID != "" {
			builder.WriteString(fmt.Sprintf("- Booking ID consultado: %s\n", strings.TrimSpace(filter.BookingID)))
		}
		if filter.ReservationCode != "" {
			builder.WriteString(fmt.Sprintf("- Codigo de reserva consultado: %s\n", strings.TrimSpace(filter.ReservationCode)))
		}
		if len(tools.Booking.Results) == 0 {
			builder.WriteString("- Resultado: nenhuma reserva encontrada com esse identificador.\n")
			builder.WriteString("- Se responder ao cliente, diga que nao encontrou a reserva e peca confirmacao do codigo.\n")
		} else {
			for index, item := range tools.Booking.Results {
				builder.WriteString(fmt.Sprintf(
					"- Reserva %d: booking %s | codigo %s | status %s | passageiro %s | total R$ %.2f | sinal R$ %.2f | restante R$ %.2f | assento %d\n",
					index+1,
					strings.TrimSpace(item.ID),
					strings.TrimSpace(item.ReservationCode),
					strings.TrimSpace(item.Status),
					strings.TrimSpace(item.PassengerName),
					item.TotalAmount,
					item.DepositAmount,
					item.RemainderAmount,
					item.SeatNumber,
				))
				if item.ExpiresAt != nil {
					builder.WriteString(fmt.Sprintf("- Reserva %d expira em: %s\n", index+1, item.ExpiresAt.UTC().Format(time.RFC3339)))
				}
			}
			builder.WriteString("- Use apenas os dados acima para falar de status, codigo, expiracao e valores da reserva.\n")
		}
	}

	if tools.BookingCreate != nil {
		builder.WriteString("\nRESULTADO DE FERRAMENTA\n")
		builder.WriteString(fmt.Sprintf("- Tool: %s\n", toolNameBookingCreate))
		builder.WriteString(fmt.Sprintf(
			"- Tentativa de reserva: %s -> %s | data %s | saida %s | qtd %d\n",
			strings.TrimSpace(tools.BookingCreate.Filter.OriginDisplayName),
			strings.TrimSpace(tools.BookingCreate.Filter.DestinationDisplayName),
			strings.TrimSpace(tools.BookingCreate.Filter.TripDate),
			strings.TrimSpace(tools.BookingCreate.Filter.DepartureTime),
			tools.BookingCreate.Filter.Qty,
		))
		builder.WriteString(fmt.Sprintf("- Resultado operacional: %s\n", strings.TrimSpace(tools.BookingCreate.Mode)))
		if tools.BookingCreate.BookingID != "" || tools.BookingCreate.ReservationCode != "" {
			builder.WriteString(fmt.Sprintf(
				"- Reserva criada: booking %s | codigo %s | status %s | total R$ %.2f | sinal R$ %.2f | restante R$ %.2f\n",
				strings.TrimSpace(tools.BookingCreate.BookingID),
				strings.TrimSpace(tools.BookingCreate.ReservationCode),
				strings.TrimSpace(tools.BookingCreate.Status),
				tools.BookingCreate.TotalAmount,
				tools.BookingCreate.DepositAmount,
				tools.BookingCreate.RemainderAmount,
			))
			if tools.BookingCreate.ReservedUntil != nil {
				builder.WriteString(fmt.Sprintf("- Reserva criada expira em: %s\n", tools.BookingCreate.ReservedUntil.UTC().Format(time.RFC3339)))
			}
		}
		for index, item := range tools.BookingCreate.Passengers {
			builder.WriteString(fmt.Sprintf(
				"- Passageiro %d: %s | documento %s %s | assento %s\n",
				index+1,
				strings.TrimSpace(item.Name),
				strings.TrimSpace(item.DocumentType),
				strings.TrimSpace(item.Document),
				strings.TrimSpace(item.SeatID),
			))
		}
		for _, item := range tools.BookingCreate.Errors {
			builder.WriteString(fmt.Sprintf("- Erro operacional: %s\n", strings.TrimSpace(item)))
		}
		if message := strings.TrimSpace(tools.BookingCreate.MessageForAgent); message != "" {
			builder.WriteString(fmt.Sprintf("- Instrucao operacional: %s\n", message))
		}
		builder.WriteString("- Se a reserva ja foi criada, nao peca os mesmos dados de novo; informe o codigo de reserva e siga para pagamento.\n")
	}

	if tools.Reschedule != nil {
		builder.WriteString("\nRESULTADO DE FERRAMENTA\n")
		builder.WriteString(fmt.Sprintf("- Tool: %s\n", toolNameRescheduleLookup))
		if tools.Reschedule.Booking != nil {
			builder.WriteString(fmt.Sprintf("- Booking consultado: %s | codigo %s | status %s\n",
				strings.TrimSpace(tools.Reschedule.Booking.ID),
				strings.TrimSpace(tools.Reschedule.Booking.ReservationCode),
				strings.TrimSpace(tools.Reschedule.Booking.Status),
			))
		}
		if tools.Reschedule.Current.Origin != "" || tools.Reschedule.Current.Destination != "" || tools.Reschedule.Current.TripDate != "" {
			builder.WriteString(fmt.Sprintf(
				"- Viagem atual: %s -> %s | data %s | passageiros %d\n",
				strings.TrimSpace(tools.Reschedule.Current.Origin),
				strings.TrimSpace(tools.Reschedule.Current.Destination),
				strings.TrimSpace(tools.Reschedule.Current.TripDate),
				tools.Reschedule.Current.PassengerCount,
			))
		}
		builder.WriteString(fmt.Sprintf(
			"- Pedido de reagendamento: %s -> %s | nova data %s | qtd %d\n",
			strings.TrimSpace(tools.Reschedule.Requested.Origin),
			strings.TrimSpace(tools.Reschedule.Requested.Destination),
			strings.TrimSpace(tools.Reschedule.Requested.TripDate),
			tools.Reschedule.Requested.Qty,
		))
		builder.WriteString(fmt.Sprintf("- Resultado operacional: %s\n", strings.TrimSpace(tools.Reschedule.Mode)))
		if len(tools.Reschedule.Errors) > 0 {
			for _, item := range tools.Reschedule.Errors {
				builder.WriteString(fmt.Sprintf("- Erro operacional: %s\n", strings.TrimSpace(item)))
			}
		}
		if len(tools.Reschedule.Options) == 0 {
			builder.WriteString("- Resultado: nenhuma opcao valida retornada para concluir o reagendamento nesta etapa.\n")
		} else {
			for index, item := range tools.Reschedule.Options {
				builder.WriteString(fmt.Sprintf(
					"- Opcao %d: %s -> %s | data %s | saida %s | %d assentos | R$ %.2f %s | pacote %s\n",
					index+1,
					strings.TrimSpace(item.Origin),
					strings.TrimSpace(item.Destination),
					strings.TrimSpace(item.TripDate),
					strings.TrimSpace(item.DepartureTime),
					item.SeatsAvailable,
					item.Price,
					strings.TrimSpace(item.Currency),
					strings.TrimSpace(item.PackageName),
				))
			}
		}
		if len(tools.Reschedule.FieldsRequiredForManualCompletion) > 0 {
			builder.WriteString(fmt.Sprintf(
				"- Campos para conclusao manual: %s\n",
				strings.Join(tools.Reschedule.FieldsRequiredForManualCompletion, ", "),
			))
		}
		if message := strings.TrimSpace(tools.Reschedule.MessageForAgent); message != "" {
			builder.WriteString(fmt.Sprintf("- Instrucao operacional: %s\n", message))
		}
		builder.WriteString("- Nunca confirme o reagendamento como concluido; apresente contexto/opcoes e deixe claro que a troca depende de revisao humana.\n")
	}

	if tools.Payments != nil {
		builder.WriteString("\nRESULTADO DE FERRAMENTA\n")
		filter := tools.Payments.Filter
		builder.WriteString(fmt.Sprintf("- Tool: %s\n", toolNamePaymentStatus))
		if filter.BookingID != "" {
			builder.WriteString(fmt.Sprintf("- Booking ID consultado: %s\n", strings.TrimSpace(filter.BookingID)))
		}
		if filter.ReservationCode != "" {
			builder.WriteString(fmt.Sprintf("- Codigo de reserva consultado: %s\n", strings.TrimSpace(filter.ReservationCode)))
		}
		if len(tools.Payments.Results) == 0 {
			builder.WriteString("- Resultado: nenhum pagamento localizado para essa reserva.\n")
			builder.WriteString("- Se responder ao cliente, diga que ainda nao encontrou pagamento registrado e peca confirmacao do codigo se necessario.\n")
		} else {
			totalAmount := 0.0
			paidAmount := 0.0
			for index, item := range tools.Payments.Results {
				builder.WriteString(fmt.Sprintf(
					"- Pagamento %d: id %s | status %s | metodo %s | valor R$ %.2f | provedor %s\n",
					index+1,
					strings.TrimSpace(item.ID),
					strings.TrimSpace(item.Status),
					strings.TrimSpace(item.Method),
					item.Amount,
					strings.TrimSpace(item.Provider),
				))
				builder.WriteString(fmt.Sprintf("- Pagamento %d criado em: %s\n", index+1, item.CreatedAt.UTC().Format(time.RFC3339)))
				if item.ProviderRef != "" {
					builder.WriteString(fmt.Sprintf("- Pagamento %d referencia do provedor: %s\n", index+1, strings.TrimSpace(item.ProviderRef)))
				}
				if item.PaidAt != nil {
					builder.WriteString(fmt.Sprintf("- Pagamento %d pago em: %s\n", index+1, item.PaidAt.UTC().Format(time.RFC3339)))
				}
				totalAmount += item.Amount
				if strings.EqualFold(item.Status, "PAID") {
					paidAmount += item.Amount
				}
			}
			builder.WriteString(fmt.Sprintf("- Total localizado em pagamentos: R$ %.2f\n", totalAmount))
			builder.WriteString(fmt.Sprintf("- Total efetivamente pago: R$ %.2f\n", paidAmount))
			builder.WriteString("- Use apenas os dados acima para falar de pagamento, PIX, cobranca e confirmacao.\n")
		}
	}

	if tools.PaymentCreate != nil {
		builder.WriteString("\nRESULTADO DE FERRAMENTA\n")
		builder.WriteString(fmt.Sprintf("- Tool: %s\n", toolNamePaymentCreate))
		builder.WriteString(fmt.Sprintf(
			"- Reserva alvo: booking %s | codigo %s | status %s\n",
			strings.TrimSpace(tools.PaymentCreate.BookingID),
			strings.TrimSpace(tools.PaymentCreate.ReservationCode),
			strings.TrimSpace(tools.PaymentCreate.BookingStatus),
		))
		builder.WriteString(fmt.Sprintf(
			"- Cobranca solicitada: tipo %s | etapa %s | total R$ %.2f | pago R$ %.2f | cobrar R$ %.2f\n",
			strings.TrimSpace(tools.PaymentCreate.PaymentType),
			strings.TrimSpace(tools.PaymentCreate.Stage),
			tools.PaymentCreate.AmountTotal,
			tools.PaymentCreate.AmountPaid,
			tools.PaymentCreate.AmountDue,
		))
		builder.WriteString(fmt.Sprintf("- Resultado operacional: %s\n", strings.TrimSpace(tools.PaymentCreate.Mode)))
		if tools.PaymentCreate.PaymentID != "" {
			builder.WriteString(fmt.Sprintf(
				"- Pagamento criado: id %s | status %s | provedor %s | referencia %s\n",
				strings.TrimSpace(tools.PaymentCreate.PaymentID),
				strings.TrimSpace(tools.PaymentCreate.PaymentStatus),
				strings.TrimSpace(tools.PaymentCreate.Provider),
				strings.TrimSpace(tools.PaymentCreate.ProviderRef),
			))
		}
		if tools.PaymentCreate.PixCode != "" {
			builder.WriteString(fmt.Sprintf("- PIX copia e cola: %s\n", strings.TrimSpace(tools.PaymentCreate.PixCode)))
		}
		for _, item := range tools.PaymentCreate.Errors {
			builder.WriteString(fmt.Sprintf("- Erro operacional: %s\n", strings.TrimSpace(item)))
		}
		if message := strings.TrimSpace(tools.PaymentCreate.MessageForAgent); message != "" {
			builder.WriteString(fmt.Sprintf("- Instrucao operacional: %s\n", message))
		}
		builder.WriteString("- Se houver PIX copia e cola, envie somente o codigo PIX ao cliente e nao envie link do provedor.\n")
	}

	if tools.BookingCancel != nil {
		builder.WriteString("\nRESULTADO DE FERRAMENTA\n")
		builder.WriteString(fmt.Sprintf("- Tool: %s\n", toolNameBookingCancel))
		builder.WriteString(fmt.Sprintf(
			"- Reserva alvo: booking %s | codigo %s | trip %s\n",
			strings.TrimSpace(tools.BookingCancel.BookingID),
			strings.TrimSpace(tools.BookingCancel.ReservationCode),
			strings.TrimSpace(tools.BookingCancel.TripID),
		))
		builder.WriteString(fmt.Sprintf(
			"- Resultado operacional: %s | status anterior %s | status atual %s | passageiros %d\n",
			strings.TrimSpace(tools.BookingCancel.Mode),
			strings.TrimSpace(tools.BookingCancel.PreviousStatus),
			strings.TrimSpace(tools.BookingCancel.BookingStatus),
			tools.BookingCancel.PassengerCount,
		))
		if reason := strings.TrimSpace(tools.BookingCancel.Reason); reason != "" {
			builder.WriteString(fmt.Sprintf("- Motivo operacional: %s\n", reason))
		}
		for _, item := range tools.BookingCancel.Errors {
			builder.WriteString(fmt.Sprintf("- Erro operacional: %s\n", strings.TrimSpace(item)))
		}
		if message := strings.TrimSpace(tools.BookingCancel.MessageForAgent); message != "" {
			builder.WriteString(fmt.Sprintf("- Instrucao operacional: %s\n", message))
		}
		builder.WriteString("- Se o cancelamento ja tiver sido aplicado ou a reserva ja estiver encerrada, responda de forma idempotente e nao prometa nova alteracao.\n")
	}

	builder.WriteString("\nTAREFA\n")
	builder.WriteString("- Responda ao cliente com o proximo passo mais util.\n")
	builder.WriteString("- Se precisar de dado operacional que nao esta no contexto, faca uma pergunta curta em vez de inventar.\n")
	builder.WriteString("- Nao mencione que voce e um agente, modelo ou sistema interno.\n")

	return builder.String()
}

type promptConversationContext struct {
	Origin                   string
	Destination              string
	PackageName              string
	RouteDirection           string
	ShortDateFollowUp        bool
	ShouldRespondWithSCTable bool
	ShouldAskSCOriginForMA   bool
}

func derivePromptConversationContext(currentTurn string, recentMessages []map[string]interface{}) promptConversationContext {
	texts := make([]string, 0, len(recentMessages))
	for _, message := range recentMessages {
		body := strings.TrimSpace(asString(message["body"]))
		if body == "" {
			continue
		}
		texts = append(texts, body)
	}
	currentContext := inferRouteContextFromText(currentTurn)
	historyContext := inferLatestRouteContextFromTexts(texts)
	merged := mergeInferredRouteContext(currentContext, historyContext)
	folded := foldChatText(currentTurn)

	return promptConversationContext{
		Origin:                   merged.Origin,
		Destination:              merged.Destination,
		PackageName:              merged.PackageName,
		RouteDirection:           merged.RouteDirection,
		ShortDateFollowUp:        looksLikeShortDateFollowUp(currentTurn),
		ShouldRespondWithSCTable: detectBroadTravelState(folded) == "SC" && !looksLikeBroadStateScheduleLookup(currentTurn) && currentContext.Origin == "" && currentContext.Destination == "",
		ShouldAskSCOriginForMA:   detectBroadTravelState(folded) == "MA" && !looksLikeBroadStateScheduleLookup(currentTurn) && currentContext.Origin == "" && currentContext.Destination == "",
	}
}

func normalizeRecentMemoryMessages(value interface{}) []map[string]interface{} {
	switch typed := value.(type) {
	case []map[string]interface{}:
		return typed
	case []interface{}:
		items := make([]map[string]interface{}, 0, len(typed))
		for _, raw := range typed {
			message, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			items = append(items, message)
		}
		return items
	default:
		return nil
	}
}

func inferLatestRouteContextFromTexts(texts []string) inferredRouteContext {
	for i := len(texts) - 1; i >= 0; i-- {
		context := inferRouteContextFromText(texts[i])
		if context.Origin != "" || context.Destination != "" || context.PackageName != "" {
			return context
		}
	}
	return inferredRouteContext{}
}
