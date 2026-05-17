import { useEffect, useState } from 'react';

type Payment = {
  amount: number;
  status: string;
  created_at: string;
};

export default function PaymentListPage() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch(`${import.meta.env.VITE_API_BASE_URL}/api/payments`)
      .then((res) => {
        if (!res.ok) throw new Error(`status ${res.status}`);
        return res.json();
      })
      .then(setPayments)
      .catch((e) => setError(`取得に失敗しました: ${e}`));
  }, []);

  if (error) return <p style={{ color: 'red' }}>{error}</p>;

  return (
    <div>
      <h1>購入履歴</h1>
      {payments.length === 0 ? (
        <p>購入履歴はありません。</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>日時</th>
              <th>金額</th>
              <th>ステータス</th>
            </tr>
          </thead>
          <tbody>
            {payments.map((p, i) => (
              <tr key={i}>
                <td>{new Date(p.created_at).toLocaleString('ja-JP')}</td>
                <td>{p.amount.toLocaleString()}円</td>
                <td>{p.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
