type RoutePublishChecklistProps = {
  missingRequirements: string[];
  isActive: boolean;
};

export default function RoutePublishChecklist({
  missingRequirements,
  isActive,
}: RoutePublishChecklistProps) {
  const canPublish = missingRequirements.length === 0;

  return (
    <div className="route-checklist">
      <div className="section-title">Checklist de publicacao</div>
      <div className="form-hint">
        Autorizações ANTT/DETER e pasta da viagem sao feitas em Operacoes da Viagem.
      </div>
      <ul className="route-checklist-list">
        {canPublish ? (
          <li className="route-checklist-item ok">
            Rota pronta para publicacao.
          </li>
        ) : (
          missingRequirements.map((item) => (
            <li className="route-checklist-item error" key={item}>
              {item}
            </li>
          ))
        )}
      </ul>
      {isActive ? <div className="route-checklist-note">Rota ja publicada.</div> : null}
    </div>
  );
}
