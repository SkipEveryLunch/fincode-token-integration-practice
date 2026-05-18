import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import type { FincodeInstance } from '@fincode/js';

export default function CardRegisterPage() {
  const fincodeRef = useRef<FincodeInstance | null>(null);
  const [cardNo, setCardNo] = useState('');
  const [expireYear, setExpireYear] = useState('');
  const [expireMonth, setExpireMonth] = useState('');
  const [holderName, setHolderName] = useState('');
  const [securityCode, setSecurityCode] = useState('');
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    fincodeRef.current = window.Fincode(import.meta.env.VITE_FINCODE_PUBLIC_KEY);
  }, []);

  const handleConfirm = async () => {
    if (!fincodeRef.current) return;
    setError(null);

    const expire = expireYear.slice(-2) + expireMonth.padStart(2, '0');
    const normalizedCardNo = cardNo.replace(/\D/g, '');

    try {
      const token = await new Promise<string>((resolve, reject) => {
        fincodeRef.current!.tokens(
          { card_no: normalizedCardNo, expire, holder_name: holderName, security_code: securityCode, number: '1' },
          (status, response) => {
            if (status === 200) resolve(response.list[0].token);
            else reject(new Error(`トークン取得失敗: status ${status} body ${JSON.stringify(response)}`));
          },
          () => reject(new Error('トークン取得中に通信エラーが発生しました')),
        );
      });

      const res = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/cards`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token }),
      });
      if (!res.ok) {
        const body = await res.json();
        throw new Error(body.error ?? 'unknown error');
      }

      navigate('/card');
    } catch (e) {
      setError(`カード登録に失敗しました: ${e}`);
    }
  };

  return (
    <div>
      <h1>カード登録・更新</h1>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <div>
        <div>
          <label>カード番号</label>
          <input value={cardNo} onChange={e => setCardNo(e.target.value)} placeholder="1234567890123456" />
        </div>
        <div>
          <label>有効期限</label>
          <input value={expireMonth} onChange={e => setExpireMonth(e.target.value)} placeholder="MM" maxLength={2} />
          <span>/</span>
          <input value={expireYear} onChange={e => setExpireYear(e.target.value)} placeholder="YY" maxLength={2} />
        </div>
        <div>
          <label>カード名義人</label>
          <input value={holderName} onChange={e => setHolderName(e.target.value)} placeholder="TARO YAMADA" />
        </div>
        <div>
          <label>セキュリティコード</label>
          <input value={securityCode} onChange={e => setSecurityCode(e.target.value)} placeholder="123" maxLength={4} />
        </div>
      </div>
      <button onClick={handleConfirm}>確認へ</button>
    </div>
  );
}
