-- Migration: Almoxarifado e Compras
-- Baseado em: docs-negocios/Procedimentos Operacionais Almoxarifado e Compras.docx

-- ENUMS
CREATE TYPE service_order_type AS ENUM ('PREVENTIVE', 'CORRECTIVE');
CREATE TYPE service_order_status AS ENUM ('OPEN', 'IN_PROGRESS', 'CLOSED', 'CANCELLED');
CREATE TYPE purchase_order_status AS ENUM ('DRAFT', 'SENT', 'PARTIAL', 'RECEIVED', 'CANCELLED');
CREATE TYPE stock_movement_type AS ENUM ('IN', 'OUT', 'ADJUSTMENT');
CREATE TYPE invoice_status AS ENUM ('PENDING', 'PROCESSED', 'CANCELLED');

-- 1. Fornecedores
CREATE TABLE suppliers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  document TEXT, -- CNPJ
  phone TEXT,
  email TEXT,
  payment_terms TEXT, -- 'CASH', 'INVOICED'
  billing_day INT, -- dia do faturamento (ex: 10)
  is_active BOOLEAN DEFAULT true,
  notes TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- 2. Produtos (peças, serviços, combustível)
CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  category TEXT, -- 'PART', 'SERVICE', 'FUEL', 'OTHER'
  unit TEXT DEFAULT 'UN', -- UN, LT, KG
  min_stock NUMERIC(10,2) DEFAULT 0,
  current_stock NUMERIC(10,2) DEFAULT 0,
  last_cost NUMERIC(10,2),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- 3. Ordens de Serviço (Preventiva/Corretiva)
-- Fluxo: Gilmar identifica problema → abre OS → cria pedidos → lança NFs → fecha OS
CREATE TABLE service_orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_number SERIAL,
  bus_id UUID NOT NULL REFERENCES buses(id),
  driver_id UUID REFERENCES drivers(id),
  order_type service_order_type NOT NULL,
  status service_order_status DEFAULT 'OPEN',
  description TEXT NOT NULL,
  odometer_km INT,
  scheduled_date DATE, -- para preventivas
  location TEXT DEFAULT 'SCHUMACHER', -- 'SCHUMACHER', 'EXTERNAL'
  opened_at TIMESTAMPTZ DEFAULT now(),
  closed_at TIMESTAMPTZ,
  closed_odometer_km INT,
  next_preventive_km INT, -- programar próxima revisão
  notes TEXT,
  created_by UUID,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- 4. Pedidos de Compra (criados dentro da OS)
-- Fluxo: dentro da OS → gerar pedido → fornecedor → data entrega → itens
CREATE TABLE purchase_orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_number SERIAL,
  service_order_id UUID REFERENCES service_orders(id),
  supplier_id UUID NOT NULL REFERENCES suppliers(id),
  status purchase_order_status DEFAULT 'DRAFT',
  order_date DATE DEFAULT CURRENT_DATE,
  expected_delivery DATE,
  own_delivery BOOLEAN DEFAULT true, -- "entrega própria sempre sim"
  subtotal NUMERIC(10,2) DEFAULT 0,
  discount NUMERIC(10,2) DEFAULT 0,
  freight NUMERIC(10,2) DEFAULT 0,
  total NUMERIC(10,2) DEFAULT 0,
  notes TEXT,
  created_by UUID,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- 5. Itens do Pedido de Compra
CREATE TABLE purchase_order_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
  product_id UUID NOT NULL REFERENCES products(id),
  quantity NUMERIC(10,3) NOT NULL,
  unit_price NUMERIC(10,2) NOT NULL,
  discount NUMERIC(10,2) DEFAULT 0,
  total NUMERIC(10,2) NOT NULL,
  received_quantity NUMERIC(10,3) DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- 6. Notas Fiscais de Entrada
