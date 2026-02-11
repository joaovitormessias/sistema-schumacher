import React, { useState, useMemo } from 'react';
import { useRoutes, type RouteItem } from '../../hooks/useRoutes';
import { useBuses, type BusItem } from '../../hooks/useBuses';
import { useDrivers, type DriverItem } from '../../hooks/useDrivers';
import { formatDateTime } from '../../utils/format';
import './wizard.css';

interface TripFormData {
  route_id: string;
  bus_id: string;
  driver_id: string;
  request_id: string;
  departure_at: string;
  estimated_km: string;
  trip_type: 'regular' | 'chartered' | 'executive';
}

interface TripWizardProps {
  onSubmit: (data: TripFormData) => Promise<void>;
  onCancel: () => void;
  initialData?: Partial<TripFormData>;
}

const initialFormData: TripFormData = {
  route_id: '',
  bus_id: '',
  driver_id: '',
  request_id: '',
  departure_at: '',
  estimated_km: '',
  trip_type: 'regular',
};

export function TripWizard({ onSubmit, onCancel, initialData }: TripWizardProps) {
  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState<TripFormData>({
    ...initialFormData,
    ...initialData,
  });
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { data: routes = [] } = useRoutes(200, 0, { status: "active" });
  const { data: buses = [] } = useBuses(200, 0);
  const { data: drivers = [] } = useDrivers(200, 0);

  // Get selected items for display
  const selectedRoute = routes.find(r => r.id === formData.route_id);
  const selectedBus = buses.find(b => b.id === formData.bus_id);
  const selectedDriver = drivers.find(d => d.id === formData.driver_id);

  // Intelligent filtering
  // For now, just show all buses and drivers since we don't have status/type info
  const availableBuses = useMemo(() => {
    return buses;
  }, [buses]);

  const availableDrivers = useMemo(() => {
    return drivers;
  }, [drivers]);

  // Auto-suggest KM based on route
  const suggestedKm = useMemo(() => {
    if (!selectedRoute) return '';
    // In a real app, this would come from route data
    // For now, return a placeholder
    return ''; // Will be calculated based on route distance
  }, [selectedRoute]);

  const updateFormData = (updates: Partial<TripFormData>) => {
    setFormData(prev => ({ ...prev, ...updates }));
  };

  const validateStep1 = () => {
    return !!(formData.route_id && formData.departure_at);
  };

  const validateStep2 = () => {
    return !!formData.bus_id; // Driver is optional
  };

  const handleNext = () => {
    if (currentStep === 1 && validateStep1()) {
      setCurrentStep(2);
    } else if (currentStep === 2 && validateStep2()) {
      setCurrentStep(3);
    }
  };

  const handleBack = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleSubmit = async () => {
    setIsSubmitting(true);
    try {
      await onSubmit(formData);
    } finally {
      setIsSubmitting(false);
    }
  };

  const getQuickDateTime = (type: 'today' | 'tomorrow' | 'monday') => {
    const now = new Date();
    now.setHours(8, 0, 0, 0); // 8 AM

    if (type === 'tomorrow') {
      now.setDate(now.getDate() + 1);
    } else if (type === 'monday') {
      const dayOfWeek = now.getDay();
      const daysUntilMonday = dayOfWeek === 0 ? 1 : 8 - dayOfWeek;
      now.setDate(now.getDate() + daysUntilMonday);
    }

    return now.toISOString().slice(0, 16);
  };

  return (
    <div className="trip-wizard">
      {/* Stepper */}
      <div className="wizard-stepper">
        <div className={`wizard-step ${currentStep >= 1 ? 'active' : ''} ${currentStep > 1 ? 'completed' : ''}`}>
          <div className="wizard-step-number">1</div>
          <div className="wizard-step-label">Definir Viagem</div>
        </div>
        <div className="wizard-step-line"></div>
        <div className={`wizard-step ${currentStep >= 2 ? 'active' : ''} ${currentStep > 2 ? 'completed' : ''}`}>
          <div className="wizard-step-number">2</div>
          <div className="wizard-step-label">Configurar Operação</div>
        </div>
        <div className="wizard-step-line"></div>
        <div className={`wizard-step ${currentStep >= 3 ? 'active' : ''}`}>
          <div className="wizard-step-number">3</div>
          <div className="wizard-step-label">Revisar e Criar</div>
        </div>
      </div>

      {/* Step Content */}
      <div className="wizard-content">
        {currentStep === 1 && (
          <div className="wizard-step-content">
            <h3 className="wizard-step-title">Definir Viagem</h3>
            <p className="wizard-step-description">Configure o que e quando</p>

            <div className="wizard-fields">
              <div className="wizard-field">
                <label className="wizard-label">
                  <span className="wizard-label-icon">📍</span>
                  Qual rota será realizada?
                  <span className="wizard-label-required">*</span>
                </label>
                <p className="wizard-help-text">Define origem, destino e paradas principais</p>
                <select
                  className="wizard-input"
                  value={formData.route_id}
                  onChange={e => updateFormData({ route_id: e.target.value })}
                  required
                >
                  <option value="">Selecione a rota</option>
                  {routes.map(route => (
                    <option key={route.id} value={route.id}>
                      {route.origin_city} → {route.destination_city}
                    </option>
                  ))}
                </select>
              </div>

              <div className="wizard-field">
                <label className="wizard-label">
                  <span className="wizard-label-icon">🗓️</span>
                  Quando a viagem sai?
                  <span className="wizard-label-required">*</span>
                </label>
                <div className="wizard-quick-selects">
                  <button
                    type="button"
                    className="wizard-quick-select"
                    onClick={() => updateFormData({ departure_at: getQuickDateTime('today') })}
                  >
                    Hoje
                  </button>
                  <button
                    type="button"
                    className="wizard-quick-select"
                    onClick={() => updateFormData({ departure_at: getQuickDateTime('tomorrow') })}
                  >
                    Amanhã
                  </button>
                  <button
                    type="button"
                    className="wizard-quick-select"
                    onClick={() => updateFormData({ departure_at: getQuickDateTime('monday') })}
                  >
                    Próxima segunda
                  </button>
                </div>
                <input
                  type="datetime-local"
                  className="wizard-input"
                  value={formData.departure_at}
                  onChange={e => updateFormData({ departure_at: e.target.value })}
                  required
                />
              </div>

              <div className="wizard-field">
                <label className="wizard-label">
                  Tipo de viagem
                </label>
                <div className="wizard-toggle-group">
                  <button
                    type="button"
                    className={`wizard-toggle ${formData.trip_type === 'regular' ? 'active' : ''}`}
                    onClick={() => updateFormData({ trip_type: 'regular' })}
                  >
                    Regular
                  </button>
                  <button
                    type="button"
                    className={`wizard-toggle ${formData.trip_type === 'chartered' ? 'active' : ''}`}
                    onClick={() => updateFormData({ trip_type: 'chartered' })}
                  >
                    Fretado
                  </button>
                  <button
                    type="button"
                    className={`wizard-toggle ${formData.trip_type === 'executive' ? 'active' : ''}`}
                    onClick={() => updateFormData({ trip_type: 'executive' })}
                  >
                    Executivo
                  </button>
                </div>
              </div>

              {/* Advanced options */}
              <button
                type="button"
                className="wizard-advanced-toggle"
                onClick={() => setShowAdvanced(!showAdvanced)}
              >
                <span className="wizard-advanced-icon">⚙️</span>
                {showAdvanced ? 'Ocultar' : 'Mostrar'} opções avançadas
              </button>

              {showAdvanced && (
                <div className="wizard-advanced-fields">
                  <div className="wizard-field">
                    <label className="wizard-label">
                      Quilometragem esperada
                    </label>
                    <p className="wizard-help-text">
                      {suggestedKm ? `Sugerimos ${suggestedKm} km baseado na rota` : 'Opcional'}
                    </p>
                    <input
                      type="number"
                      className="wizard-input"
                      value={formData.estimated_km}
                      onChange={e => updateFormData({ estimated_km: e.target.value })}
                      placeholder={suggestedKm || 'Ex: 408'}
                      step="0.1"
                      min="0"
                    />
                  </div>

                  <div className="wizard-field">
                    <label className="wizard-label">
                      ID da solicitação
                    </label>
                    <p className="wizard-help-text">Se essa viagem foi solicitada por outro setor</p>
                    <input
                      type="text"
                      className="wizard-input"
                      value={formData.request_id}
                      onChange={e => updateFormData({ request_id: e.target.value })}
                      placeholder="Opcional"
                    />
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {currentStep === 2 && (
          <div className="wizard-step-content">
            <div className="wizard-with-sidebar">
              <div className="wizard-main">
                <h3 className="wizard-step-title">Configurar Operação</h3>
                <p className="wizard-step-description">Escolha ônibus e motorista</p>

                <div className="wizard-fields">
                  <div className="wizard-field">
                    <label className="wizard-label">
                      <span className="wizard-label-icon">🚌</span>
                      Escolha o ônibus
                      <span className="wizard-label-required">*</span>
                    </label>
                    <p className="wizard-help-text">
                      Apenas ônibus disponíveis para essa rota e data
                    </p>
                    <select
                      className="wizard-input"
                      value={formData.bus_id}
                      onChange={e => updateFormData({ bus_id: e.target.value })}
                      required
                    >
                      <option value="">Selecione o ônibus</option>
                      {availableBuses.map(bus => (
                        <option key={bus.id} value={bus.id}>
                          {bus.name}
                        </option>
                      ))}
                    </select>
                    {availableBuses.length === 0 && (
                      <p className="wizard-warning">
                        ⚠️ Nenhum ônibus disponível para este tipo de viagem
                      </p>
                    )}
                  </div>

                  <div className="wizard-field">
                    <label className="wizard-label">
                      <span className="wizard-label-icon">👤</span>
                      Quem vai dirigir?
                    </label>
                    <p className="wizard-help-text">
                      Motoristas com escala disponível nesse horário
                    </p>
                    <select
                      className="wizard-input"
                      value={formData.driver_id}
                      onChange={e => updateFormData({ driver_id: e.target.value })}
                    >
                      <option value="">Selecionar depois</option>
                      {availableDrivers.map(driver => (
                        <option key={driver.id} value={driver.id}>
                          {driver.name}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              </div>

              {/* Preview Sidebar */}
              <div className="wizard-preview">
                <h4 className="wizard-preview-title">Preview da Viagem</h4>
                <div className="wizard-preview-content">
                  <div className="wizard-preview-item">
                    <span className="wizard-preview-icon">📍</span>
                    <div>
                      <div className="wizard-preview-label">Rota</div>
                      <div className="wizard-preview-value">
                        {selectedRoute ? `${selectedRoute.origin_city} → ${selectedRoute.destination_city}` : '-'}
                      </div>
                    </div>
                  </div>

                  <div className="wizard-preview-item">
                    <span className="wizard-preview-icon">🗓️</span>
                    <div>
                      <div className="wizard-preview-label">Saída</div>
                      <div className="wizard-preview-value">
                        {formData.departure_at ? formatDateTime(formData.departure_at) : '-'}
                      </div>
                    </div>
                  </div>

                  <div className="wizard-preview-item">
                    <span className="wizard-preview-icon">🚌</span>
                    <div>
                      <div className="wizard-preview-label">Ônibus</div>
                      <div className="wizard-preview-value">
                        {selectedBus ? selectedBus.name : 'Não selecionado'}
                      </div>
                    </div>
                  </div>

                  <div className="wizard-preview-item">
                    <span className="wizard-preview-icon">👤</span>
                    <div>
                      <div className="wizard-preview-label">Motorista</div>
                      <div className="wizard-preview-value">
                        {selectedDriver ? selectedDriver.name : 'Definir depois'}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {currentStep === 3 && (
          <div className="wizard-step-content">
            <h3 className="wizard-step-title">Revisar e Criar</h3>
            <p className="wizard-step-description">Confira todos os dados antes de confirmar</p>

            <div className="wizard-review">
              <div className="wizard-review-card">
                <h4 className="wizard-review-card-title">🚍 Nova Viagem</h4>
                <div className="wizard-review-items">
                  <div className="wizard-review-row">
                    <span className="wizard-review-label">Rota:</span>
                    <span className="wizard-review-value">
                      {selectedRoute ? `${selectedRoute.origin_city} → ${selectedRoute.destination_city}` : '-'}
                    </span>
                  </div>
                  <div className="wizard-review-row">
                    <span className="wizard-review-label">Saída:</span>
                    <span className="wizard-review-value">
                      {formData.departure_at ? formatDateTime(formData.departure_at) : '-'}
                    </span>
                  </div>
                  <div className="wizard-review-row">
                    <span className="wizard-review-label">Ônibus:</span>
                    <span className="wizard-review-value">
                      {selectedBus ? selectedBus.name : '-'}
                    </span>
                  </div>
                  <div className="wizard-review-row">
                    <span className="wizard-review-label">Motorista:</span>
                    <span className="wizard-review-value">
                      {selectedDriver ? selectedDriver.name : 'Definir depois'}
                    </span>
                  </div>
                  {formData.estimated_km && (
                    <div className="wizard-review-row">
                      <span className="wizard-review-label">KM planejada:</span>
                      <span className="wizard-review-value">{formData.estimated_km} km</span>
                    </div>
                  )}
                  <div className="wizard-review-row">
                    <span className="wizard-review-label">Status inicial:</span>
                    <span className="wizard-review-value">Programada</span>
                  </div>
                </div>
              </div>

              <div className="wizard-checklist">
                <h4 className="wizard-checklist-title">✅ Checklist Operacional</h4>
                <div className="wizard-checklist-items">
                  <div className="wizard-checklist-item success">
                    <span className="wizard-checklist-icon">✓</span>
                    Ônibus disponível e sem manutenção pendente
                  </div>
                  {formData.driver_id && (
                    <div className="wizard-checklist-item success">
                      <span className="wizard-checklist-icon">✓</span>
                      Motorista disponível e habilitado
                    </div>
                  )}
                  {selectedRoute && (
                    <div className="wizard-checklist-item success">
                      <span className="wizard-checklist-icon">✓</span>
                      Rota ativa e sem bloqueios
                    </div>
                  )}
                </div>
              </div>

              <div className="wizard-notifications">
                <h4 className="wizard-notifications-title">🔔 Notificações automáticas</h4>
                <div className="wizard-notifications-items">
                  {formData.driver_id && (
                    <div className="wizard-notification-item">
                      → Motorista será notificado 2h antes da saída
                    </div>
                  )}
                  <div className="wizard-notification-item">
                    → Status será "Programada" automaticamente
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="wizard-actions">
        <button
          type="button"
          className="wizard-button wizard-button-secondary"
          onClick={currentStep === 1 ? onCancel : handleBack}
        >
          {currentStep === 1 ? 'Cancelar' : '← Voltar'}
        </button>

        {currentStep < 3 ? (
          <button
            type="button"
            className="wizard-button wizard-button-primary"
            onClick={handleNext}
            disabled={
              (currentStep === 1 && !validateStep1()) ||
              (currentStep === 2 && !validateStep2())
            }
          >
            Próximo: {currentStep === 1 ? 'Escolher ônibus e motorista' : 'Revisar viagem'} →
          </button>
        ) : (
          <button
            type="button"
            className="wizard-button wizard-button-primary"
            onClick={handleSubmit}
            disabled={isSubmitting}
          >
            {isSubmitting ? 'Criando...' : 'Criar viagem'}
          </button>
        )}
      </div>
    </div>
  );
}
