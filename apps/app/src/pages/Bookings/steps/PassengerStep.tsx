import FormField from "../../../components/FormField";
import { useBookingForm } from "../BookingFormContext";

type PassengerStepProps = {
  stepPassengerComplete: boolean;
  onBack: () => void;
  onNext: () => void;
};

export default function PassengerStep({ stepPassengerComplete, onBack, onNext }: PassengerStepProps) {
  const { form, setForm } = useBookingForm();

  return (
    <>
      <div className="form-step-grid">
        <FormField label="Nome do passageiro" required>
          <input
            className="input"
            placeholder="Ex: Maria Oliveira"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            required
          />
        </FormField>
        <FormField label="Documento" hint="Opcional">
          <input
            className="input"
            placeholder="CPF ou documento"
            value={form.document}
            onChange={(e) => setForm({ ...form, document: e.target.value })}
          />
        </FormField>
        <FormField label="Telefone" hint="Opcional">
          <input
            className="input"
            placeholder="(00) 00000-0000"
            value={form.phone}
            onChange={(e) => setForm({ ...form, phone: e.target.value })}
          />
        </FormField>
        <FormField label="E-mail" hint="Opcional">
          <input
            className="input"
            placeholder="email@exemplo.com"
            value={form.email}
            onChange={(e) => setForm({ ...form, email: e.target.value })}
          />
        </FormField>
      </div>
      <div className="form-step-actions">
        <button className="button secondary" type="button" onClick={onBack}>
          Voltar
        </button>
        <button className="button" type="button" onClick={onNext} disabled={!stepPassengerComplete}>
          Continuar
        </button>
      </div>
    </>
  );
}
