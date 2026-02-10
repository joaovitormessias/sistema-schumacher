import EmptyState from "../../components/EmptyState";
import { Skeleton } from "../../components/feedback/SkeletonLoader";
import SearchToolbar from "../../components/input/SearchToolbar";
import PaginationControls from "../../components/input/PaginationControls";
import StatusBadge, { type StatusTone } from "../../components/StatusBadge";
import DataTable, { type DataTableColumn } from "../../components/table/DataTable";
import VirtualDataTable from "../../components/data-display/VirtualDataTable";
import useMediaQuery from "../../hooks/useMediaQuery";
import { formatCurrency } from "../../utils/format";

type BookingItem = {
  id: string;
  trip_id: string;
  status: string;
  passenger_name: string;
  passenger_email: string;
  seat_number: number;
  total_amount: number;
};

type BookingListProps = {
  bookings: BookingItem[];
  loading: boolean;
  query: string;
  onQueryChange: (value: string) => void;
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  tripLabel: (tripId: string) => string;
};

const statusTone = (status: string): StatusTone => {
  if (status === "CONFIRMED") return "success";
  if (status === "PENDING") return "warning";
  if (status === "CANCELLED") return "danger";
  return "neutral";
};

export default function BookingList({
  bookings,
  loading,
  query,
  onQueryChange,
  page,
  pageSize,
  onPageChange,
  tripLabel,
}: BookingListProps) {
  const isMobile = useMediaQuery("(max-width: 900px)");
  const columns: DataTableColumn<BookingItem>[] = [
    { label: "Passageiro", accessor: (booking) => booking.passenger_name },
    {
      label: "Viagem",
      accessor: (booking) => tripLabel(booking.trip_id),
      hideOnMobile: true,
    },
    { label: "Poltrona", accessor: (booking) => booking.seat_number, width: "80px" },
    {
      label: "Status",
      render: (booking) => (
        <StatusBadge tone={statusTone(booking.status)}>{booking.status}</StatusBadge>
      ),
      width: "140px",
    },
    {
      label: "Total",
      accessor: (booking) => formatCurrency(booking.total_amount),
      align: "right",
      width: "120px",
    },
  ];
  const shouldVirtualize = !isMobile && bookings.length > 100;
  const tableElement = shouldVirtualize ? (
    <VirtualDataTable
      columns={columns}
      rows={bookings}
      rowKey={(booking) => booking.id}
      emptyState={
        <EmptyState
          title="Nenhuma reserva encontrada"
          description="Tente ajustar a busca ou crie uma nova reserva."
        />
      }
    />
  ) : (
    <DataTable
      columns={columns}
      rows={bookings}
      rowKey={(booking) => booking.id}
      emptyState={
        <EmptyState
          title="Nenhuma reserva encontrada"
          description="Tente ajustar a busca ou crie uma nova reserva."
        />
      }
    />
  );

  return (
    <div className="section">
      <div className="section-header">
        <div className="section-title">Reservas cadastradas</div>
      </div>
      <SearchToolbar
        value={query}
        onChange={onQueryChange}
        placeholder="Buscar por passageiro, e-mail, poltrona ou status"
        inputLabel="Buscar reservas"
        resultCount={bookings.length}
      />

      <PaginationControls
        page={page}
        pageSize={pageSize}
        itemCount={bookings.length}
        onPageChange={onPageChange}
        disabled={loading}
      />

      {loading ? <Skeleton.Table rows={6} columns={columns.length} /> : tableElement}
    </div>
  );
}
