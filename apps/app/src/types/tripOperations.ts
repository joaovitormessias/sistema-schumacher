export type OperationalStatus =
  | "REQUESTED"
  | "PASSENGERS_READY"
  | "ITINERARY_READY"
  | "DISPATCH_VALIDATED"
  | "AUTHORIZED"
  | "IN_PROGRESS"
  | "RETURNED"
  | "RETURN_CHECKED"
  | "SETTLED"
  | "CLOSED";

export type TripOperationsTrip = {
  id: string;
  status: string;
  operational_status: OperationalStatus;
  departure_at: string;
  estimated_km: number;
  dispatch_validated_at?: string;
  dispatch_validated_by?: string;
};

export type TripRequest = {
  id: string;
  route_id?: string;
  source: "EMAIL" | "SYSTEM";
  status: "OPEN" | "IN_REVIEW" | "APPROVED" | "REJECTED";
  requester_name?: string;
  requester_contact?: string;
  requested_departure_at?: string;
  notes?: string;
  created_at: string;
};

export type TripManifestEntry = {
  id: string;
  trip_id: string;
  booking_passenger_id?: string;
  passenger_name: string;
  passenger_document?: string;
  passenger_phone?: string;
  source: "BOOKING" | "MANUAL";
  status: "EXPECTED" | "BOARDED" | "NO_SHOW" | "CANCELLED";
  seat_number?: number;
  is_active: boolean;
};

export type TripAuthorization = {
  id: string;
  trip_id: string;
  authority: "ANTT" | "DETER" | "EXCEPTIONAL";
  status: "PENDING" | "ISSUED" | "REJECTED" | "EXPIRED";
  protocol_number?: string;
  license_number?: string;
  issued_at?: string;
  valid_until?: string;
  src_policy_number?: string;
  src_valid_until?: string;
  exceptional_deadline_ok: boolean;
  notes?: string;
};

export type TripChecklist = {
  id: string;
  trip_id: string;
  stage: "PRE_DEPARTURE" | "RETURN";
  checklist_data: Record<string, unknown>;
  is_complete: boolean;
  documents_checked: boolean;
  tachograph_checked: boolean;
  receipts_checked: boolean;
  rest_compliance_ok: boolean;
  notes?: string;
};

export type TripDriverReport = {
  id: string;
  trip_id: string;
  driver_id?: string;
  odometer_start?: number;
  odometer_end?: number;
  fuel_used_liters?: number;
  incidents?: string;
  delays?: string;
  rest_hours?: number;
  notes?: string;
};

export type TripReconciliation = {
  id: string;
  trip_id: string;
  total_receipts_amount: number;
  total_approved_expenses: number;
  difference: number;
  receipts_validated: boolean;
  verified_expense_ids: string[];
  notes?: string;
};

export type TripAttachment = {
  id: string;
  trip_id: string;
  attachment_type: string;
  storage_bucket: string;
  storage_path: string;
  file_name: string;
  mime_type?: string;
  file_size?: number;
  uploaded_at: string;
};

export type WorkflowBlockedResponse = {
  code: string;
  message: string;
  requirements_missing: string[];
};
