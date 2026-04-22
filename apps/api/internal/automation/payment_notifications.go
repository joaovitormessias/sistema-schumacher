package automation

import (
	"fmt"
	"math"
	"schumacher-tur/api/internal/payments"
	"strconv"
	"strings"
	"time"
)

type paymentNotificationPreview struct {
	CustomerPhone string
	ContactKey    string
	Messages      []string
	Body          string
}

func buildPaymentNotificationPayload(payment payments.Payment, notification payments.PaymentNotificationContext, observedAt time.Time) payments.PaymentNotificationPayload {
	return payments.PaymentNotificationPayload{
		Event:           "payment.confirmed",
		SentAt:          observedAt.UTC(),
		PaymentID:       payment.ID,
		PaymentAmount:   payment.Amount,
		PaymentMethod:   payment.Method,
		BookingID:       notification.BookingID,
		ReservationCode: notification.ReservationCode,
		CustomerName:    notification.CustomerName,
		CustomerPhone:   notification.CustomerPhone,
		AmountTotal:     notification.AmountTotal,
		AmountPaid:      notification.AmountPaid,
		AmountDue:       notification.AmountDue,
		PaymentStatus:   notification.PaymentStatus,
	}
}

func buildPaymentNotificationPreview(payload payments.PaymentNotificationPayload) paymentNotificationPreview {
	reservationCode := strings.TrimSpace(payload.ReservationCode)
	customerPhone := normalizePhone(payload.CustomerPhone)
	contactKey := buildPaymentNotificationContactKey(customerPhone)
	amountDue := maxFloat(payload.AmountDue, 0)
	paymentAmount := payload.PaymentAmount
	if paymentAmount <= 0 {
		paymentAmount = payload.AmountPaid
	}

	messages := make([]string, 0, 4)
	switch {
	case reservationCode != "" && amountDue > 0.009:
		messages = append(messages, fmt.Sprintf("Pagamento do sinal confirmado para a reserva %s. Valor recebido: %s.", reservationCode, formatBRL(paymentAmount)))
	case reservationCode != "":
		messages = append(messages, fmt.Sprintf("Pagamento confirmado com sucesso para a reserva %s. Sua reserva esta quitada.", reservationCode))
	case amountDue > 0.009:
		messages = append(messages, fmt.Sprintf("Pagamento do sinal confirmado. Valor recebido: %s.", formatBRL(paymentAmount)))
	default:
		messages = append(messages, "Pagamento confirmado com sucesso. Sua reserva esta quitada.")
	}

	messages = append(messages, "No embarque, tenha em maos o documento informado na reserva de todos os passageiros.")
	messages = append(messages, "Se houver menor de idade viajando com voce e voce nao for o responsavel legal, e obrigatorio apresentar autorizacao ou documento registrado em cartorio no embarque.")

	if amountDue > 0.009 {
		messages = append(messages, fmt.Sprintf("Atencao: no embarque havera cobranca do valor restante da reserva. Valor total: %s. Saldo pendente: %s.", formatBRL(payload.AmountTotal), formatBRL(amountDue)))
	}

	body := strings.Join(messages, "\n\n")
	if contactKey == "" {
		messages = nil
		body = ""
	}

	return paymentNotificationPreview{
		CustomerPhone: customerPhone,
		ContactKey:    contactKey,
		Messages:      messages,
		Body:          body,
	}
}

func buildPaymentNotificationDraftKey(paymentID string) string {
	return "payment-notification-" + strings.TrimSpace(paymentID)
}

func buildPaymentNotificationContactKey(phone string) string {
	digits := normalizePhone(phone)
	if digits == "" {
		return ""
	}
	if !strings.HasPrefix(digits, "55") {
		digits = "55" + digits
	}
	return digits + "@s.whatsapp.net"
}

func formatBRL(amount float64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = math.Abs(amount)
	}

	cents := int64(math.Round(amount * 100))
	whole := cents / 100
	fraction := cents % 100

	wholeText := strconv.FormatInt(whole, 10)
	if len(wholeText) > 3 {
		parts := make([]string, 0, (len(wholeText)+2)/3)
		for len(wholeText) > 3 {
			parts = append([]string{wholeText[len(wholeText)-3:]}, parts...)
			wholeText = wholeText[:len(wholeText)-3]
		}
		parts = append([]string{wholeText}, parts...)
		wholeText = strings.Join(parts, ".")
	}

	return fmt.Sprintf("%sR$ %s,%02d", sign, wholeText, fraction)
}

func maxFloat(left float64, right float64) float64 {
	if left > right {
		return left
	}
	return right
}
