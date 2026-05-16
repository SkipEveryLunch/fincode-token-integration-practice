import { useEffect, useRef, useState } from 'react';
import { getCardToken } from '@fincode/js';
import type { FincodeInstance, FincodeUI } from '@fincode/js';

export default function CardRegisterPage() {
  const fincodeRef = useRef<FincodeInstance | null>(null);
  const uiRef = useRef<FincodeUI | null>(null);
  const initStarted = useRef(false);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
      console.log('token:', tokenResult.list[0].token);
      console.log('card_no:', tokenResult.card_no);
      console.log('expire:', formData.expire);
      console.log('holderName:', formData.holderName);
    } catch {
      setError('カード情報の取得に失敗しました。入力内容を確認してください。');
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
