import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

export default function PurchasePage() {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  const handlePurchase = async () => {
    setError(null);
    try {
      const res = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/payments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ method: '1' }),
      });
      if (!res.ok) {
        const text = await res.text();
        let message = text;
        try { message = JSON.parse(text).error ?? text; } catch { /* plain text */ }
        throw new Error(message);
      }
      const { redirect_url } = await res.json();
      if (redirect_url) {
        window.location.href = redirect_url;
      } else {
        navigate('/purchase/success');
      }
    } catch (e) {
      setError(`購入に失敗しました: ${e}`);
    }
  };

  return (
    <div>
      <h1>購入</h1>
      <img src="/sushi.png" alt="寿司" style={{ width: '200px' }} />
      <p>寿司 500円</p>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <button onClick={handlePurchase}>購入する</button>
    </div>
  );
}
