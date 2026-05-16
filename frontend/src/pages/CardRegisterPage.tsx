import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getCardToken } from '@fincode/js';
import type { FincodeInstance, FincodeUI } from '@fincode/js';

export default function CardRegisterPage() {
  const fincodeRef = useRef<FincodeInstance | null>(null);
  const uiRef = useRef<FincodeUI | null>(null);
  const initStarted = useRef(false);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (initStarted.current) return;
    initStarted.current = true;

    try {
      const fc = window.Fincode(import.meta.env.VITE_FINCODE_PUBLIC_KEY);
      const ui = fc.ui({ layout: 'horizontal' });
      ui.create('token', { layout: 'horizontal' });
      ui.mount('fincode-ui', '400');
      fincodeRef.current = fc;
      uiRef.current = ui;
      setReady(true);
    } catch (e) {
      setError(`fincode の初期化に失敗しました: ${e}`);
    }
  }, []);

  const handleConfirm = async () => {
    if (!fincodeRef.current || !uiRef.current) return;
    setError(null);
    try {
      const tokenResult = await getCardToken({
        fincode: fincodeRef.current,
        ui: uiRef.current,
        number: '1',
      });
      const formData = await uiRef.current.getFormData();
      const token = tokenResult.list[0].token;

      const res = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/cards`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token }),
      });
      if (!res.ok) {
        const body = await res.json();
        throw new Error(body.error ?? 'unknown error');
      }
      const card = await res.json();

      navigate('/card', {
        state: {
          maskedCardNumber: card.masked_card_number,
          expire: card.expire,
          brand: card.brand,
          holderName: formData.holderName,
        },
      });
    } catch (e) {
      setError(`カード登録に失敗しました: ${e}`);
    }
  };

  return (
    <div>
      <h1>カード登録・更新</h1>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <div id="fincode-ui">
        <div id="fincode-ui-form" />
      </div>
      {ready && <button onClick={handleConfirm}>確認へ</button>}
    </div>
  );
}
