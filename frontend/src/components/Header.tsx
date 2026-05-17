import { NavLink } from 'react-router-dom';

export default function Header() {
  return (
    <header>
      <nav style={{ display: 'flex', gap: '1rem', padding: '0.5rem' }}>
        <NavLink to="/card">カード確認</NavLink>
        <NavLink to="/card/register">カード登録・更新</NavLink>
        <NavLink to="/purchase">購入</NavLink>
        <NavLink to="/payments">購入履歴</NavLink>
      </nav>
    </header>
  );
}
