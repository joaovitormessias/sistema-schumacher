import { type PropsWithChildren, type ReactNode } from "react";
import Modal from "./Modal";

type DrawerProps = PropsWithChildren<{
  open: boolean;
  title: string;
  description?: string;
  onClose: () => void;
  side?: "left" | "right";
  footer?: ReactNode;
}>;

export default function Drawer({
  open,
  title,
  description,
  onClose,
  side = "right",
  footer,
  children,
}: DrawerProps) {
  return (
    <Modal
      open={open}
      title={title}
      description={description}
      onClose={onClose}
      footer={footer}
      size="lg"
    >
      <div className={`overlay-drawer overlay-drawer-${side}`}>{children}</div>
    </Modal>
  );
}
