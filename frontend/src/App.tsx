import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Header from './components/Header';
import CardConfirmPage from './pages/CardConfirmPage';
import CardRegisterPage from './pages/CardRegisterPage';
import PurchasePage from './pages/PurchasePage';
import PaymentListPage from './pages/PaymentListPage';

export default function App() {
  return (
    <BrowserRouter>
      <Header />
      <main>
        <Routes>
          <Route path="/card" element={<CardConfirmPage />} />
          <Route path="/card/register" element={<CardRegisterPage />} />
          <Route path="/purchase" element={<PurchasePage />} />
          <Route path="/payments" element={<PaymentListPage />} />
          <Route path="*" element={<Navigate to="/card" replace />} />
        </Routes>
      </main>
    </BrowserRouter>
  );
}