-- Fluxo: código de barras → data emissão → CFOP 1000 → placa → itens → salva
CREATE TABLE invoices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_number TEXT NOT NULL,
  barcode TEXT,
  supplier_id UUID NOT NULL REFERENCES suppliers(id),
  purchase_order_id UUID REFERENCES purchase_orders(id),
  service_order_id UUID REFERENCES service_orders(id),
  bus_id UUID REFERENCES buses(id), -- para abastecimento
  issue_date DATE NOT NULL,
  issue_time TIME,
  entry_date DATE DEFAULT CURRENT_DATE,
  entry_time TIME DEFAULT CURRENT_TIME,
  cfop TEXT DEFAULT '1000',
  payment_type TEXT, -- 'CASH', 'INVOICED'
  due_date DATE,
  subtotal NUMERIC(10,2) NOT NULL,
  discount NUMERIC(10,2) DEFAULT 0,
  freight NUMERIC(10,2) DEFAULT 0,
  total NUMERIC(10,2) NOT NULL,
  status invoice_status DEFAULT 'PENDING',
  notes TEXT, -- dados adicionais: placa, outras informações
  driver_id UUID REFERENCES drivers(id),
  odometer_km INT, -- para abastecimento
  created_by UUID,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- 7. Itens da NF
CREATE TABLE invoice_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
  product_id UUID NOT NULL REFERENCES products(id),
  quantity NUMERIC(10,3) NOT NULL,
  unit_price NUMERIC(10,2) NOT NULL,
  discount NUMERIC(10,2) DEFAULT 0,
  total NUMERIC(10,2) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- 8. Movimentações de Estoque
CREATE TABLE stock_movements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id UUID NOT NULL REFERENCES products(id),
  movement_type stock_movement_type NOT NULL,
  quantity NUMERIC(10,3) NOT NULL,
  unit_cost NUMERIC(10,2),
  reference_type TEXT, -- 'INVOICE', 'SERVICE_ORDER', 'ADJUSTMENT'
  reference_id UUID,
  balance_before NUMERIC(10,2),
  balance_after NUMERIC(10,2),
  notes TEXT,
  performed_by UUID,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- ÍNDICES para performance
CREATE INDEX idx_service_orders_bus ON service_orders(bus_id);
CREATE INDEX idx_service_orders_status ON service_orders(status);
CREATE INDEX idx_service_orders_type ON service_orders(order_type);
CREATE INDEX idx_purchase_orders_supplier ON purchase_orders(supplier_id);
CREATE INDEX idx_purchase_orders_service ON purchase_orders(service_order_id);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX idx_invoices_supplier ON invoices(supplier_id);
CREATE INDEX idx_invoices_purchase ON invoices(purchase_order_id);
CREATE INDEX idx_invoices_service ON invoices(service_order_id);
CREATE INDEX idx_invoices_bus ON invoices(bus_id);
CREATE INDEX idx_stock_movements_product ON stock_movements(product_id);
CREATE INDEX idx_stock_movements_type ON stock_movements(movement_type);

-- Dados iniciais: Categorias de produtos comuns
INSERT INTO products (code, name, category, unit) VALUES
  ('9765', 'Diesel S10', 'FUEL', 'LT'),
  ('OLEO-MOTOR', 'Óleo de Motor', 'PART', 'LT'),
  ('FILTRO-OLEO', 'Filtro de Óleo', 'PART', 'UN'),
  ('FILTRO-DIESEL', 'Filtro de Diesel', 'PART', 'UN'),
  ('FILTRO-AR', 'Filtro de Ar', 'PART', 'UN'),
  ('PNEU-275', 'Pneu 275/80 R22.5', 'PART', 'UN'),
  ('PASTILHA-FREIO', 'Pastilha de Freio', 'PART', 'JG'),
  ('LONA-FREIO', 'Lona de Freio', 'PART', 'JG'),
  ('SERVICO-MAO-OBRA', 'Mão de Obra Mecânica', 'SERVICE', 'HR'),
  ('SERVICO-BORRACHARIA', 'Serviço de Borracharia', 'SERVICE', 'UN');

-- Fornecedores iniciais conforme documento
INSERT INTO suppliers (name, payment_terms, billing_day, notes) VALUES
  ('Frai Peças', 'INVOICED', 10, 'Volume alto de compras'),
  ('Piata', 'INVOICED', 10, 'Faturado'),
  ('Borracharia Mifrai', 'INVOICED', 10, 'Faturado'),
  ('Kuko Elétrica', 'INVOICED', 10, 'Faturado'),
  ('Posto Maca', 'CASH', NULL, 'Vale combustível');
