export type TripAdvance = {
  id: string;
  trip_id: string;
  driver_id: string;
  amount: number;
  status: "PENDING" | "DELIVERED" | "SETTLED" | "CANCELLED";
  purpose?: string;
  delivered_at?: string;
  delivered_by?: string;
  settled_at?: string;
  notes?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
};

export type TripExpense = {
  id: string;
  trip_id: string;
  driver_id: string;
  expense_type: "FUEL" | "FOOD" | "LODGING" | "TOLL" | "MAINTENANCE" | "OTHER";
  amount: number;
  description: string;
  expense_date: string;
  payment_method: "ADVANCE" | "CARD" | "PERSONAL" | "COMPANY";
  driver_card_id?: string;
  receipt_number?: string;
  is_approved: boolean;
  approved_by?: string;
  approved_at?: string;
  notes?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
};

export type TripSettlement = {
  id: string;
  trip_id: string;
  driver_id: string;
  status: "DRAFT" | "UNDER_REVIEW" | "APPROVED" | "REJECTED" | "COMPLETED";
  advance_amount: number;
  expenses_total: number;
  balance: number;
  amount_to_return: number;
  amount_to_reimburse: number;
  reviewed_by?: string;
  reviewed_at?: string;
  approved_by?: string;
  approved_at?: string;
  completed_at?: string;
  notes?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
};

export type DriverCard = {
  id: string;
  driver_id: string;
  card_number: string;
  card_type: "FUEL" | "MULTIPURPOSE" | "FOOD";
  current_balance: number;
  is_active: boolean;
  is_blocked: boolean;
  issued_at: string;
  blocked_at?: string;
  blocked_by?: string;
  block_reason?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
};

export type DriverCardTransaction = {
  id: string;
  card_id: string;
  transaction_type: "CREDIT" | "DEBIT" | "ADJUSTMENT" | "REFUND";
  amount: number;
  balance_before: number;
  balance_after: number;
  description?: string;
  trip_expense_id?: string;
  performed_by?: string;
  created_at: string;
};

export type TripValidation = {
  id: string;
  trip_id: string;
  odometer_initial?: number;
  odometer_final?: number;
  distance_km?: number;
  passengers_expected: number;
  passengers_boarded: number;
  passengers_no_show: number;
  validation_notes?: string;
  validated_by?: string;
  validated_at?: string;
  created_at: string;
  updated_at: string;
};

export type FiscalDocument = {
  id: string;
  trip_id: string;
  document_type: string;
  document_number?: string;
  issue_date: string;
  amount: number;
  recipient_name?: string;
  recipient_document?: string;
  status: string;
  external_id?: string;
  metadata?: any;
  created_by?: string;
  created_at: string;
};
