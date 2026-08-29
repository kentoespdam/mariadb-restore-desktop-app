-- ============================================================
-- Test Dump — for ujicoba MariaDB Restore Desktop App
-- Multi-database dump with mixed DDL + DML + edge cases
-- ============================================================

-- ------------------------------------------------------------
-- Database: shop
-- ------------------------------------------------------------
CREATE DATABASE IF NOT EXISTS `shop`
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE `shop`;

-- ---- DDL ----

CREATE TABLE `shop`.`customers` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL,
  `email` VARCHAR(255) DEFAULT NULL,
  `phone` VARCHAR(20) DEFAULT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `shop`.`products` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `sku` VARCHAR(30) NOT NULL,
  `name` VARCHAR(200) NOT NULL,
  `price` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  `stock` INT NOT NULL DEFAULT 0,
  `category` ENUM('electronics','clothing','food','other') NOT NULL DEFAULT 'other',
  `description` TEXT DEFAULT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_sku` (`sku`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `shop`.`orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `customer_id` INT UNSIGNED NOT NULL,
  `order_date` DATE NOT NULL,
  `status` ENUM('pending','paid','shipped','completed','cancelled') NOT NULL DEFAULT 'pending',
  `total` DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  `notes` TEXT DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_customer` (`customer_id`),
  KEY `idx_status` (`status`),
  CONSTRAINT `fk_orders_customer` FOREIGN KEY (`customer_id`) REFERENCES `customers` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `shop`.`order_items` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL,
  `product_id` INT UNSIGNED NOT NULL,
  `qty` INT UNSIGNED NOT NULL DEFAULT 1,
  `unit_price` DECIMAL(10,2) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order` (`order_id`),
  KEY `idx_product` (`product_id`),
  CONSTRAINT `fk_items_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`),
  CONSTRAINT `fk_items_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---- DML ----

INSERT INTO `shop`.`customers` (`name`, `email`, `phone`) VALUES
  ('Budi Santoso', 'budi@example.com', '+62812345678'),
  ('Siti Rahayu', 'siti@example.com', '+62812987654'),
  ('Andi Wijaya', 'andi@example.com', NULL),
  ('Rina Kusuma', 'rina@example.com', '+628135551234'),
  ('Dedi Prasetyo', 'dedi@example.com', '+628146661234');

INSERT INTO `shop`.`products` (`sku`, `name`, `price`, `stock`, `category`, `description`) VALUES
  ('ELEC-001', 'Laptop ASUS VivoBook 14', 8499000.00, 25, 'electronics', '14 inch FHD, Ryzen 5, 16GB RAM'),
  ('ELEC-002', 'Mouse Logitech MX Master 3', 1299000.00, 150, 'electronics', 'Wireless ergonomic mouse'),
  ('CLOTH-001', 'Kaos Katun Premium', 129000.00, 500, 'clothing', 'Ukuran S-XL, warna tersedia'),
  ('CLOTH-002', 'Celana Jeans Slim Fit', 299000.00, 200, 'clothing', 'Denim stretch, waist 28-36'),
  ('FOOD-001', 'Kopi Arabica Toraja 250g', 85000.00, 300, 'food', 'Roasted bean, origin Sulawesi'),
  ('FOOD-002', 'Madu Hutan Asli 500ml', 125000.00, 180, 'food', '100% pure honey, no additives'),
  ('OTHER-001', 'Tas Ransel Anti-Air', 199000.00, 90, 'other', 'Kapasitas 25L, cocok untuk laptop');

INSERT INTO `shop`.`orders` (`customer_id`, `order_date`, `status`, `total`, `notes`) VALUES
  (1, '2026-08-01', 'completed', 9798000.00, 'Pengiriman cepat'),
  (2, '2026-08-03', 'shipped', 1428000.00, NULL),
  (1, '2026-08-10', 'paid', 85000.00, 'Repeat order kopi'),
  (3, '2026-08-15', 'pending', 4298000.00, 'Tunggu stok laptop'),
  (5, '2026-08-20', 'cancelled', 299000.00, 'Batal, barang tidak sesuai'),
  (4, '2026-08-22', 'completed', 384000.00, NULL);

INSERT INTO `shop`.`order_items` (`order_id`, `product_id`, `qty`, `unit_price`) VALUES
  (1, 1, 1, 8499000.00),
  (1, 5, 10, 85000.00),
  (1, 7, 4, 199000.00),
  (2, 2, 1, 1299000.00),
  (2, 3, 1, 129000.00),
  (3, 5, 1, 85000.00),
  (4, 1, 1, 8499000.00),
  (4, 2, 2, 1299000.00),
  (4, 6, 5, 125000.00),
  (5, 4, 1, 299000.00),
  (6, 3, 2, 129000.00),
  (6, 6, 1, 125000.00);

-- Definer clause (edge case for the Definer Stripper)
DELIMITER $$
CREATE DEFINER=`root`@`localhost` PROCEDURE `shop`.`get_customer_orders`(IN p_customer_id INT)
BEGIN
  SELECT o.id, o.order_date, o.status, o.total
  FROM orders o
  WHERE o.customer_id = p_customer_id
  ORDER BY o.order_date DESC;
END$$

CREATE DEFINER=`root`@`localhost` TRIGGER `shop`.`trg_after_order_insert`
AFTER INSERT ON `orders`
FOR EACH ROW
BEGIN
  UPDATE customers SET created_at = created_at WHERE id = NEW.customer_id;
END$$
DELIMITER ;

-- ------------------------------------------------------------
-- Database: inventory
-- ------------------------------------------------------------
CREATE DATABASE IF NOT EXISTS `inventory`
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE `inventory`;

CREATE TABLE `inventory`.`warehouses` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL,
  `location` VARCHAR(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `inventory`.`stock_movements` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `product_sku` VARCHAR(30) NOT NULL,
  `warehouse_id` INT UNSIGNED NOT NULL,
  `qty_change` INT NOT NULL,
  `reason` VARCHAR(100) NOT NULL,
  `moved_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sku` (`product_sku`),
  KEY `idx_warehouse` (`warehouse_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `inventory`.`warehouses` (`name`, `location`) VALUES
  ('Gudang Pusat', 'Jakarta Selatan'),
  ('Gudang Cabang', 'Bandung');

