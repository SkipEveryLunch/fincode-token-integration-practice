import { useEffect, useState } from 'react';

type Card = {
  masked_card_number: string;
  expire: string;
  brand: string;
};

export default function CardConfirmPage() {
  const [card, setCard] = useState<Card | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch(`${import.meta.env.VITE_API_BASE_URL}/api/cards/active`)
      .then((res) => {
        if (res.status === 404) return null;
        if (!res.ok) throw new Error(`status ${res.status}`);
        return res.json();
      })
      .then(setCard)
      .catch((e) => setError(`カード情報の取得に失敗しました: ${e}`));
  }, []);

  if (error) return <p style={{ color: 'red' }}>{error}</p>;

  return (
    <div>
      <h1>カード確認</h1>
      {card ? (
        <dl>
          <dt>カード番号</dt>
          <dd>{card.masked_card_number}</dd>
          <dt>有効期限</dt>
          <dd>{card.expire}</dd>
          <dt>ブランド</dt>
          <dd>{card.brand}</dd>
        </dl>
      ) : (
        <p>登録済みのカードはありません。</p>
      )}
    </div>
  );
}
