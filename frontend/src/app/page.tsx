"use client";

import { useState, useEffect, useCallback } from "react";

const GATEWAY_URL = process.env.NEXT_PUBLIC_GATEWAY_URL || "";

// --- Type definitions ---

interface Product {
  id: string;
  name: string;
  price: number;
  stock: number;
  createdAt: string;
  updatedAt: string;
}

interface CreateProductForm {
  name: string;
  price: string;
  stock: string;
}

interface CreateOrderForm {
  productId: string;
  quantity: string;
}

interface ApiErrorResponse {
  error: string;
}

interface Message {
  type: "success" | "error";
  text: string;
}

type Tab = "products" | "add-product" | "create-order";

// --- Main page component ---

export default function Home() {
  const [activeTab, setActiveTab] = useState<Tab>("products");
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<Message | null>(null);

  const [productForm, setProductForm] = useState<CreateProductForm>({
    name: "",
    price: "",
    stock: "",
  });

  const [orderForm, setOrderForm] = useState<CreateOrderForm>({
    productId: "",
    quantity: "",
  });

  // --- API calls ---

  const fetchProducts = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch(`${GATEWAY_URL}/products`);
      if (!res.ok) {
        const errorData: ApiErrorResponse = await res.json();
        throw new Error(errorData.error || `error ${res.status}`);
      }
      const data: Product[] = await res.json();
      setProducts(data);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "failed to fetch products";
      setMessage({ type: "error", text: errorMessage });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  const handleAddProduct = async () => {
    setMessage(null);

    const name = productForm.name.trim();
    const price = parseFloat(productForm.price);
    const stock = parseInt(productForm.stock, 10);

    if (!name) {
      setMessage({ type: "error", text: "nama produk tidak boleh kosong" });
      return;
    }
    if (isNaN(price) || price <= 0) {
      setMessage({ type: "error", text: "harga harus lebih dari 0" });
      return;
    }
    if (isNaN(stock) || stock < 0) {
      setMessage({ type: "error", text: "stok tidak boleh negatif" });
      return;
    }

    try {
      const res = await fetch(`${GATEWAY_URL}/products`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, price, stock }),
      });

      if (!res.ok) {
        const errorData: ApiErrorResponse = await res.json();
        throw new Error(errorData.error || `error ${res.status}`);
      }

      setMessage({ type: "success", text: "produk berhasil ditambahkan" });
      setProductForm({ name: "", price: "", stock: "" });
      await fetchProducts();
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "gagal menambahkan produk";
      setMessage({ type: "error", text: errorMessage });
    }
  };

  const handleCreateOrder = async () => {
    setMessage(null);

    const productId = orderForm.productId;
    const quantity = parseInt(orderForm.quantity, 10);

    if (!productId) {
      setMessage({ type: "error", text: "pilih produk terlebih dahulu" });
      return;
    }
    if (isNaN(quantity) || quantity <= 0) {
      setMessage({ type: "error", text: "quantity harus lebih dari 0" });
      return;
    }

    try {
      const res = await fetch(`${GATEWAY_URL}/orders`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ productId, quantity }),
      });

      if (!res.ok) {
        const errorData: ApiErrorResponse = await res.json();
        throw new Error(errorData.error || `error ${res.status}`);
      }

      setMessage({ type: "success", text: "order berhasil dibuat" });
      setOrderForm({ productId: "", quantity: "" });
      await fetchProducts();
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "gagal membuat order";
      setMessage({ type: "error", text: errorMessage });
    }
  };

  const switchTab = (tab: Tab) => {
    setActiveTab(tab);
    setMessage(null);
  };

  // --- Render ---

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-4xl mx-auto px-4 py-6">
          <h1 className="text-2xl font-bold text-gray-900">BIP ERP</h1>
          <p className="text-sm text-gray-500 mt-1">
            Inventory &amp; Order Management
          </p>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        {/* Tab Navigation */}
        <div className="border-b border-gray-200 mb-6">
          <nav className="flex gap-6">
            {[
              { key: "products" as Tab, label: "Daftar Produk" },
              { key: "add-product" as Tab, label: "Tambah Produk" },
              { key: "create-order" as Tab, label: "Buat Order" },
            ].map((tab) => (
              <button
                key={tab.key}
                id={`tab-${tab.key}`}
                onClick={() => switchTab(tab.key)}
                className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === tab.key
                    ? "border-blue-600 text-blue-600"
                    : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Message Alert */}
        {message && (
          <div
            id="message-alert"
            className={`mb-6 p-4 rounded-lg text-sm ${
              message.type === "success"
                ? "bg-green-50 text-green-800 border border-green-200"
                : "bg-red-50 text-red-800 border border-red-200"
            }`}
          >
            {message.text}
          </div>
        )}

        {/* Tab: Daftar Produk */}
        {activeTab === "products" && (
          <section>
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-semibold text-gray-800">
                Daftar Produk
              </h2>
              <button
                id="btn-refresh-products"
                onClick={fetchProducts}
                className="text-sm text-blue-600 hover:text-blue-800 transition-colors"
              >
                Refresh
              </button>
            </div>

            {loading ? (
              <p className="text-gray-500 text-center py-12">Memuat data...</p>
            ) : products.length === 0 ? (
              <p className="text-gray-500 text-center py-12">
                Belum ada produk. Tambahkan produk di tab &quot;Tambah
                Produk&quot;.
              </p>
            ) : (
              <div className="overflow-x-auto rounded-lg border border-gray-200">
                <table className="w-full bg-white">
                  <thead>
                    <tr className="bg-gray-50">
                      <th className="text-left px-4 py-3 text-xs font-semibold text-gray-600 uppercase tracking-wider">
                        Nama
                      </th>
                      <th className="text-right px-4 py-3 text-xs font-semibold text-gray-600 uppercase tracking-wider">
                        Harga
                      </th>
                      <th className="text-right px-4 py-3 text-xs font-semibold text-gray-600 uppercase tracking-wider">
                        Stok
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {products.map((product) => (
                      <tr
                        key={product.id}
                        className="hover:bg-gray-50 transition-colors"
                      >
                        <td className="px-4 py-3 text-sm text-gray-900">
                          {product.name}
                        </td>
                        <td className="px-4 py-3 text-sm text-gray-900 text-right">
                          Rp {product.price.toLocaleString("id-ID")}
                        </td>
                        <td className="px-4 py-3 text-sm text-right">
                          <span
                            className={`inline-block min-w-[2rem] text-center px-2 py-0.5 rounded-full text-xs font-medium ${
                              product.stock === 0
                                ? "bg-red-100 text-red-700"
                                : product.stock <= 5
                                ? "bg-yellow-100 text-yellow-700"
                                : "bg-green-100 text-green-700"
                            }`}
                          >
                            {product.stock}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        )}

        {/* Tab: Tambah Produk */}
        {activeTab === "add-product" && (
          <section className="max-w-md">
            <h2 className="text-lg font-semibold text-gray-800 mb-4">
              Tambah Produk
            </h2>
            <div className="space-y-4 bg-white p-6 rounded-lg border border-gray-200">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Nama Produk
                </label>
                <input
                  id="input-product-name"
                  type="text"
                  value={productForm.name}
                  onChange={(e) =>
                    setProductForm({ ...productForm, name: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="contoh: Paracetamol 500mg"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Harga (Rp)
                </label>
                <input
                  id="input-product-price"
                  type="number"
                  value={productForm.price}
                  onChange={(e) =>
                    setProductForm({ ...productForm, price: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="contoh: 5000"
                  min="0"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Stok
                </label>
                <input
                  id="input-product-stock"
                  type="number"
                  value={productForm.stock}
                  onChange={(e) =>
                    setProductForm({ ...productForm, stock: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="contoh: 100"
                  min="0"
                />
              </div>
              <button
                id="btn-add-product"
                onClick={handleAddProduct}
                className="w-full bg-blue-600 text-white py-2.5 px-4 rounded-lg text-sm font-medium hover:bg-blue-700 active:bg-blue-800 transition-colors"
              >
                Tambah Produk
              </button>
            </div>
          </section>
        )}

        {/* Tab: Buat Order */}
        {activeTab === "create-order" && (
          <section className="max-w-md">
            <h2 className="text-lg font-semibold text-gray-800 mb-4">
              Buat Order
            </h2>
            <div className="space-y-4 bg-white p-6 rounded-lg border border-gray-200">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Pilih Produk
                </label>
                <select
                  id="select-product"
                  value={orderForm.productId}
                  onChange={(e) =>
                    setOrderForm({ ...orderForm, productId: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
                >
                  <option value="">— Pilih Produk —</option>
                  {products.map((product) => (
                    <option key={product.id} value={product.id}>
                      {product.name} (Stok: {product.stock}) — Rp{" "}
                      {product.price.toLocaleString("id-ID")}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Quantity
                </label>
                <input
                  id="input-order-quantity"
                  type="number"
                  value={orderForm.quantity}
                  onChange={(e) =>
                    setOrderForm({ ...orderForm, quantity: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="contoh: 5"
                  min="1"
                />
              </div>
              <button
                id="btn-create-order"
                onClick={handleCreateOrder}
                className="w-full bg-green-600 text-white py-2.5 px-4 rounded-lg text-sm font-medium hover:bg-green-700 active:bg-green-800 transition-colors"
              >
                Buat Order
              </button>
            </div>
          </section>
        )}
      </main>
    </div>
  );
}