INSERT INTO `inventory`.`stock_movements` (`product_sku`, `warehouse_id`, `qty_change`, `reason`) VALUES
  ('ELEC-001', 1, -1, 'Penjualan order #1'),
  ('ELEC-001', 1, +30, 'Restock dari supplier'),
  ('CLOTH-001', 2, -5, 'Transfer ke retail'),
  ('FOOD-001', 1, -10, 'Penjualan order #1'),
  ('ELEC-002', 2, +100, 'Restock dari supplier');

-- Large table edge case: a table with NULL-heavy columns
CREATE TABLE `inventory`.`audit_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `table_name` VARCHAR(64) NOT NULL,
  `record_id` BIGINT UNSIGNED DEFAULT NULL,
  `action` ENUM('INSERT','UPDATE','DELETE') NOT NULL,
  `old_data` JSON DEFAULT NULL,
  `new_data` JSON DEFAULT NULL,
  `actor` VARCHAR(100) DEFAULT NULL,
  `acted_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_table_action` (`table_name`, `action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `inventory`.`audit_log` (`table_name`, `record_id`, `action`, `old_data`, `new_data`, `actor`) VALUES
  ('products', 1, 'INSERT', NULL, '{"name":"Laptop ASUS VivoBook 14","price":8499000}', 'system'),
  ('orders', 1, 'INSERT', NULL, '{"status":"pending","total":9798000}', 'budi@example.com'),
  ('orders', 1, 'UPDATE', '{"status":"pending"}', '{"status":"completed"}', 'admin'),
  ('orders', 5, 'DELETE', '{"status":"cancelled","total":299000}', NULL, 'admin');
