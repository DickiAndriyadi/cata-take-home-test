CREATE TABLE orders (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  customer_id VARCHAR(36) NOT NULL,
  amount DECIMAL(10,2) NOT NULL,
  status ENUM('PENDING', 'PAID', 'CANCELLED') NOT NULL,
  created_at DATETIME NOT NULL
);


SELECT 
  status, 
  COUNT(*) AS order_count,
  SUM(amount) AS total_amount
FROM orders
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY status;


SELECT 
  customer_id, 
  SUM(amount) AS total_spent
FROM orders
WHERE status = 'PAID'
GROUP BY customer_id
ORDER BY total_spent DESC
LIMIT 5;