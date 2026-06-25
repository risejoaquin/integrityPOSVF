let cart = [];
let products = window.INITIAL_DATA.products || [];

function init() {
  renderProducts();
  if (window.INITIAL_DATA.tablesActive) {
    fetchTables();
  }
}

function renderProducts() {
  const container = document.getElementById("menu-section");
  if (!container) return;
  container.innerHTML = "";
  products.forEach((p) => {
    const btn = document.createElement("button");
    btn.className = "product-btn";
    btn.onclick = () => addToCart(p);
    btn.innerHTML = `<span>${p.name}</span><strong>$${(p.price_cents / 100).toFixed(2)}</strong>`;
    container.appendChild(btn);
  });
}

function fetchTables() {
  fetch("/api/v1/tables")
    .then((res) => res.json())
    .then((tables) => {
      const select = document.getElementById("table-select");
      if (!select || !tables) return;
      tables.forEach((t) => {
        const opt = document.createElement("option");
        opt.value = t.id;
        opt.textContent = `${t.name} (${t.status})`;
        select.appendChild(opt);
      });
    })
    .catch((err) => console.error("No tables available", err));
}

function addToCart(p) {
  const existing = cart.find((i) => i.product_id === p.id);
  if (existing) {
    existing.quantity++;
  } else {
    cart.push({
      product_id: p.id,
      name: p.name,
      price: p.price_cents,
      quantity: 1,
    });
  }
  renderCart();
}

function renderCart() {
  const list = document.getElementById("cart-items");
  if (!list) return;
  list.innerHTML = "";
  let total = 0;
  cart.forEach((item, index) => {
    const li = document.createElement("li");
    li.innerHTML = `<span>${item.name} <strong style="color:#7f8c8d">x${item.quantity}</strong></span> <span>$${((item.price * item.quantity) / 100).toFixed(2)}</span>`;
    li.onclick = () => {
      cart.splice(index, 1);
      renderCart();
    };
    li.title = "Click para remover";
    list.appendChild(li);
    total += item.price * item.quantity;
  });
  const ct = document.getElementById("cart-total");
  if (ct) ct.textContent = (total / 100).toFixed(2);
}

function checkout() {
  if (cart.length === 0) return alert("El carrito está vacío");
  const orderReq = {
    items: cart.map((i) => ({
      product_id: i.product_id,
      quantity: i.quantity,
    })),
    source: "pos",
    customer_name: "Walk-in",
  };

  let orderId = "";

  fetch("/api/v1/orders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(orderReq),
  })
    .then((res) => {
      if (!res.ok) throw new Error("Error creando orden");
      return res.json();
    })
    .then((order) => {
      orderId = order.id;
      // Confirm order automatically
      return fetch(`/api/v1/orders/${order.id}/status`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: "confirmed" }),
      });
    })
    .then((res) => {
      if (!res.ok) throw new Error("Error confirmando orden");

      // Assign to table if selected
      const tableId = document.getElementById("table-select")?.value;
      if (tableId && orderId) {
        return fetch(`/api/v1/tables/${tableId}/assign`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ order_id: orderId }),
        });
      }
    })
    .then(() => {
      alert("¡Orden confirmada e impresa con éxito!");
      cart = [];
      renderCart();
      if (window.INITIAL_DATA.tablesActive) fetchTables(); // Refresh tables status
    })
    .catch((err) => alert("Error: " + err.message));
}

document.addEventListener("DOMContentLoaded", init);
